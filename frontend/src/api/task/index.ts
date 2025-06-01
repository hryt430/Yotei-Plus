import { 
  ApiResponse, 
  Task, 
  TaskRequest, 
  TaskListResponse, 
  TaskResponse,
  TaskFilter,
  TaskStatus,
  TaskPriority,
  TaskCategory,
  DailyStats,
  WeeklyStats,
  DashboardStats,
  ProgressLevel
} from '@/types'
import { apiClient } from '@/api/client'

// === Basic Task Operations ===

// タスク一覧取得のパラメータ
export interface GetTasksParams extends TaskFilter {
  page?: number
  page_size?: number
  sort_field?: string
  sort_direction?: 'ASC' | 'DESC'
}

// タスク検索パラメータ
export interface SearchTasksParams {
  q: string
  limit?: number
}

// タスク割り当てリクエスト
export interface AssignTaskRequest {
  assignee_id: string
}

// タスク一覧を取得
export async function getTasks(params?: GetTasksParams): Promise<TaskListResponse> {
  const queryParams: Record<string, string> = {}
  
  if (params) {
    if (params.page) queryParams.page = params.page.toString()
    if (params.page_size) queryParams.page_size = params.page_size.toString()
    if (params.sort_field) queryParams.sort_field = params.sort_field
    if (params.sort_direction) queryParams.sort_direction = params.sort_direction
    if (params.status) queryParams.status = params.status
    if (params.priority) queryParams.priority = params.priority
    if (params.category) queryParams.category = params.category // ✅ category追加
    if (params.assignee_id) queryParams.assignee_id = params.assignee_id
    if (params.created_by) queryParams.created_by = params.created_by
    if (params.due_date_from) queryParams.due_date_from = params.due_date_from
    if (params.due_date_to) queryParams.due_date_to = params.due_date_to
    if (params.search) queryParams.search = params.search
  }

  return apiClient.get<TaskListResponse>('/tasks', queryParams)
}

// 新しいタスクを作成
export async function createTask(data: TaskRequest): Promise<TaskResponse> {
  return apiClient.post<TaskResponse>('/tasks', data)
}

// 特定のタスクを取得
export async function getTask(id: string): Promise<TaskResponse> {
  return apiClient.get<TaskResponse>(`/tasks/${id}`)
}

// getTaskById エイリアス（既存コードとの互換性のため）
export const getTaskById = getTask

// タスクを更新
export async function updateTask(id: string, data: TaskRequest): Promise<TaskResponse> {
  return apiClient.put<TaskResponse>(`/tasks/${id}`, data)
}

// タスクを削除
export async function deleteTask(id: string): Promise<ApiResponse<{ success: true }>> {
  return apiClient.delete<ApiResponse<{ success: true }>>(`/tasks/${id}`)
}

// タスクのステータスを変更
export async function changeTaskStatus(
  id: string, 
  status: TaskStatus
): Promise<TaskResponse> {
  return apiClient.put<TaskResponse>(`/tasks/${id}/status`, { status })
}

// タスクをユーザーに割り当てる
export async function assignTask(taskId: string, assigneeId: string): Promise<TaskResponse> {
  return apiClient.put<TaskResponse>(`/tasks/${taskId}/assign`, { 
    assignee_id: assigneeId 
  })
}

// 期限切れのタスクを取得
export async function getOverdueTasks(): Promise<ApiResponse<{
  tasks: Task[]
  count: number
}>> {
  return apiClient.get<ApiResponse<{
    tasks: Task[]
    count: number
  }>>('/tasks/overdue')
}

// 現在のユーザーのタスクを取得
export async function getMyTasks(): Promise<ApiResponse<{
  tasks: Task[]
  count: number
}>> {
  return apiClient.get<ApiResponse<{
    tasks: Task[]
    count: number
  }>>('/tasks/my')
}

// 特定のユーザーのタスクを取得
export async function getUserTasks(userId: string): Promise<ApiResponse<{
  tasks: Task[]
  count: number
}>> {
  return apiClient.get<ApiResponse<{
    tasks: Task[]
    count: number
  }>>(`/tasks/user/${userId}`)
}

// タスクを検索
export async function searchTasks(params: SearchTasksParams): Promise<ApiResponse<{
  tasks: Task[]
  count: number
}>> {
  const queryParams: Record<string, string> = {
    q: params.q
  }
  
  if (params.limit) {
    queryParams.limit = params.limit.toString()
  }

  return apiClient.get<ApiResponse<{
    tasks: Task[]
    count: number
  }>>('/tasks/search', queryParams)
}

// === Statistics API (完全対応) ===

// ダッシュボード統計を取得
export async function getDashboardStats(): Promise<ApiResponse<DashboardStats>> {
  return apiClient.get<ApiResponse<DashboardStats>>('/tasks/stats/dashboard')
}

// 今日の統計を取得
export async function getTodayStats(): Promise<ApiResponse<DailyStats>> {
  return apiClient.get<ApiResponse<DailyStats>>('/tasks/stats/today')
}

