package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase"
)

// TaskStatsController はタスク統計のHTTPリクエストを処理するコントローラー
type TaskStatsController struct {
	statsService *usecase.TaskStatsService
}

// NewTaskStatsController は新しいTaskStatsControllerを作成する
func NewTaskStatsController(statsService *usecase.TaskStatsService) *TaskStatsController {
	return &TaskStatsController{
		statsService: statsService,
	}
}

// DashboardStatsData はダッシュボード統計のデータ構造
type DashboardStatsData struct {
	TodayStats        *DailyStatsData    `json:"today_stats"`
	WeeklyOverview    *WeeklyStatsData   `json:"weekly_overview"`
	UpcomingWeekTasks *WeeklyPreviewData `json:"upcoming_week_tasks"`
	CategoryBreakdown map[string]int     `json:"category_breakdown"`
	PriorityBreakdown map[string]int     `json:"priority_breakdown"`
	RecentCompletions []TaskSummary      `json:"recent_completions"`
	OverdueTasksCount int                `json:"overdue_tasks_count"`
} // @name DashboardStatsData

// DailyStatsData は日次統計のデータ構造
type DailyStatsData struct {
	Date            string  `json:"date" example:"2024-01-01"`
	TotalTasks      int     `json:"total_tasks" example:"10"`
	CompletedTasks  int     `json:"completed_tasks" example:"7"`
	InProgressTasks int     `json:"in_progress_tasks" example:"2"`
	TodoTasks       int     `json:"todo_tasks" example:"1"`
	OverdueTasks    int     `json:"overdue_tasks" example:"0"`
	CompletionRate  float64 `json:"completion_rate" example:"70.0"`
} // @name DailyStatsData

// WeeklyStatsData は週次統計のデータ構造
type WeeklyStatsData struct {
	WeekStart      string                     `json:"week_start" example:"2024-01-01"`
	WeekEnd        string                     `json:"week_end" example:"2024-01-07"`
	TotalTasks     int                        `json:"total_tasks" example:"50"`
	CompletedTasks int                        `json:"completed_tasks" example:"35"`
	CompletionRate float64                    `json:"completion_rate" example:"70.0"`
	DailyStats     map[string]*DailyStatsData `json:"daily_stats"`
} // @name WeeklyStatsData

// WeeklyPreviewData は今後1週間のタスクプレビュー
type WeeklyPreviewData struct {
	WeekStart    string                       `json:"week_start" example:"2024-01-01"`
	WeekEnd      string                       `json:"week_end" example:"2024-01-07"`
	TotalTasks   int                          `json:"total_tasks" example:"15"`
	DailyPreview map[string]*DailyPreviewData `json:"daily_preview"`
} // @name WeeklyPreviewData

// DailyPreviewData は日次のタスクプレビュー
type DailyPreviewData struct {
	Date       string `json:"date" example:"2024-01-01"`
	TaskCount  int    `json:"task_count" example:"3"`
	HasOverdue bool   `json:"has_overdue" example:"false"`
} // @name DailyPreviewData

// TaskSummary はタスクの要約情報
type TaskSummary struct {
	ID          string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title       string `json:"title" example:"重要なタスク"`
	Status      string `json:"status" example:"DONE"`
	CompletedAt string `json:"completed_at" example:"2024-01-01T12:00:00Z"`
} // @name TaskSummary

// ProgressLevelData は進捗レベル情報
type ProgressLevelData struct {
	Percentage int    `json:"percentage" example:"85"`
	Color      string `json:"color" example:"#22c55e"`
	Label      string `json:"label" example:"良好"`
} // @name ProgressLevelData

// CategoryBreakdownData はカテゴリ別統計
type CategoryBreakdownData struct {
	Count       int    `json:"count" example:"15"`
	DisplayName string `json:"display_name" example:"仕事"`
	Color       string `json:"color" example:"#3b82f6"`
} // @name CategoryBreakdownData

// PriorityBreakdownData は優先度別統計
type PriorityBreakdownData struct {
	Count       int    `json:"count" example:"8"`
	DisplayName string `json:"display_name" example:"高"`
	Color       string `json:"color" example:"#dc2626"`
} // @name PriorityBreakdownData

// DashboardStatsResponse はダッシュボード統計のレスポンス
type DashboardStatsResponse struct {
	Success bool               `json:"success" example:"true"`
	Data    DashboardStatsData `json:"data"`
} // @name DashboardStatsResponse

