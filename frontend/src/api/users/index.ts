import { 
  ApiResponse, 
  User,
  UserResponse 
} from '@/types'
import { apiClient } from '@/api/client'

// === User Management APIs ===

// ユーザー一覧取得のパラメータ
export interface GetUsersParams {
  search?: string
  limit?: number
  offset?: number
  role?: 'user' | 'admin'
  department?: string
  status?: 'active' | 'inactive'
  sort_field?: 'username' | 'email' | 'created_at' | 'last_login'
  sort_direction?: 'ASC' | 'DESC'
}

// ユーザー一覧レスポンス
export interface UsersListResponse {
  success: boolean
  data: User[]
  total_count?: number
  pagination?: {
    page: number
    page_size: number
    total_pages: number
  }
}

// ユーザー作成用の入力データ
export interface CreateUserInput {
  username: string
  email: string
  password?: string
  role?: 'user' | 'admin'
  department?: string
  profile?: {
    first_name?: string
    last_name?: string
    phone?: string
    avatar_url?: string
  }
}

// ユーザー更新用の入力データ
export interface UpdateUserInput {
  username?: string
  email?: string
  role?: 'user' | 'admin'
  department?: string
  status?: 'active' | 'inactive'
  profile?: {
    first_name?: string
    last_name?: string
    phone?: string
    avatar_url?: string
  }
}

// === Basic User Operations ===

// ユーザー一覧を取得（タスク割り当て等で使用）
export async function getUsers(params?: GetUsersParams): Promise<UsersListResponse> {
  const queryParams: Record<string, string> = {}
  
  if (params) {
    if (params.search) queryParams.search = params.search
    if (params.limit) queryParams.limit = params.limit.toString()
    if (params.offset) queryParams.offset = params.offset.toString()
    if (params.role) queryParams.role = params.role
    if (params.department) queryParams.department = params.department
    if (params.status) queryParams.status = params.status
    if (params.sort_field) queryParams.sort_field = params.sort_field
    if (params.sort_direction) queryParams.sort_direction = params.sort_direction
  }

  return apiClient.get<UsersListResponse>('/users', queryParams)
}

// 特定のユーザー情報を取得
export async function getUserById(id: string): Promise<ApiResponse<User>> {
  return apiClient.get<ApiResponse<User>>(`/users/${id}`)
}

// 現在のユーザー情報を取得（/users/me エンドポイント使用）
export async function getCurrentUserProfile(): Promise<UserResponse> {
  return apiClient.get<UserResponse>('/users/me')
}

// 現在のユーザー情報を更新
export async function updateCurrentUser(data: UpdateUserInput): Promise<ApiResponse<User>> {
  return apiClient.put<ApiResponse<User>>('/users/me', data)
}

// === Admin User Operations ===

// 新しいユーザーを作成（管理者用）
export async function createUser(data: CreateUserInput): Promise<ApiResponse<User>> {
  return apiClient.post<ApiResponse<User>>('/users', data)
}

// ユーザー情報を更新（管理者用）
export async function updateUser(id: string, data: UpdateUserInput): Promise<ApiResponse<User>> {
  return apiClient.put<ApiResponse<User>>(`/users/${id}`, data)
}

// ユーザーを削除（管理者用）
export async function deleteUser(id: string): Promise<ApiResponse<{ success: true }>> {
  return apiClient.delete<ApiResponse<{ success: true }>>(`/users/${id}`)
}

// ユーザーのステータスを変更（管理者用）
export async function updateUserStatus(
  id: string, 
  status: 'active' | 'inactive'
): Promise<ApiResponse<User>> {
  return apiClient.patch<ApiResponse<User>>(`/users/${id}/status`, { status })
}

// ユーザーの役割を変更（管理者用）
export async function updateUserRole(
  id: string, 
  role: 'user' | 'admin'
): Promise<ApiResponse<User>> {
  return apiClient.patch<ApiResponse<User>>(`/users/${id}/role`, { role })
}

// === File Upload ===

