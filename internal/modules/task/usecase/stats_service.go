package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/pkg/logger"
)

// StatsRepository は統計情報取得のためのリポジトリインターフェース
type StatsRepository interface {
	// GetTasksByDateRange は指定された日付範囲のタスクを取得する
	GetTasksByDateRange(ctx context.Context, userID string, start, end time.Time) ([]*domain.Task, error)

	// GetTasksByDueDate は指定された期限日のタスクを取得する
	GetTasksByDueDate(ctx context.Context, userID string, dueDate time.Time) ([]*domain.Task, error)

	// GetRecentCompletedTasks は最近完了したタスクを取得する
	GetRecentCompletedTasks(ctx context.Context, userID string, limit int) ([]*domain.Task, error)

	// GetOverdueTasksCount は期限切れタスク数を取得する
	GetOverdueTasksCount(ctx context.Context, userID string) (int, error)
}

// TaskStatsService はタスク統計情報を提供するサービス
type TaskStatsService struct {
	taskRepo  TaskRepository
	statsRepo StatsRepository
	logger    logger.Logger
}

// NewTaskStatsService は新しいTaskStatsServiceを作成する
func NewTaskStatsService(
	taskRepo TaskRepository,
	statsRepo StatsRepository,
	logger logger.Logger,
) *TaskStatsService {
	return &TaskStatsService{
		taskRepo:  taskRepo,
		statsRepo: statsRepo,
		logger:    logger,
	}
}

// GetDashboardStats はダッシュボード用の統計情報を取得する
func (s *TaskStatsService) GetDashboardStats(ctx context.Context, userID string) (*domain.DashboardStats, error) {
	now := time.Now()

	// 今日の統計
	todayStats, err := s.GetDailyStats(ctx, userID, now)
	if err != nil {
		s.logger.Error("Failed to get today stats", logger.Any("userID", userID), logger.Error(err))
		return nil, fmt.Errorf("failed to get today stats: %w", err)
	}

	// 今週の統計
	weeklyOverview, err := s.GetWeeklyStats(ctx, userID, now)
	if err != nil {
		s.logger.Error("Failed to get weekly stats", logger.Any("userID", userID), logger.Error(err))
		return nil, fmt.Errorf("failed to get weekly stats: %w", err)
	}

	// 来週のプレビュー
	nextWeek := now.AddDate(0, 0, 7)
	upcomingWeekTasks, err := s.GetWeeklyPreview(ctx, userID, nextWeek)
	if err != nil {
		s.logger.Error("Failed to get upcoming week tasks", logger.Any("userID", userID), logger.Error(err))
		return nil, fmt.Errorf("failed to get upcoming week tasks: %w", err)
	}

	// カテゴリ別統計
	categoryBreakdown, err := s.GetCategoryBreakdown(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get category breakdown", logger.Any("userID", userID), logger.Error(err))
		categoryBreakdown = make(map[domain.Category]int) // エラー時は空のマップ
	}

	// 優先度別統計
	priorityBreakdown, err := s.GetPriorityBreakdown(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get priority breakdown", logger.Any("userID", userID), logger.Error(err))
		priorityBreakdown = make(map[domain.Priority]int) // エラー時は空のマップ
	}

	// 最近の完了タスク
	recentCompletions, err := s.statsRepo.GetRecentCompletedTasks(ctx, userID, 5)
	if err != nil {
		s.logger.Error("Failed to get recent completions", logger.Any("userID", userID), logger.Error(err))
		recentCompletions = []*domain.Task{} // エラー時は空のスライス
	}

	// 期限切れタスク数
	overdueCount, err := s.statsRepo.GetOverdueTasksCount(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get overdue count", logger.Any("userID", userID), logger.Error(err))
		overdueCount = 0 // エラー時は0
	}

	return &domain.DashboardStats{
		TodayStats:        todayStats,
		WeeklyOverview:    weeklyOverview,
		UpcomingWeekTasks: upcomingWeekTasks,
		CategoryBreakdown: categoryBreakdown,
		PriorityBreakdown: priorityBreakdown,
		RecentCompletions: recentCompletions,
		OverdueTasksCount: overdueCount,
	}, nil
}