// DailyStatsResponse は日次統計のレスポンス
type DailyStatsResponse struct {
	Success bool           `json:"success" example:"true"`
	Data    DailyStatsData `json:"data"`
} // @name DailyStatsResponse

// WeeklyStatsResponse は週次統計のレスポンス
type WeeklyStatsResponse struct {
	Success bool            `json:"success" example:"true"`
	Data    WeeklyStatsData `json:"data"`
} // @name WeeklyStatsResponse

// ProgressSummaryResponse は進捗サマリーのレスポンス
type ProgressSummaryResponse struct {
	Success bool             `json:"success" example:"true"`
	Data    []DailyStatsData `json:"data"`
} // @name ProgressSummaryResponse

// ProgressLevelResponse は進捗レベルのレスポンス
type ProgressLevelResponse struct {
	Success bool              `json:"success" example:"true"`
	Data    ProgressLevelData `json:"data"`
} // @name ProgressLevelResponse


// GetDashboardStats ダッシュボード統計取得
// @Summary      ダッシュボード統計取得
// @Description  ダッシュボード表示用の包括的な統計情報を取得します
// @Tags         stats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} DashboardStatsResponse "ダッシュボード統計取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/stats/dashboard [get]
func (c *TaskStatsController) GetDashboardStats(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	stats, err := c.statsService.GetDashboardStats(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Failed to get dashboard stats",
	})
		return
	}

	ctx.JSON(http.StatusOK, DashboardStatsResponse{
		Success: true,
		Data:    *convertDashboardStats(stats),
	})
}

// GetTodayStats 今日の統計取得
// @Summary      今日の統計取得
// @Description  本日のタスク統計情報を取得します
// @Tags         stats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} DailyStatsResponse "今日の統計取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/stats/today [get]
func (c *TaskStatsController) GetTodayStats(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	today := time.Now()
	stats, err := c.statsService.GetDailyStats(ctx, userID, today)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Failed to get today stats",
	})
		return
	}

	ctx.JSON(http.StatusOK, DailyStatsResponse{
		Success: true,
		Data:    *convertDailyStats(stats),
	})
}

// GetDailyStats 指定日の統計取得
// @Summary      指定日の統計取得
// @Description  指定された日付のタスク統計情報を取得します
// @Tags         stats
// @Accept       json
// @Produce      json
// @Param        date path string true "日付(YYYY-MM-DD形式)" example:"2024-01-01"
// @Security     BearerAuth
// @Success      200 {object} DailyStatsResponse "日次統計取得成功"
// @Failure      400 {object} ErrorResponse "日付形式が無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/stats/daily/{date} [get]
func (c *TaskStatsController) GetDailyStats(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	// 日付パラメータの取得
	dateStr := ctx.Param("date")
	if dateStr == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Date parameter is required",
	})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Invalid date format. Use YYYY-MM-DD",
	})
		return
	}

	stats, err := c.statsService.GetDailyStats(ctx, userID, date)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Failed to get daily stats",
	})
		return
	}

	ctx.JSON(http.StatusOK, DailyStatsResponse{
		Success: true,
		Data:    *convertDailyStats(stats),
	})
}

// GetWeeklyStats 週次統計取得
// @Summary      週次統計取得
// @Description  指定された週のタスク統計情報を取得します
// @Tags         stats
// @Accept       json
// @Produce      json
// @Param        date query string false "基準日付(YYYY-MM-DD形式)" example:"2024-01-01"
// @Security     BearerAuth
// @Success      200 {object} WeeklyStatsResponse "週次統計取得成功"
// @Failure      400 {object} ErrorResponse "日付形式が無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/stats/weekly [get]
func (c *TaskStatsController) GetWeeklyStats(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	// 日付パラメータの取得（その週に含まれる任意の日付）
	dateStr := ctx.Query("date")
	var date time.Time
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Invalid date format. Use YYYY-MM-DD",
	})
			return
		}
		date = parsedDate
	} else {
		date = time.Now() // デフォルトは今週
	}

	stats, err := c.statsService.GetWeeklyStats(ctx, userID, date)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Failed to get weekly stats",
	})
		return
	}

	ctx.JSON(http.StatusOK, WeeklyStatsResponse{
		Success: true,
		Data:    *convertWeeklyStats(stats),
	})
}