// ユーザーのプロフィール画像をアップロード
export async function uploadUserAvatar(
  id: string, 
  file: File
): Promise<ApiResponse<{ avatar_url: string }>> {
  return apiClient.uploadFile<ApiResponse<{ avatar_url: string }>>(
    `/users/${id}/avatar`, 
    file, 
    'avatar'
  )
}

// 現在のユーザーのプロフィール画像をアップロード
export async function uploadCurrentUserAvatar(
  file: File
): Promise<ApiResponse<{ avatar_url: string }>> {
  return apiClient.uploadFile<ApiResponse<{ avatar_url: string }>>(
    '/users/me/avatar', 
    file, 
    'avatar'
  )
}

// === Search and Analytics ===

// ユーザー検索（名前、メール、部署での検索）
export async function searchUsers(
  query: string, 
  limit: number = 10
): Promise<ApiResponse<User[]>> {
  return apiClient.get<ApiResponse<User[]>>('/users/search', { 
    q: query, 
    limit: limit.toString() 
  })
}

// 部署別ユーザー数を取得
export async function getUsersByDepartment(): Promise<ApiResponse<{
  departments: Array<{
    name: string
    user_count: number
    users: User[]
  }>
}>> {
  return apiClient.get<ApiResponse<{
    departments: Array<{
      name: string
      user_count: number
      users: User[]
    }>
  }>>('/users/by-department')
}

// アクティブユーザー数を取得
export async function getActiveUserCount(): Promise<ApiResponse<{ count: number }>> {
  return apiClient.get<ApiResponse<{ count: number }>>('/users/active/count')
}

// ユーザーの統計情報を取得
export async function getUserStats(): Promise<ApiResponse<{
  total_users: number
  active_users: number
  admin_users: number
  recent_registrations: number
  departments: Array<{
    name: string
    count: number
  }>
}>> {
  return apiClient.get<ApiResponse<{
    total_users: number
    active_users: number
    admin_users: number
    recent_registrations: number
    departments: Array<{
      name: string
      count: number
    }>
  }>>('/users/stats')
}

// === Task Assignment Related ===

// ユーザーのタスク割り当て可能性をチェック
export async function checkUserAvailability(userId: string): Promise<ApiResponse<{
  available: boolean
  current_tasks_count: number
  capacity: number
  workload_percentage: number
}>> {
  return apiClient.get<ApiResponse<{
    available: boolean
    current_tasks_count: number
    capacity: number
    workload_percentage: number
  }>>(`/users/${userId}/availability`)
}

// ユーザーの現在のタスク負荷を取得
export async function getUserWorkload(userId: string): Promise<ApiResponse<{
  user_id: string
  total_tasks: number
  todo_tasks: number
  in_progress_tasks: number
  overdue_tasks: number
  workload_score: number
  recommended_capacity: number
}>> {
  return apiClient.get<ApiResponse<{
    user_id: string
    total_tasks: number
    todo_tasks: number
    in_progress_tasks: number
    overdue_tasks: number
    workload_score: number
    recommended_capacity: number
  }>>(`/users/${userId}/workload`)
}

// タスク割り当てに最適なユーザーを取得
export async function getOptimalUsersForTask(
  priority: 'LOW' | 'MEDIUM' | 'HIGH',
  category: string
): Promise<ApiResponse<{
  recommended_users: Array<User & {
    workload_score: number
    expertise_score: number
    availability_score: number
    total_score: number
  }>
}>> {
  return apiClient.get<ApiResponse<{
    recommended_users: Array<User & {
      workload_score: number
      expertise_score: number
      availability_score: number
      total_score: number
    }>
  }>>('/users/optimal-for-task', {
    priority,
    category
  })
}

// === User Activity ===

