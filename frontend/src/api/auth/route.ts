// src/api/auth.ts
// 認証関連のAPI関数

import { 
  LoginRequest, 
  RegisterRequest, 
  AuthResponse, 
  UserResponse,
  ApiResponse 
} from '@/types'
import { apiClient } from '@/api/client/route'

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