// GetProgressSummary 進捗サマリー取得
// @Summary      進捗サマリー取得
// @Description  指定された日数分の進捗サマリーを取得します
// @Tags         stats
// @Accept       json
// @Produce      json
// @Param        days query int false "対象日数" default(7) minimum(1) maximum(365)
// @Security     BearerAuth
// @Success      200 {object} ProgressSummaryResponse "進捗サマリー取得成功"
// @Failure      400 {object} ErrorResponse "パラメータが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/stats/progress-summary [get]
func (c *TaskStatsController) GetProgressSummary(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	// 日数パラメータの取得
	daysStr := ctx.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Invalid days parameter. Must be between 1 and 365",
	})
		return
	}

	summary, err := c.statsService.GetProgressSummary(ctx, userID, days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Failed to get progress summary",
	})
		return
	}

	ctx.JSON(http.StatusOK, ProgressSummaryResponse{
		Success: true,
		Data:    convertDailyStatsList(summary),
	})
}

// GetProgressLevel 進捗レベル取得
// @Summary      進捗レベル取得
// @Description  完了率に基づく進捗レベル情報を取得します
// @Tags         stats
// @Accept       json
// @Produce      json
// @Param        rate query number true "完了率(0-100)" example:"85.5"
// @Security     BearerAuth
// @Success      200 {object} ProgressLevelResponse "進捗レベル取得成功"
// @Failure      400 {object} ErrorResponse "完了率パラメータが無効"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/stats/progress-level [get]
func (c *TaskStatsController) GetProgressLevel(ctx *gin.Context) {
	// 完了率パラメータの取得
	rateStr := ctx.Query("rate")
	if rateStr == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Completion rate parameter is required",
	})
		return
	}

	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil || rate < 0 || rate > 100 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Invalid completion rate. Must be between 0 and 100",
	})
		return
	}

	level := domain.GetProgressLevel(rate)

	ctx.JSON(http.StatusOK, ProgressLevelResponse{
		Success: true,
		Data: ProgressLevelData{
			Percentage: level.Percentage,
			Color:      string(level.Color),
			Label:      level.Label,
		},
	})
}

// GetCategoryBreakdown カテゴリ別統計取得
// @Summary      カテゴリ別統計取得
// @Description  タスクのカテゴリ別内訳統計を取得します
// @Tags         stats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{success=bool,data=object{WORK=CategoryBreakdownData,PERSONAL=CategoryBreakdownData}} "カテゴリ別統計取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/stats/category-breakdown [get]
func (c *TaskStatsController) GetCategoryBreakdown(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	breakdown, err := c.statsService.GetCategoryBreakdown(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Failed to get category breakdown",
	})
		return
	}

	// カテゴリ名の日本語変換
	result := make(map[string]interface{})
	for category, count := range breakdown {
		result[string(category)] = gin.H{
			"count":        count,
			"display_name": category.GetDisplayName(),
			"color":        getCategoryColor(category),
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetPriorityBreakdown 優先度別統計取得
// @Summary      優先度別統計取得
// @Description  タスクの優先度別内訳統計を取得します
// @Tags         stats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{success=bool,data=object{HIGH=PriorityBreakdownData,MEDIUM=PriorityBreakdownData,LOW=PriorityBreakdownData}} "優先度別統計取得成功"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/stats/priority-breakdown [get]
func (c *TaskStatsController) GetPriorityBreakdown(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	breakdown, err := c.statsService.GetPriorityBreakdown(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Failed to get priority breakdown",
	})
		return
	}

	// 優先度名の日本語変換
	result := make(map[string]interface{})
	for priority, count := range breakdown {
		result[string(priority)] = gin.H{
			"count":        count,
			"display_name": priority.GetDisplayName(),
			"color":        getPriorityColor(priority),
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetMonthlyStats 月次統計取得
// @Summary      月次統計取得
// @Description  指定された月のタスク統計情報を取得します
// @Tags         stats
// @Accept       json
// @Produce      json
// @Param        year query int false "年" default(2024) minimum(2000) maximum(3000)
// @Param        month query int false "月" default(1) minimum(1) maximum(12)
// @Security     BearerAuth
// @Success      200 {object} WeeklyStatsResponse "月次統計取得成功"
// @Failure      400 {object} ErrorResponse "パラメータが無効"
// @Failure      401 {object} ErrorResponse "認証が必要"
// @Failure      500 {object} ErrorResponse "内部サーバーエラー"
// @Router       /tasks/stats/monthly [get]
func (c *TaskStatsController) GetMonthlyStats(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: err.Error(),
	})
		return
	}

	// 年月パラメータの取得
	yearStr := ctx.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	monthStr := ctx.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 3000 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Invalid year parameter",
	})
		return
	}

	monthInt, err := strconv.Atoi(monthStr)
	if err != nil || monthInt < 1 || monthInt > 12 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Invalid month parameter",
	})
		return
	}

	month := time.Month(monthInt)

	stats, err := c.statsService.GetMonthlyStats(ctx, userID, year, month)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "REQUEST_ERROR",
		Message: "Failed to get monthly stats",
	})
		return
	}

	ctx.JSON(http.StatusOK, WeeklyStatsResponse{
		Success: true,
		Data:    *convertWeeklyStats(stats),
	})
}

