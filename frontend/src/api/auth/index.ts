import { 
  LoginRequest, 
  RegisterRequest, 
  AuthResponse, 
  UserResponse,
  ApiResponse
} from '@/types'
import { apiClient, TokenManager } from '@/api/client'

// ユーザー登録
export async function registerUser(data: RegisterRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>('/auth/register', data)
  
  // 登録成功時にトークンを保存
  if (response.success && response.data) {
    TokenManager.setTokens(response.data.access_token, response.data.refresh_token)
  }
  
  return response
}

// ログイン
export async function loginUser(data: LoginRequest): Promise<AuthResponse> {
  const response = await apiClient.post<AuthResponse>('/auth/login', data)
  
  // ログイン成功時にトークンを保存
  if (response.success && response.data) {
    TokenManager.setTokens(response.data.access_token, response.data.refresh_token)
  }
  
  return response
}

// ログアウト
export async function logoutUser(): Promise<ApiResponse<null>> {
  try {
    const response = await apiClient.post<ApiResponse<null>>('/auth/logout')
    return response
  } finally {
    // API呼び出しの結果に関わらずトークンをクリア
    TokenManager.clearTokens()
  }
}

// トークンリフレッシュ
export async function refreshToken(): Promise<AuthResponse> {
  const refreshToken = TokenManager.getRefreshToken()
  if (!refreshToken) {
    throw new Error('リフレッシュトークンが見つかりません')
  }

  const response = await apiClient.post<AuthResponse>('/auth/refresh-token', {
    refresh_token: refreshToken
  })
  
  // トークン更新成功時に新しいトークンを保存
  if (response.success && response.data) {
    TokenManager.setTokens(response.data.access_token, response.data.refresh_token)
  }
  
  return response
}

// 現在のユーザー情報取得
export async function getCurrentUser(): Promise<UserResponse> {
  return apiClient.get<UserResponse>('/auth/me')
}

// セッション管理
export async function validateSession(): Promise<boolean> {
  try {
    const token = TokenManager.getAccessToken()
    if (!token) return false
    
    // トークンの期限チェック
    if (TokenManager.isTokenExpired(token)) {
      try {
        await refreshToken()
        return true
      } catch {
        return false
      }
    }
    
    // サーバーサイドでの検証
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
    const isValid = await validateSession()
    if (!isValid) {
      return { isAuthenticated: false }
    }
    
    const response = await getCurrentUser()
    return {
      isAuthenticated: true,
      user: response.data
    }
  } catch (error) {
    return { isAuthenticated: false }
  }
}

// トークンの有効性を確認
export function isTokenValid(): boolean {
  const token = TokenManager.getAccessToken()
  if (!token) return false
  
  return !TokenManager.isTokenExpired(token)
}

// 手動でのトークン設定（外部認証プロバイダー用など）
export function setAuthTokens(accessToken: string, refreshToken: string): void {
  TokenManager.setTokens(accessToken, refreshToken)
}

// トークンをクリア
export function clearAuthTokens(): void {
  TokenManager.clearTokens()
}

// 現在のアクセストークンを取得
export function getAccessToken(): string | null {
  return TokenManager.getAccessToken()
}

// 現在のリフレッシュトークンを取得
export function getRefreshToken(): string | null {
  return TokenManager.getRefreshToken()
}