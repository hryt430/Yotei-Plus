import { 
  ApiResponse, 
  Task, 
  TaskRequest, 
  TaskListResponse, 
  TaskResponse,
  TaskFilter,
  User
} from '@/types'
import { apiClient } from '@/api/client'

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

// タスク統計レスポンス
export interface TaskStatsResponse {
  total: number
  todo: number
  in_progress: number
  done: number
  overdue: number
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
    if (params.assignee_id) queryParams.assignee_id = params.assignee_id
    if (params.created_by) queryParams.created_by = params.created_by
    if (params.due_date_from) queryParams.due_date_from = params.due_date_from
    if (params.due_date_to) queryParams.due_date_to = params.due_date_to
    if (params.search) queryParams.search = params.search
  }

  return apiClient.get<TaskListResponse>('/api/tasks', queryParams)
}

// 新しいタスクを作成
export async function createTask(data: TaskRequest): Promise<TaskResponse> {
  return apiClient.post<TaskResponse>('/api/tasks', data)
}

// 特定のタスクを取得
export async function getTask(id: string): Promise<TaskResponse> {
  return apiClient.get<TaskResponse>(`/api/tasks/${id}`)
}

// getTaskById エイリアス（既存コードとの互換性のため）
export const getTaskById = getTask

// タスクを更新
export async function updateTask(id: string, data: TaskRequest): Promise<TaskResponse> {
  return apiClient.put<TaskResponse>(`/api/tasks/${id}`, data)
}

// タスクを削除
export async function deleteTask(id: string): Promise<ApiResponse<{ success: true }>> {
  return apiClient.delete<ApiResponse<{ success: true }>>(`/api/tasks/${id}`)
}

// タスクのステータスを変更
export async function changeTaskStatus(
  id: string, 
  status: 'TODO' | 'IN_PROGRESS' | 'DONE'
): Promise<TaskResponse> {
  return apiClient.patch<TaskResponse>(`/api/tasks/${id}/status`, { status })
}

// タスクをユーザーに割り当てる
export async function assignTask(taskId: string, assigneeId: string): Promise<TaskResponse> {
  return apiClient.post<TaskResponse>(`/api/tasks/${taskId}/assign`, { 
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
  }>>('/api/tasks/overdue')
}

// 現在のユーザーのタスクを取得
export async function getMyTasks(): Promise<ApiResponse<{
  tasks: Task[]
  count: number
}>> {
  return apiClient.get<ApiResponse<{
    tasks: Task[]
    count: number
  }>>('/api/tasks/my')
}

// 特定のユーザーのタスクを取得
export async function getUserTasks(userId: string): Promise<ApiResponse<{
  tasks: Task[]
  count: number
}>> {
  return apiClient.get<ApiResponse<{
    tasks: Task[]
    count: number
  }>>(`/api/tasks/user/${userId}`)
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
  }>>('/api/tasks/search', queryParams)
}

// タスク統計を取得
export async function getTaskStats(): Promise<ApiResponse<TaskStatsResponse>> {
  return apiClient.get<ApiResponse<TaskStatsResponse>>('/api/tasks/stats')
}

// ステータス別タスク数を更新（フロントエンド用ヘルパー）
export async function updateTaskStatus(taskId: string, status: Task['status']): Promise<TaskResponse> {
  return changeTaskStatus(taskId, status)
}