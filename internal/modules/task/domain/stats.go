package domain

import (
	"time"
)

// DailyStats は日次のタスク統計を表す
type DailyStats struct {
	Date            time.Time `json:"date"`
	TotalTasks      int       `json:"total_tasks"`
	CompletedTasks  int       `json:"completed_tasks"`
	InProgressTasks int       `json:"in_progress_tasks"`
	TodoTasks       int       `json:"todo_tasks"`
	OverdueTasks    int       `json:"overdue_tasks"`
	CompletionRate  float64   `json:"completion_rate"` // 0-100の範囲
}

// WeeklyStats は週次のタスク統計を表す
type WeeklyStats struct {
	WeekStart      time.Time              `json:"week_start"`
	WeekEnd        time.Time              `json:"week_end"`
	TotalTasks     int                    `json:"total_tasks"`
	CompletedTasks int                    `json:"completed_tasks"`
	CompletionRate float64                `json:"completion_rate"`
	DailyStats     map[string]*DailyStats `json:"daily_stats"` // key: "Monday", "Tuesday", etc.
}

// ProgressColor は進捗率に応じた色を表す
type ProgressColor string

const (
	ColorDarkGreen ProgressColor = "#22c55e" // 100%完了：濃い緑
	ColorGreen     ProgressColor = "#84cc16" // 80-99%：緑
	ColorYellow    ProgressColor = "#eab308" // 60-79%：黄色
	ColorOrange    ProgressColor = "#f97316" // 40-59%：オレンジ
	ColorLightRed  ProgressColor = "#ef4444" // 20-39%：薄い赤
	ColorRed       ProgressColor = "#dc2626" // 1-19%：赤
	ColorGray      ProgressColor = "#9ca3af" // 0%：灰色
)

// DashboardStats はダッシュボード用の統計情報を表す
type DashboardStats struct {
	TodayStats        *DailyStats      `json:"today_stats"`
	WeeklyOverview    *WeeklyStats     `json:"weekly_overview"`
	UpcomingWeekTasks *WeeklyPreview   `json:"upcoming_week_tasks"`
	CategoryBreakdown map[Category]int `json:"category_breakdown"`
	PriorityBreakdown map[Priority]int `json:"priority_breakdown"`
	RecentCompletions []*Task          `json:"recent_completions"`
	OverdueTasksCount int              `json:"overdue_tasks_count"`
}

// WeeklyPreview は今後1週間のタスクプレビューを表す
type WeeklyPreview struct {
	WeekStart    time.Time                `json:"week_start"`
	WeekEnd      time.Time                `json:"week_end"`
	TotalTasks   int                      `json:"total_tasks"`
	DailyPreview map[string]*DailyPreview `json:"daily_preview"` // key: "Monday", "Tuesday", etc.
}

// DailyPreview は日次のタスクプレビューを表す
type DailyPreview struct {
	Date       time.Time `json:"date"`
	TaskCount  int       `json:"task_count"`
	HasOverdue bool      `json:"has_overdue"`
}

// ProgressLevel は進捗レベルを表す
type ProgressLevel struct {
	Percentage int           `json:"percentage"`
	Color      ProgressColor `json:"color"`
	Label      string        `json:"label"`
}

// GetProgressColor は進捗率に応じた色を取得する
func GetProgressColor(completionRate float64) ProgressColor {
	switch {
	case completionRate >= 100:
		return ColorDarkGreen
	case completionRate >= 80:
		return ColorGreen
	case completionRate >= 60:
		return ColorYellow
	case completionRate >= 40:
		return ColorOrange
	case completionRate >= 20:
		return ColorLightRed
	case completionRate >= 1:
		return ColorRed
	default:
		return ColorGray
	}
}

