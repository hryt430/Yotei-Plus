import { 
  ApiResponse, 
  User,
  UserResponse 
} from '@/types'
import { apiClient } from '@/api/client'

// ユーザー一覧取得のパラメータ
export interface GetUsersParams {
  search?: string
  limit?: number
  offset?: number
  role?: string
  department?: string
  status?: 'active' | 'inactive'
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
  role?: string
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
  role?: string
  department?: string
  status?: 'active' | 'inactive'
  profile?: {
    first_name?: string
    last_name?: string
    phone?: string
    avatar_url?: string
  }
}

// ユーザー一覧を取得
export async function getUsers(params?: GetUsersParams): Promise<UsersListResponse> {
  const queryParams: Record<string, string> = {}
  
  if (params) {
    if (params.search) queryParams.search = params.search
    if (params.limit) queryParams.limit = params.limit.toString()
    if (params.offset) queryParams.offset = params.offset.toString()
    if (params.role) queryParams.role = params.role
    if (params.department) queryParams.department = params.department
    if (params.status) queryParams.status = params.status
  }

  return apiClient.get<UsersListResponse>('/users', queryParams)
}

// 特定のユーザー情報を取得
export async function getUserById(id: string): Promise<ApiResponse<User>> {
  return apiClient.get<ApiResponse<User>>(`/users/${id}`)
}

// 新しいユーザーを作成（管理者用）
export async function createUser(data: CreateUserInput): Promise<ApiResponse<User>> {
  return apiClient.post<ApiResponse<User>>('/users', data)
}

// ユーザー情報を更新
export async function updateUser(id: string, data: UpdateUserInput): Promise<ApiResponse<User>> {
  return apiClient.put<ApiResponse<User>>(`/users/${id}`, data)
}

// ユーザーを削除
export async function deleteUser(id: string): Promise<ApiResponse<{ success: true }>> {
  return apiClient.delete<ApiResponse<{ success: true }>>(`/users/${id}`)
}

// ユーザーのステータスを変更（アクティブ/非アクティブ）
export async function updateUserStatus(
  id: string, 
  status: 'active' | 'inactive'
): Promise<ApiResponse<User>> {
  return apiClient.patch<ApiResponse<User>>(`/users/${id}/status`, { status })
}

// ユーザーの役割を変更
export async function updateUserRole(
  id: string, 
  role: string
): Promise<ApiResponse<User>> {
  return apiClient.patch<ApiResponse<User>>(`/users/${id}/role`, { role })
}

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

// ユーザー検索（名前、メール、部署での検索）
export async function searchUsers(query: string, limit: number = 10): Promise<ApiResponse<User[]>> {
  return apiClient.get<ApiResponse<User[]>>('/users/search', { 
    q: query, 
    limit: limit.toString() 
  })
}

// ユーザーのタスク割り当て可能性をチェック
export async function checkUserAvailability(userId: string): Promise<ApiResponse<{
  available: boolean
  current_tasks_count: number
  capacity: number
}>> {
  return apiClient.get<ApiResponse<{
    available: boolean
    current_tasks_count: number
    capacity: number
  }>>(`/users/${userId}/availability`)
}