// 特定日の統計を取得
export async function getDailyStats(date: string): Promise<ApiResponse<DailyStats>> {
  return apiClient.get<ApiResponse<DailyStats>>(`/tasks/stats/daily/${date}`)
}

// 週次統計を取得
export async function getWeeklyStats(
  weekStart?: string
): Promise<ApiResponse<WeeklyStats>> {
  const params = weekStart ? { week_start: weekStart } : {}
  return apiClient.get<ApiResponse<WeeklyStats>>('/tasks/stats/weekly', params)
}

// 月次統計を取得
export async function getMonthlyStats(
  year?: number,
  month?: number
): Promise<ApiResponse<{
  year: number
  month: number
  total_tasks: number
  completed_tasks: number
  completion_rate: number
  weekly_breakdown: WeeklyStats[]
}>> {
  const params: Record<string, string> = {}
  if (year) params.year = year.toString()
  if (month) params.month = month.toString()
  
  return apiClient.get('/tasks/stats/monthly', params)
}

// 進捗サマリーを取得
export async function getProgressSummary(): Promise<ApiResponse<{
  overall_completion_rate: number
  today_completion_rate: number
  week_completion_rate: number
  month_completion_rate: number
  trend: 'up' | 'down' | 'stable'
}>> {
  return apiClient.get<ApiResponse<{
    overall_completion_rate: number
    today_completion_rate: number
    week_completion_rate: number
    month_completion_rate: number
    trend: 'up' | 'down' | 'stable'
  }>>('/tasks/stats/progress-summary')
}

// 進捗レベルを取得
export async function getProgressLevel(): Promise<ApiResponse<ProgressLevel>> {
  return apiClient.get<ApiResponse<ProgressLevel>>('/tasks/stats/progress-level')
}

// カテゴリ別の統計を取得
export async function getCategoryBreakdown(): Promise<ApiResponse<Record<TaskCategory, number>>> {
  return apiClient.get<ApiResponse<Record<TaskCategory, number>>>('/tasks/stats/category-breakdown')
}

// 優先度別の統計を取得
export async function getPriorityBreakdown(): Promise<ApiResponse<Record<TaskPriority, number>>> {
  return apiClient.get<ApiResponse<Record<TaskPriority, number>>>('/tasks/stats/priority-breakdown')
}

// === Custom Statistics Queries ===

// 期間指定での統計取得
export async function getStatsForDateRange(
  startDate: string,
  endDate: string
): Promise<ApiResponse<{
  daily_stats: DailyStats[]
  summary: {
    total_tasks: number
    completed_tasks: number
    completion_rate: number
  }
}>> {
  return apiClient.get('/tasks/stats/date-range', {
    start_date: startDate,
    end_date: endDate
  })
}

// ユーザー別統計取得
export async function getUserStats(
  userId: string,
  period?: 'today' | 'week' | 'month'
): Promise<ApiResponse<{
  user_id: string
  period: string
  total_tasks: number
  completed_tasks: number
  completion_rate: number
  category_breakdown: Record<TaskCategory, number>
  priority_breakdown: Record<TaskPriority, number>
}>> {
  const params: Record<string, string> = { user_id: userId }
  if (period) params.period = period
  
  return apiClient.get('/tasks/stats/user', params)
}

// === Utility Functions ===

// ステータス別タスク数を更新（フロントエンド用ヘルパー）
export async function updateTaskStatus(taskId: string, status: TaskStatus): Promise<TaskResponse> {
  return changeTaskStatus(taskId, status)
}

// タスク統計の簡易取得（後方互換性）
export async function getTaskStats(): Promise<ApiResponse<{
  total: number
  todo: number
  in_progress: number
  done: number
  overdue: number
}>> {
  const dashboardStats = await getDashboardStats()
  
  if (dashboardStats.success && dashboardStats.data) {
    const { today_stats } = dashboardStats.data
    return {
      success: true,
      data: {
        total: today_stats.total_tasks,
        todo: today_stats.todo_tasks,
        in_progress: today_stats.in_progress_tasks,
        done: today_stats.completed_tasks,
        overdue: today_stats.overdue_tasks
      }
    }
  }
  
  throw new Error('統計データの取得に失敗しました')
}

// カテゴリとプライオリティのコンスタント取得
export async function getTaskConstants(): Promise<ApiResponse<{
  categories: TaskCategory[]
  priorities: TaskPriority[]
  statuses: TaskStatus[]
}>> {
  // 型定義から定数を返す（バックエンドAPIがない場合）
  return {
    success: true,
    data: {
      categories: ['WORK', 'PERSONAL', 'STUDY', 'HEALTH', 'SHOPPING', 'OTHER'],
      priorities: ['LOW', 'MEDIUM', 'HIGH'],
      statuses: ['TODO', 'IN_PROGRESS', 'DONE']
    }
  }
}

// 統計データのキャッシュクリア
export async function clearStatsCache(): Promise<ApiResponse<{ message: string }>> {
  return apiClient.post<ApiResponse<{ message: string }>>('/tasks/stats/clear-cache')
}