// GetProgressLevel は進捗率に応じたレベル情報を取得する
func GetProgressLevel(completionRate float64) ProgressLevel {
	color := GetProgressColor(completionRate)
	var label string

	switch {
	case completionRate >= 100:
		label = "完了"
	case completionRate >= 80:
		label = "優秀"
	case completionRate >= 60:
		label = "良好"
	case completionRate >= 40:
		label = "普通"
	case completionRate >= 20:
		label = "要改善"
	case completionRate >= 1:
		label = "低調"
	default:
		label = "未着手"
	}

	return ProgressLevel{
		Percentage: int(completionRate),
		Color:      color,
		Label:      label,
	}
}

// CalculateCompletionRate は完了率を計算する
func CalculateCompletionRate(completed, total int) float64 {
	if total == 0 {
		return 0.0
	}
	return (float64(completed) / float64(total)) * 100
}

// GetWeekdayName は曜日名を取得する
func GetWeekdayName(weekday time.Weekday) string {
	weekdays := map[time.Weekday]string{
		time.Monday:    "Monday",
		time.Tuesday:   "Tuesday",
		time.Wednesday: "Wednesday",
		time.Thursday:  "Thursday",
		time.Friday:    "Friday",
		time.Saturday:  "Saturday",
		time.Sunday:    "Sunday",
	}
	return weekdays[weekday]
}

// GetWeekdayNameJP は日本語の曜日名を取得する
func GetWeekdayNameJP(weekday time.Weekday) string {
	weekdays := map[time.Weekday]string{
		time.Monday:    "月",
		time.Tuesday:   "火",
		time.Wednesday: "水",
		time.Thursday:  "木",
		time.Friday:    "金",
		time.Saturday:  "土",
		time.Sunday:    "日",
	}
	return weekdays[weekday]
}

// GetWeekStartEnd は指定された日付の週の開始日と終了日を取得する（月曜開始）
func GetWeekStartEnd(date time.Time) (time.Time, time.Time) {
	// 月曜日を週の開始とする
	weekday := date.Weekday()
	daysFromMonday := int(weekday) - int(time.Monday)
	if daysFromMonday < 0 {
		daysFromMonday += 7
	}

	weekStart := date.AddDate(0, 0, -daysFromMonday)
	weekEnd := weekStart.AddDate(0, 0, 6)

	// 時刻を0時0分0秒に設定
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	weekEnd = time.Date(weekEnd.Year(), weekEnd.Month(), weekEnd.Day(), 23, 59, 59, 999999999, weekEnd.Location())

	return weekStart, weekEnd
}

// GetDayStartEnd は指定された日付の開始時刻と終了時刻を取得する
func GetDayStartEnd(date time.Time) (time.Time, time.Time) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())
	return dayStart, dayEnd
}

// NewDailyStats は新しいDailyStatsを作成する
func NewDailyStats(date time.Time, tasks []*Task) *DailyStats {
	stats := &DailyStats{
		Date:       date,
		TotalTasks: len(tasks),
	}

	for _, task := range tasks {
		switch task.Status {
		case TaskStatusDone:
			stats.CompletedTasks++
		case TaskStatusInProgress:
			stats.InProgressTasks++
		case TaskStatusTodo:
			stats.TodoTasks++
		}

		if task.CheckIsOverdue() {
			stats.OverdueTasks++
		}
	}

	stats.CompletionRate = CalculateCompletionRate(stats.CompletedTasks, stats.TotalTasks)
	return stats
}

// NewWeeklyStats は新しいWeeklyStatsを作成する
func NewWeeklyStats(weekStart, weekEnd time.Time, dailyStats map[string]*DailyStats) *WeeklyStats {
	stats := &WeeklyStats{
		WeekStart:  weekStart,
		WeekEnd:    weekEnd,
		DailyStats: dailyStats,
	}

	for _, daily := range dailyStats {
		stats.TotalTasks += daily.TotalTasks
		stats.CompletedTasks += daily.CompletedTasks
	}

	stats.CompletionRate = CalculateCompletionRate(stats.CompletedTasks, stats.TotalTasks)
	return stats
}