// ユーザーの最近のアクティビティを取得
export async function getUserActivity(
  userId: string,
  limit: number = 20
): Promise<ApiResponse<{
  activities: Array<{
    id: string
    type: 'task_created' | 'task_completed' | 'task_assigned' | 'login' | 'profile_updated'
    description: string
    timestamp: string
    metadata?: Record<string, any>
  }>
}>> {
  return apiClient.get<ApiResponse<{
    activities: Array<{
      id: string
      type: 'task_created' | 'task_completed' | 'task_assigned' | 'login' | 'profile_updated'
      description: string
      timestamp: string
      metadata?: Record<string, any>
    }>
  }>>(`/users/${userId}/activity`, { limit: limit.toString() })
}

// 現在のユーザーのアクティビティを取得
export async function getCurrentUserActivity(
  limit: number = 20
): Promise<ApiResponse<{
  activities: Array<{
    id: string
    type: 'task_created' | 'task_completed' | 'task_assigned' | 'login' | 'profile_updated'
    description: string
    timestamp: string
    metadata?: Record<string, any>
  }>
}>> {
  return apiClient.get<ApiResponse<{
    activities: Array<{
      id: string
      type: 'task_created' | 'task_completed' | 'task_assigned' | 'login' | 'profile_updated'
      description: string
      timestamp: string
      metadata?: Record<string, any>
    }>
  }>>('/users/me/activity', { limit: limit.toString() })
}

// === Batch Operations ===

// 複数ユーザーの情報を一括取得
export async function getUsersBatch(userIds: string[]): Promise<ApiResponse<{
  users: User[]
  not_found: string[]
}>> {
  return apiClient.post<ApiResponse<{
    users: User[]
    not_found: string[]
  }>>('/users/batch', { user_ids: userIds })
}

// 複数ユーザーのステータスを一括更新（管理者用）
export async function updateUsersBatchStatus(
  userIds: string[],
  status: 'active' | 'inactive'
): Promise<ApiResponse<{
  updated_users: User[]
  failed_updates: Array<{
    user_id: string
    error: string
  }>
}>> {
  return apiClient.patch<ApiResponse<{
    updated_users: User[]
    failed_updates: Array<{
      user_id: string
      error: string
    }>
  }>>('/users/batch/status', {
    user_ids: userIds,
    status
  })
}

// === Utility Functions ===

// ユーザーがオンラインかどうかを確認
export async function checkUserOnlineStatus(userId: string): Promise<ApiResponse<{
  is_online: boolean
  last_seen?: string
}>> {
  return apiClient.get<ApiResponse<{
    is_online: boolean
    last_seen?: string
  }>>(`/users/${userId}/online-status`)
}

// ユーザーの設定を取得
export async function getUserSettings(userId?: string): Promise<ApiResponse<{
  notifications: {
    email: boolean
    push: boolean
    in_app: boolean
  }
  preferences: {
    theme: 'light' | 'dark' | 'auto'
    language: string
    timezone: string
  }
  privacy: {
    profile_visibility: 'public' | 'private'
    activity_visibility: 'public' | 'private'
  }
}>> {
  const endpoint = userId ? `/users/${userId}/settings` : '/users/me/settings'
  return apiClient.get<ApiResponse<{
    notifications: {
      email: boolean
      push: boolean
      in_app: boolean
    }
    preferences: {
      theme: 'light' | 'dark' | 'auto'
      language: string
      timezone: string
    }
    privacy: {
      profile_visibility: 'public' | 'private'
      activity_visibility: 'public' | 'private'
    }
  }>>(endpoint)
}

// ユーザーの設定を更新
export async function updateUserSettings(
  settings: {
    notifications?: {
      email?: boolean
      push?: boolean
      in_app?: boolean
    }
    preferences?: {
      theme?: 'light' | 'dark' | 'auto'
      language?: string
      timezone?: string
    }
    privacy?: {
      profile_visibility?: 'public' | 'private'
      activity_visibility?: 'public' | 'private'
    }
  },
  userId?: string
): Promise<ApiResponse<{ message: string }>> {
  const endpoint = userId ? `/users/${userId}/settings` : '/users/me/settings'
  return apiClient.put<ApiResponse<{ message: string }>>(endpoint, settings)
}