// GetDailyStats は指定日の統計情報を取得する
func (s *TaskStatsService) GetDailyStats(ctx context.Context, userID string, date time.Time) (*domain.DailyStats, error) {
	dayStart, dayEnd := domain.GetDayStartEnd(date)

	// その日が期限のタスクを取得
	tasks, err := s.statsRepo.GetTasksByDueDate(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by due date: %w", err)
	}

	// その日に作成されたタスクも含める
	createdTasks, err := s.statsRepo.GetTasksByDateRange(ctx, userID, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by date range: %w", err)
	}

	// 重複を除去してマージ
	taskMap := make(map[string]*domain.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	for _, task := range createdTasks {
		taskMap[task.ID] = task
	}

	allTasks := make([]*domain.Task, 0, len(taskMap))
	for _, task := range taskMap {
		allTasks = append(allTasks, task)
	}

	return domain.NewDailyStats(date, allTasks), nil
}

// GetWeeklyStats は指定週の統計情報を取得する
func (s *TaskStatsService) GetWeeklyStats(ctx context.Context, userID string, date time.Time) (*domain.WeeklyStats, error) {
	weekStart, weekEnd := domain.GetWeekStartEnd(date)

	dailyStats := make(map[string]*domain.DailyStats)

	// 各曜日の統計を取得
	for d := weekStart; !d.After(weekEnd); d = d.AddDate(0, 0, 1) {
		dayStats, err := s.GetDailyStats(ctx, userID, d)
		if err != nil {
			s.logger.Error("Failed to get daily stats",
				logger.Any("userID", userID),
				logger.Any("date", d),
				logger.Error(err))
			// エラーでも継続（空の統計で代替）
			dayStats = domain.NewDailyStats(d, []*domain.Task{})
		}

		weekdayName := domain.GetWeekdayName(d.Weekday())
		dailyStats[weekdayName] = dayStats
	}

	return domain.NewWeeklyStats(weekStart, weekEnd, dailyStats), nil
}

// GetWeeklyPreview は指定週のプレビュー情報を取得する
func (s *TaskStatsService) GetWeeklyPreview(ctx context.Context, userID string, date time.Time) (*domain.WeeklyPreview, error) {
	weekStart, weekEnd := domain.GetWeekStartEnd(date)

	preview := &domain.WeeklyPreview{
		WeekStart:    weekStart,
		WeekEnd:      weekEnd,
		DailyPreview: make(map[string]*domain.DailyPreview),
	}

	// 各曜日のプレビューを取得
	for d := weekStart; !d.After(weekEnd); d = d.AddDate(0, 0, 1) {
		tasks, err := s.statsRepo.GetTasksByDueDate(ctx, userID, d)
		if err != nil {
			s.logger.Error("Failed to get tasks for preview",
				logger.Any("userID", userID),
				logger.Any("date", d),
				logger.Error(err))
			tasks = []*domain.Task{} // エラー時は空のスライス
		}

		hasOverdue := false
		for _, task := range tasks {
			if task.IsOverdue() {
				hasOverdue = true
				break
			}
		}

		weekdayName := domain.GetWeekdayName(d.Weekday())
		preview.DailyPreview[weekdayName] = &domain.DailyPreview{
			Date:       d,
			TaskCount:  len(tasks),
			HasOverdue: hasOverdue,
		}

		preview.TotalTasks += len(tasks)
	}

	return preview, nil
}

// GetCategoryBreakdown はカテゴリ別のタスク分布を取得する
func (s *TaskStatsService) GetCategoryBreakdown(ctx context.Context, userID string) (map[domain.Category]int, error) {
	// 過去30日間のタスクを対象とする
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	tasks, err := s.statsRepo.GetTasksByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for category breakdown: %w", err)
	}

	breakdown := make(map[domain.Category]int)
	for _, task := range tasks {
		breakdown[task.Category]++
	}

	return breakdown, nil
}

// GetPriorityBreakdown は優先度別のタスク分布を取得する
func (s *TaskStatsService) GetPriorityBreakdown(ctx context.Context, userID string) (map[domain.Priority]int, error) {
	// アクティブなタスク（完了していないタスク）を対象とする
	filter := domain.ListFilter{
		AssigneeID: &userID,
	}

	pagination := domain.Pagination{
		Page:     1,
		PageSize: 1000, // 十分に大きな値
	}

	sortOptions := domain.SortOptions{
		Field:     "created_at",
		Direction: "DESC",
	}

	tasks, _, err := s.taskRepo.ListTasks(ctx, filter, pagination, sortOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for priority breakdown: %w", err)
	}

	breakdown := make(map[domain.Priority]int)
	for _, task := range tasks {
		if task.Status != domain.TaskStatusDone {
			breakdown[task.Priority]++
		}
	}

	return breakdown, nil
}

// GetProgressSummary は進捗サマリーを取得する
func (s *TaskStatsService) GetProgressSummary(ctx context.Context, userID string, days int) ([]*domain.DailyStats, error) {
	summary := make([]*domain.DailyStats, 0, days)

	today := time.Now()
	for i := days - 1; i >= 0; i-- {
		date := today.AddDate(0, 0, -i)
		dailyStats, err := s.GetDailyStats(ctx, userID, date)
		if err != nil {
			s.logger.Error("Failed to get daily stats for summary",
				logger.Any("userID", userID),
				logger.Any("date", date),
				logger.Error(err))
			// エラーでも継続（空の統計で代替）
			dailyStats = domain.NewDailyStats(date, []*domain.Task{})
		}
		summary = append(summary, dailyStats)
	}

	return summary, nil
}

// GetMonthlyStats は月次統計を取得する
func (s *TaskStatsService) GetMonthlyStats(ctx context.Context, userID string, year int, month time.Month) (*domain.WeeklyStats, error) {
	// 月の開始日と終了日を取得
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, -1)
	monthEnd = time.Date(monthEnd.Year(), monthEnd.Month(), monthEnd.Day(), 23, 59, 59, 999999999, time.UTC)

	// 月間のタスクを取得
	tasks, err := s.statsRepo.GetTasksByDateRange(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly tasks: %w", err)
	}

	// 月間の統計を週単位で分割
	dailyStats := make(map[string]*domain.DailyStats)

	// 各日の統計を計算
	for d := monthStart; !d.After(monthEnd); d = d.AddDate(0, 0, 1) {
		dayTasks := make([]*domain.Task, 0)
		for _, task := range tasks {
			// その日が期限日または作成日のタスクを抽出
			if (task.DueDate != nil && task.DueDate.Format("2006-01-02") == d.Format("2006-01-02")) ||
				task.CreatedAt.Format("2006-01-02") == d.Format("2006-01-02") {
				dayTasks = append(dayTasks, task)
			}
		}

		dayKey := d.Format("2006-01-02")
		dailyStats[dayKey] = domain.NewDailyStats(d, dayTasks)
	}

	return domain.NewWeeklyStats(monthStart, monthEnd, dailyStats), nil
}
