import { useState, useEffect, useCallback } from 'react'
import {
  DashboardStats,
  DailyStats,
  WeeklyStats,
  ProgressLevel,
  TaskCategory,
  TaskPriority,
  ApiResponse
} from '@/types'
import {
  getDashboardStats,
  getTodayStats,
  getDailyStats,
  getWeeklyStats,
  getMonthlyStats,
  getProgressLevel,
  getCategoryBreakdown,
  getPriorityBreakdown,
  getProgressSummary
} from '@/api/task'
import { handleApiError } from '@/lib/utils'

// 統計の状態管理
interface StatsState {
  dashboardStats: DashboardStats | null
  todayStats: DailyStats | null
  weeklyStats: WeeklyStats | null
  progressLevel: ProgressLevel | null
  categoryBreakdown: Record<TaskCategory, number> | null
  priorityBreakdown: Record<TaskPriority, number> | null
  progressSummary: {
    overall_completion_rate: number
    today_completion_rate: number
    week_completion_rate: number
    month_completion_rate: number
    trend: 'up' | 'down' | 'stable'
  } | null
  isLoading: boolean
  error: string | null
  lastUpdated: Date | null
}

// キャッシュ設定
interface CacheConfig {
  dashboardCacheDuration: number // ダッシュボード統計のキャッシュ時間（秒）
  dailyCacheDuration: number     // 日次統計のキャッシュ時間（秒）
  weeklyCacheDuration: number    // 週次統計のキャッシュ時間（秒）
}

const DEFAULT_CACHE_CONFIG: CacheConfig = {
  dashboardCacheDuration: 300,  // 5分
  dailyCacheDuration: 600,      // 10分
  weeklyCacheDuration: 1800,    // 30分
}

