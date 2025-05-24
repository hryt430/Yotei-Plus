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

// ユーザー一覧取得（タスク割り当て用）
export async function getUsers(): Promise<ApiResponse<{
  id: string
  username: string
  email: string
  role: string
}[]>> {
  return apiClient.get<ApiResponse<{
    id: string
    username: string
    email: string
    role: string
  }[]>>('/api/users')
}

// パスワード変更（将来の拡張用）
export async function changePassword(
  currentPassword: string,
  newPassword: string
): Promise<ApiResponse<null>> {
  return apiClient.put<ApiResponse<null>>('/api/auth/password', {
    current_password: currentPassword,
    new_password: newPassword
  })
}

// プロフィール更新（将来の拡張用）
export async function updateProfile(data: {
  username?: string
  email?: string
}): Promise<UserResponse> {
  return apiClient.put<UserResponse>('/api/auth/profile', data)
}

// メール認証（将来の拡張用）
export async function verifyEmail(token: string): Promise<ApiResponse<null>> {
  return apiClient.post<ApiResponse<null>>('/api/auth/verify-email', {
    token
  })
}

// パスワードリセット要求（将来の拡張用）
export async function requestPasswordReset(email: string): Promise<ApiResponse<null>> {
  return apiClient.post<ApiResponse<null>>('/api/auth/reset-password', {
    email
  })
}

// パスワードリセット実行（将来の拡張用）
export async function resetPassword(
  token: string,
  newPassword: string
): Promise<ApiResponse<null>> {
  return apiClient.post<ApiResponse<null>>('/api/auth/reset-password/confirm', {
    token,
    new_password: newPassword
  })
}

// アカウント削除（将来の拡張用）
export async function deleteAccount(password: string): Promise<ApiResponse<null>> {
  return apiClient.delete<ApiResponse<null>>('/api/auth/account', {
    body: JSON.stringify({ password })
  })
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