// ヘルパー関数群
func convertDashboardStats(stats *domain.DashboardStats) *DashboardStatsData {
	return &DashboardStatsData{
		TodayStats:        convertDailyStats(stats.TodayStats),
		WeeklyOverview:    convertWeeklyStats(stats.WeeklyOverview),
		UpcomingWeekTasks: convertWeeklyPreview(stats.UpcomingWeekTasks),
		OverdueTasksCount: stats.OverdueTasksCount,
	}
}

func convertDailyStats(stats *domain.DailyStats) *DailyStatsData {
	return &DailyStatsData{
		Date:            stats.Date.Format("2006-01-02"),
		TotalTasks:      stats.TotalTasks,
		CompletedTasks:  stats.CompletedTasks,
		InProgressTasks: stats.InProgressTasks,
		TodoTasks:       stats.TodoTasks,
		OverdueTasks:    stats.OverdueTasks,
		CompletionRate:  stats.CompletionRate,
	}
}

func convertWeeklyStats(stats *domain.WeeklyStats) *WeeklyStatsData {
	dailyStats := make(map[string]*DailyStatsData)
	for key, daily := range stats.DailyStats {
		dailyStats[key] = convertDailyStats(daily)
	}

	return &WeeklyStatsData{
		WeekStart:      stats.WeekStart.Format("2006-01-02"),
		WeekEnd:        stats.WeekEnd.Format("2006-01-02"),
		TotalTasks:     stats.TotalTasks,
		CompletedTasks: stats.CompletedTasks,
		CompletionRate: stats.CompletionRate,
		DailyStats:     dailyStats,
	}
}

func convertWeeklyPreview(preview *domain.WeeklyPreview) *WeeklyPreviewData {
	dailyPreview := make(map[string]*DailyPreviewData)
	for key, daily := range preview.DailyPreview {
		dailyPreview[key] = &DailyPreviewData{
			Date:       daily.Date.Format("2006-01-02"),
			TaskCount:  daily.TaskCount,
			HasOverdue: daily.HasOverdue,
		}
	}

	return &WeeklyPreviewData{
		WeekStart:    preview.WeekStart.Format("2006-01-02"),
		WeekEnd:      preview.WeekEnd.Format("2006-01-02"),
		TotalTasks:   preview.TotalTasks,
		DailyPreview: dailyPreview,
	}
}

func convertDailyStatsList(statsList []*domain.DailyStats) []DailyStatsData {
	result := make([]DailyStatsData, len(statsList))
	for i, stats := range statsList {
		result[i] = *convertDailyStats(stats)
	}
	return result
}

// getCategoryColor はカテゴリの色を取得する
func getCategoryColor(category domain.Category) string {
	colors := map[domain.Category]string{
		domain.CategoryWork:     "#3b82f6", // 青
		domain.CategoryPersonal: "#10b981", // 緑
		domain.CategoryStudy:    "#8b5cf6", // 紫
		domain.CategoryHealth:   "#f59e0b", // 黄
		domain.CategoryShopping: "#ef4444", // 赤
		domain.CategoryOther:    "#6b7280", // グレー
	}
	if color, exists := colors[category]; exists {
		return color
	}
	return "#6b7280" // デフォルトはグレー
}

// getPriorityColor は優先度の色を取得する
func getPriorityColor(priority domain.Priority) string {
	colors := map[domain.Priority]string{
		domain.PriorityHigh:   "#dc2626", // 赤
		domain.PriorityMedium: "#eab308", // 黄色
		domain.PriorityLow:    "#22c55e", // 緑
	}
	if color, exists := colors[priority]; exists {
		return color
	}
	return "#6b7280" // デフォルトはグレー
}