export function useTaskStats(cacheConfig: Partial<CacheConfig> = {}) {
  const config = { ...DEFAULT_CACHE_CONFIG, ...cacheConfig }
  
  const [state, setState] = useState<StatsState>({
    dashboardStats: null,
    todayStats: null,
    weeklyStats: null,
    progressLevel: null,
    categoryBreakdown: null,
    priorityBreakdown: null,
    progressSummary: null,
    isLoading: false,
    error: null,
    lastUpdated: null,
  })

  // キャッシュ有効性をチェック
  const isCacheValid = useCallback((duration: number): boolean => {
    if (!state.lastUpdated) return false
    const now = new Date()
    const diff = (now.getTime() - state.lastUpdated.getTime()) / 1000
    return diff < duration
  }, [state.lastUpdated])

  // エラーハンドリング用ヘルパー
  const handleError = useCallback((error: any, context: string) => {
    const errorMessage = handleApiError(error)
    console.error(`${context} error:`, error)
    setState(prev => ({ 
      ...prev, 
      error: errorMessage, 
      isLoading: false 
    }))
  }, [])

  // ダッシュボード統計を取得
  const fetchDashboardStats = useCallback(async (forceRefresh = false) => {
    if (!forceRefresh && state.dashboardStats && isCacheValid(config.dashboardCacheDuration)) {
      return state.dashboardStats
    }

    setState(prev => ({ ...prev, isLoading: true, error: null }))
    
    try {
      const response = await getDashboardStats()
      
      if (response.success) {
        setState(prev => ({
          ...prev,
          dashboardStats: response.data,
          isLoading: false,
          lastUpdated: new Date(),
        }))
        return response.data
      }
    } catch (error) {
      handleError(error, 'Dashboard stats fetch')
    }
    
    return null
  }, [state.dashboardStats, isCacheValid, config.dashboardCacheDuration, handleError])

  // 今日の統計を取得
  const fetchTodayStats = useCallback(async (forceRefresh = false) => {
    if (!forceRefresh && state.todayStats && isCacheValid(config.dailyCacheDuration)) {
      return state.todayStats
    }

    setState(prev => ({ ...prev, isLoading: true, error: null }))
    
    try {
      const response = await getTodayStats()
      
      if (response.success) {
        setState(prev => ({
          ...prev,
          todayStats: response.data,
          isLoading: false,
          lastUpdated: new Date(),
        }))
        return response.data
      }
    } catch (error) {
      handleError(error, 'Today stats fetch')
    }
    
    return null
  }, [state.todayStats, isCacheValid, config.dailyCacheDuration, handleError])

  // 特定日の統計を取得
  const fetchDailyStats = useCallback(async (date: string) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }))
    
    try {
      const response = await getDailyStats(date)
      
      if (response.success) {
        setState(prev => ({ ...prev, isLoading: false }))
        return response.data
      }
    } catch (error) {
      handleError(error, `Daily stats fetch for ${date}`)
    }
    
    return null
  }, [handleError])

  // 週次統計を取得
  const fetchWeeklyStats = useCallback(async (weekStart?: string, forceRefresh = false) => {
    if (!forceRefresh && !weekStart && state.weeklyStats && isCacheValid(config.weeklyCacheDuration)) {
      return state.weeklyStats
    }

    setState(prev => ({ ...prev, isLoading: true, error: null }))
    
    try {
      const response = await getWeeklyStats(weekStart)
      
      if (response.success) {
        setState(prev => ({
          ...prev,
          weeklyStats: response.data,
          isLoading: false,
          lastUpdated: new Date(),
        }))
        return response.data
      }
    } catch (error) {
      handleError(error, 'Weekly stats fetch')
    }
    
    return null
  }, [state.weeklyStats, isCacheValid, config.weeklyCacheDuration, handleError])

  // 進捗レベルを取得
  const fetchProgressLevel = useCallback(async (forceRefresh = false) => {
    if (!forceRefresh && state.progressLevel && isCacheValid(config.dailyCacheDuration)) {
      return state.progressLevel
    }

    try {
      const response = await getProgressLevel()
      
      if (response.success) {
        setState(prev => ({
          ...prev,
          progressLevel: response.data,
          lastUpdated: new Date(),
        }))
        return response.data
      }
    } catch (error) {
      handleError(error, 'Progress level fetch')
    }
    
    return null
  }, [state.progressLevel, isCacheValid, config.dailyCacheDuration, handleError])

  // カテゴリ別統計を取得
  const fetchCategoryBreakdown = useCallback(async (forceRefresh = false) => {
    if (!forceRefresh && state.categoryBreakdown && isCacheValid(config.dailyCacheDuration)) {
      return state.categoryBreakdown
    }

    try {
      const response = await getCategoryBreakdown()
      
      if (response.success) {
        setState(prev => ({
          ...prev,
          categoryBreakdown: response.data,
          lastUpdated: new Date(),
        }))
        return response.data
      }
    } catch (error) {
      handleError(error, 'Category breakdown fetch')
    }
    
    return null
  }, [state.categoryBreakdown, isCacheValid, config.dailyCacheDuration, handleError])

  // 優先度別統計を取得
  const fetchPriorityBreakdown = useCallback(async (forceRefresh = false) => {
    if (!forceRefresh && state.priorityBreakdown && isCacheValid(config.dailyCacheDuration)) {
      return state.priorityBreakdown
    }

    try {
      const response = await getPriorityBreakdown()
      
      if (response.success) {
        setState(prev => ({
          ...prev,
          priorityBreakdown: response.data,
          lastUpdated: new Date(),
        }))
        return response.data
      }
    } catch (error) {
      handleError(error, 'Priority breakdown fetch')
    }
    
    return null
  }, [state.priorityBreakdown, isCacheValid, config.dailyCacheDuration, handleError])

  // 進捗サマリーを取得
  const fetchProgressSummary = useCallback(async (forceRefresh = false) => {
    if (!forceRefresh && state.progressSummary && isCacheValid(config.dailyCacheDuration)) {
      return state.progressSummary
    }

    try {
      const response = await getProgressSummary()
      
      if (response.success) {
        setState(prev => ({
          ...prev,
          progressSummary: response.data,
          lastUpdated: new Date(),
        }))
        return response.data
      }
    } catch (error) {
      handleError(error, 'Progress summary fetch')
    }
    
    return null
  }, [state.progressSummary, isCacheValid, config.dailyCacheDuration, handleError])

  // 月次統計を取得
  const fetchMonthlyStats = useCallback(async (year?: number, month?: number) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }))
    
    try {
      const response = await getMonthlyStats(year, month)
      
      if (response.success) {
        setState(prev => ({ ...prev, isLoading: false }))
        return response.data
      }
    } catch (error) {
      handleError(error, 'Monthly stats fetch')
    }
    
    return null
  }, [handleError])

  // 全統計データを一括取得
  const fetchAllStats = useCallback(async (forceRefresh = false) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }))
    
    try {
      const [
        dashboard,
        today,
        weekly,
        progress,
        category,
        priority,
        summary
      ] = await Promise.allSettled([
        fetchDashboardStats(forceRefresh),
        fetchTodayStats(forceRefresh),
        fetchWeeklyStats(undefined, forceRefresh),
        fetchProgressLevel(forceRefresh),
        fetchCategoryBreakdown(forceRefresh),
        fetchPriorityBreakdown(forceRefresh),
        fetchProgressSummary(forceRefresh)
      ])

      setState(prev => ({ ...prev, isLoading: false }))
      
      // 失敗した場合のログ出力
      const failures = [dashboard, today, weekly, progress, category, priority, summary]
        .map((result, index) => ({ result, index }))
        .filter(({ result }) => result.status === 'rejected')
      
      if (failures.length > 0) {
        console.warn('Some stats failed to load:', failures)
      }
      
    } catch (error) {
      handleError(error, 'Batch stats fetch')
    }
  }, [
    fetchDashboardStats,
    fetchTodayStats,
    fetchWeeklyStats,
    fetchProgressLevel,
    fetchCategoryBreakdown,
    fetchPriorityBreakdown,
    fetchProgressSummary,
    handleError
  ])

  // キャッシュをクリア
  const clearCache = useCallback(() => {
    setState(prev => ({
      ...prev,
      dashboardStats: null,
      todayStats: null,
      weeklyStats: null,
      progressLevel: null,
      categoryBreakdown: null,
      priorityBreakdown: null,
      progressSummary: null,
      lastUpdated: null,
    }))
  }, [])

  // エラーをクリア
  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }))
  }, [])

  // コンポーネントマウント時に統計データを取得
  useEffect(() => {
    fetchAllStats()
  }, []) // fetchAllStatsは含めない（無限ループ防止）

  // 定期的な自動更新（オプション）
  useEffect(() => {
    const interval = setInterval(() => {
      if (!state.isLoading) {
        fetchAllStats()
      }
    }, config.dashboardCacheDuration * 1000) // 最も短いキャッシュ時間で更新

    return () => clearInterval(interval)
  }, [state.isLoading, config.dashboardCacheDuration])

  return {
    // 状態
    ...state,
    
    // 個別取得メソッド
    fetchDashboardStats,
    fetchTodayStats,
    fetchDailyStats,
    fetchWeeklyStats,
    fetchMonthlyStats,
    fetchProgressLevel,
    fetchCategoryBreakdown,
    fetchPriorityBreakdown,
    fetchProgressSummary,
    
    // 一括操作
    fetchAllStats,
    clearCache,
    clearError,
    
    // ユーティリティ
    isCacheValid: (duration: number) => isCacheValid(duration),
    isDataFresh: isCacheValid(config.dashboardCacheDuration),
  }
}

