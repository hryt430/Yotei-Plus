import { 
  LoginRequest, 
  RegisterRequest, 
  AuthResponse, 
  UserResponse,
  ApiResponse,
  User
} from '@/types'
import { apiClient } from '@/api/client'

// ユーザー一覧取得のパラメータ
export interface GetUsersParams {
  search?: string
  limit?: number
  offset?: number
}

// ユーザー一覧レスポンス
export interface UsersListResponse {
  success: boolean
  data: User[]
  total_count?: number
}

// ユーザー登録
export async function registerUser(data: RegisterRequest): Promise<ApiResponse<{
  user_id: string
  username: string
  email: string
}>> {
  return apiClient.post<ApiResponse<{
    user_id: string
    username: string
    email: string
  }>>('/api/auth/register', data)
}

// ログイン
export async function loginUser(data: LoginRequest): Promise<AuthResponse> {
  return apiClient.post<AuthResponse>('/api/auth/login', data)
}

// ログアウト
export async function logoutUser(): Promise<ApiResponse<null>> {
  return apiClient.post<ApiResponse<null>>('/api/auth/logout')
}

// トークンリフレッシュ
export async function refreshToken(): Promise<AuthResponse> {
  return apiClient.post<AuthResponse>('/api/auth/refresh')
}

// 現在のユーザー情報取得
export async function getCurrentUser(): Promise<UserResponse> {
  return apiClient.get<UserResponse>('/api/auth/me')
}

// ユーザー一覧を取得（タスク割り当て等で使用）
export async function getUsers(params?: GetUsersParams): Promise<UsersListResponse> {
  const queryParams: Record<string, string> = {}
  
  if (params) {
    if (params.search) queryParams.search = params.search
    if (params.limit) queryParams.limit = params.limit.toString()
    if (params.offset) queryParams.offset = params.offset.toString()
  }

  return apiClient.get<UsersListResponse>('/api/users', queryParams)
}

// 特定のユーザー情報を取得
export async function getUserById(id: string): Promise<ApiResponse<User>> {
  return apiClient.get<ApiResponse<User>>(`/api/users/${id}`)
}

// ユーザー情報を更新
export async function updateUser(id: string, data: Partial<User>): Promise<ApiResponse<User>> {
  return apiClient.put<ApiResponse<User>>(`/api/users/${id}`, data)
}

// セッション管理
export async function validateSession(): Promise<boolean> {
  try {
    await getCurrentUser()
    return true
  } catch (error) {
    return false
  }
}

// 認証状態チェック
export async function checkAuthStatus(): Promise<{
  isAuthenticated: boolean
  user?: UserResponse['data']
}> {
  try {
    const response = await getCurrentUser()
    return {
      isAuthenticated: true,
      user: response.data
    }
  } catch (error) {
    return {
      isAuthenticated: false
    }
  }
}