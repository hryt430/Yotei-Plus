// src/api/tasks/route.ts
// タスク関連のAPI関数

import { 
  ApiResponse, 
  Task, 
  TaskRequest, 
  TaskListResponse, 
  TaskResponse,
  TaskFilter,
  Pagination,
  SortOptions
} from '@/types'
import { apiClient } from '@/api/client/route'

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