// 統計データ用のユーティリティフック
export function useStatsCalculations() {
  // 完了率の計算
  const calculateCompletionRate = useCallback((completed: number, total: number): number => {
    if (total === 0) return 0
    return Math.round((completed / total) * 100)
  }, [])

  // 進捗の色を取得
  const getProgressColor = useCallback((rate: number): string => {
    if (rate >= 100) return '#22c55e' // ColorDarkGreen
    if (rate >= 80) return '#84cc16'  // ColorGreen
    if (rate >= 60) return '#eab308'  // ColorYellow
    if (rate >= 40) return '#f97316'  // ColorOrange
    if (rate >= 20) return '#ef4444'  // ColorLightRed
    if (rate >= 1) return '#dc2626'   // ColorRed
    return '#9ca3af'                  // ColorGray
  }, [])

  // 進捗のラベルを取得
  const getProgressLabel = useCallback((rate: number): string => {
    if (rate >= 100) return '完了'
    if (rate >= 80) return '優秀'
    if (rate >= 60) return '良好'
    if (rate >= 40) return '普通'
    if (rate >= 20) return '要改善'
    if (rate >= 1) return '低調'
    return '未着手'
  }, [])

  // 週の開始日と終了日を取得
  const getWeekRange = useCallback((date: Date = new Date()): { start: Date; end: Date } => {
    const start = new Date(date)
    const day = start.getDay()
    const diff = start.getDate() - day + (day === 0 ? -6 : 1) // Monday as start
    start.setDate(diff)
    start.setHours(0, 0, 0, 0)

    const end = new Date(start)
    end.setDate(start.getDate() + 6)
    end.setHours(23, 59, 59, 999)

    return { start, end }
  }, [])

  // 月の開始日と終了日を取得
  const getMonthRange = useCallback((date: Date = new Date()): { start: Date; end: Date } => {
    const start = new Date(date.getFullYear(), date.getMonth(), 1)
    const end = new Date(date.getFullYear(), date.getMonth() + 1, 0, 23, 59, 59, 999)
    return { start, end }
  }, [])

  // トレンドの方向を計算
  const calculateTrend = useCallback((current: number, previous: number): 'up' | 'down' | 'stable' => {
    const threshold = 5 // 5%の閾値
    const diff = current - previous
    
    if (Math.abs(diff) < threshold) return 'stable'
    return diff > 0 ? 'up' : 'down'
  }, [])

  return {
    calculateCompletionRate,
    getProgressColor,
    getProgressLabel,
    getWeekRange,
    getMonthRange,
    calculateTrend,
  }
}