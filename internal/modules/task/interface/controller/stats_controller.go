package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hryt430/Yotei+/internal/modules/task/domain"
	"github.com/hryt430/Yotei+/internal/modules/task/usecase"
	"github.com/hryt430/Yotei+/pkg/utils"
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

// DashboardStatsResponse はダッシュボード統計のレスポンス
type DashboardStatsResponse struct {
	Success bool                   `json:"success"`
	Data    *domain.DashboardStats `json:"data"`
}

// DailyStatsResponse は日次統計のレスポンス
type DailyStatsResponse struct {
	Success bool               `json:"success"`
	Data    *domain.DailyStats `json:"data"`
}

// WeeklyStatsResponse は週次統計のレスポンス
type WeeklyStatsResponse struct {
	Success bool                `json:"success"`
	Data    *domain.WeeklyStats `json:"data"`
}

// ProgressSummaryResponse は進捗サマリーのレスポンス
type ProgressSummaryResponse struct {
	Success bool                 `json:"success"`
	Data    []*domain.DailyStats `json:"data"`
}

// ProgressLevelResponse は進捗レベルのレスポンス
type ProgressLevelResponse struct {
	Success bool                  `json:"success"`
	Data    *domain.ProgressLevel `json:"data"`
}

// GetDashboardStats はダッシュボード用の統計情報を取得する
func (c *TaskStatsController) GetDashboardStats(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse(err.Error()))
		return
	}

	stats, err := c.statsService.GetDashboardStats(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to get dashboard stats"))
		return
	}

	ctx.JSON(http.StatusOK, DashboardStatsResponse{
		Success: true,
		Data:    stats,
	})
}

// GetTodayStats は今日の統計情報を取得する
func (c *TaskStatsController) GetTodayStats(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse(err.Error()))
		return
	}

	today := time.Now()
	stats, err := c.statsService.GetDailyStats(ctx, userID, today)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to get today stats"))
		return
	}

	ctx.JSON(http.StatusOK, DailyStatsResponse{
		Success: true,
		Data:    stats,
	})
}

// GetDailyStats は指定日の統計情報を取得する
func (c *TaskStatsController) GetDailyStats(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse(err.Error()))
		return
	}

	// 日付パラメータの取得
	dateStr := ctx.Param("date")
	if dateStr == "" {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Date parameter is required"))
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format. Use YYYY-MM-DD"))
		return
	}

	stats, err := c.statsService.GetDailyStats(ctx, userID, date)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to get daily stats"))
		return
	}

	ctx.JSON(http.StatusOK, DailyStatsResponse{
		Success: true,
		Data:    stats,
	})
}

// GetWeeklyStats は指定週の統計情報を取得する
func (c *TaskStatsController) GetWeeklyStats(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse(err.Error()))
		return
	}

	// 日付パラメータの取得（その週に含まれる任意の日付）
	dateStr := ctx.Query("date")
	var date time.Time
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format. Use YYYY-MM-DD"))
			return
		}
		date = parsedDate
	} else {
		date = time.Now() // デフォルトは今週
	}

	stats, err := c.statsService.GetWeeklyStats(ctx, userID, date)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to get weekly stats"))
		return
	}

	ctx.JSON(http.StatusOK, WeeklyStatsResponse{
		Success: true,
		Data:    stats,
	})
}

// GetProgressSummary は指定期間の進捗サマリーを取得する
func (c *TaskStatsController) GetProgressSummary(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse(err.Error()))
		return
	}

	// 日数パラメータの取得
	daysStr := ctx.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid days parameter. Must be between 1 and 365"))
		return
	}

	summary, err := c.statsService.GetProgressSummary(ctx, userID, days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to get progress summary"))
		return
	}

	ctx.JSON(http.StatusOK, ProgressSummaryResponse{
		Success: true,
		Data:    summary,
	})
}

// GetProgressLevel は進捗レベル情報を取得する
func (c *TaskStatsController) GetProgressLevel(ctx *gin.Context) {
	// 完了率パラメータの取得
	rateStr := ctx.Query("rate")
	if rateStr == "" {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Completion rate parameter is required"))
		return
	}

	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil || rate < 0 || rate > 100 {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid completion rate. Must be between 0 and 100"))
		return
	}

	level := domain.GetProgressLevel(rate)

	ctx.JSON(http.StatusOK, ProgressLevelResponse{
		Success: true,
		Data:    &level,
	})
}

// GetCategoryBreakdown はカテゴリ別の統計を取得する
func (c *TaskStatsController) GetCategoryBreakdown(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse(err.Error()))
		return
	}

	breakdown, err := c.statsService.GetCategoryBreakdown(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to get category breakdown"))
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

// GetPriorityBreakdown は優先度別の統計を取得する
func (c *TaskStatsController) GetPriorityBreakdown(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse(err.Error()))
		return
	}

	breakdown, err := c.statsService.GetPriorityBreakdown(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to get priority breakdown"))
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

// GetMonthlyStats は月次統計を取得する
func (c *TaskStatsController) GetMonthlyStats(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorResponse(err.Error()))
		return
	}

	// 年月パラメータの取得
	yearStr := ctx.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	monthStr := ctx.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 3000 {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid year parameter"))
		return
	}

	monthInt, err := strconv.Atoi(monthStr)
	if err != nil || monthInt < 1 || monthInt > 12 {
		ctx.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid month parameter"))
		return
	}

	month := time.Month(monthInt)

	stats, err := c.statsService.GetMonthlyStats(ctx, userID, year, month)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to get monthly stats"))
		return
	}

	ctx.JSON(http.StatusOK, WeeklyStatsResponse{
		Success: true,
		Data:    stats,
	})
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
		domain.PriorityHigh: "#dc2626", // 赤
		// domain.PriorityHigh:   "#f97316", // オレンジ
		domain.PriorityMedium: "#eab308", // 黄色
		domain.PriorityLow:    "#22c55e", // 緑
	}
	if color, exists := colors[priority]; exists {
		return color
	}
	return "#6b7280" // デフォルトはグレー
}
