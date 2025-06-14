import { ApiResponse, ErrorResponse } from '@/types'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

// リクエスト情報を保持するインターフェース
interface RequestInfo {
  url: string
  method: string
  headers: Record<string, string>
  body?: any
}

// Token管理
class TokenManager {
  private static readonly ACCESS_TOKEN_KEY = 'access_token'
  private static readonly REFRESH_TOKEN_KEY = 'refresh_token'

  static getAccessToken(): string | null {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(this.ACCESS_TOKEN_KEY)
  }

  static getRefreshToken(): string | null {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(this.REFRESH_TOKEN_KEY)
  }

  static setTokens(accessToken: string, refreshToken: string): void {
    if (typeof window === 'undefined') return
    localStorage.setItem(this.ACCESS_TOKEN_KEY, accessToken)
    localStorage.setItem(this.REFRESH_TOKEN_KEY, refreshToken)
  }

  static clearTokens(): void {
    if (typeof window === 'undefined') return
    localStorage.removeItem(this.ACCESS_TOKEN_KEY)
    localStorage.removeItem(this.REFRESH_TOKEN_KEY)
  }

  static isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]))
      return payload.exp * 1000 < Date.now()
    } catch {
      return true
    }
  }
}

export class ApiError extends Error {
  constructor(
    message: string,
    public status?: number,
    public code?: string,
    public response?: any
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

// リクエストインターセプター型
type RequestInterceptor = (config: RequestInit) => RequestInit | Promise<RequestInit>
// レスポンスインターセプター型
type ResponseInterceptor = (response: Response) => Response | Promise<Response>

// APIクライアントクラス
class ApiClient {
  private baseUrl: string
  private requestInterceptors: RequestInterceptor[] = []
  private responseInterceptors: ResponseInterceptor[] = []
  private isRefreshing = false
  private refreshSubscribers: Array<(token: string) => void> = []

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl
    this.setupTokenInterceptors()
  }

  // Token認証のインターセプターを設定
  private setupTokenInterceptors() {
    // リクエストインターセプター: JWTトークンを自動設定
    this.addRequestInterceptor((config) => {
      const token = TokenManager.getAccessToken()
      if (token) {
        config.headers = {
          ...config.headers,
          'Authorization': `Bearer ${token}`
        }
      }
      return config
    })

    // レスポンスインターセプター
    this.addResponseInterceptor(async (response) => {
      if (response.status === 401 && !response.url.includes('/auth/login')) {
        
        if (!this.isRefreshing) {
          this.isRefreshing = true
          
          try {
            const newToken = await this.refreshAccessToken()
            this.isRefreshing = false
            this.onTokenRefreshed(newToken)
            
            // 401エラーの場合は、呼び出し元で再試行する必要があることを示す特別なエラーを投げる
            throw new ApiError('Token refreshed, retry needed', 401, 'TOKEN_REFRESHED', { newToken })
          } catch (refreshError) {
            this.isRefreshing = false
            this.onTokenRefreshFailed()
            throw new ApiError('認証が失効しました。再ログインしてください。', 401, 'TOKEN_EXPIRED')
          }
        } else {
          // 既にリフレッシュ中の場合は待機
          return new Promise((resolve, reject) => {
            this.refreshSubscribers.push((newToken: string) => {
              // リフレッシュ完了後に再試行が必要
              reject(new ApiError('Token refreshed, retry needed', 401, 'TOKEN_REFRESHED', { newToken }))
            })
          })
        }
      }
      
      return response
    })
  }

  // トークンリフレッシュ
  private async refreshAccessToken(): Promise<string> {
    const refreshToken = TokenManager.getRefreshToken()
    if (!refreshToken) {
      throw new Error('No refresh token available')
    }

    const response = await fetch(`${this.baseUrl}/auth/refresh-token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${refreshToken}`
      }
    })

    if (!response.ok) {
      throw new Error('Token refresh failed')
    }

    const data = await response.json()
    const { access_token, refresh_token } = data.data
    
    TokenManager.setTokens(access_token, refresh_token)
    return access_token
  }

  // トークンリフレッシュ成功時の処理
  private onTokenRefreshed(newToken: string) {
    this.refreshSubscribers.forEach(callback => callback(newToken))
    this.refreshSubscribers = []
  }

  // トークンリフレッシュ失敗時の処理
  private onTokenRefreshFailed() {
    TokenManager.clearTokens()
    this.refreshSubscribers = []
    
    // ログインページにリダイレクト
    if (typeof window !== 'undefined') {
      window.location.href = '/auth/login'
    }
  }

  // リクエストインターセプターを追加
  addRequestInterceptor(interceptor: RequestInterceptor) {
    this.requestInterceptors.push(interceptor)
  }

  // レスポンスインターセプターを追加
  addResponseInterceptor(interceptor: ResponseInterceptor) {
    this.responseInterceptors.push(interceptor)
  }

  // 基本的なfetch設定を取得
  private getDefaultOptions(): RequestInit {
    return {
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    }
  }

  // レスポンスの解析
  private async parseResponse(response: Response): Promise<any> {
    const contentType = response.headers.get('content-type')
    
    if (contentType && contentType.includes('application/json')) {
      return response.json()
    } else if (contentType && contentType.includes('text/')) {
      return response.text()
    } else {
      return response.blob()
    }
  }

  // エラーレスポンスの処理
  private createApiError(response: Response, data: any): ApiError {
    let message = `HTTP ${response.status}: ${response.statusText}`
    
    // バックエンドのエラー形式に対応
    if (data && typeof data === 'object') {
      if (data.error) {
        message = data.error
      } else if (data.message) {
        message = data.message
      } else if (data.errors && Array.isArray(data.errors)) {
        message = data.errors.join(', ')
      }
    } else if (typeof data === 'string') {
      message = data
    }

    return new ApiError(message, response.status, undefined, data)
  }

  // 基本リクエストメソッド: 自動リトライ機能付き
  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    retryCount = 0
  ): Promise<T> {
    let url = endpoint
    if (!url.startsWith('http')) {
      url = `${this.baseUrl}${endpoint}`
    }

    // デフォルトオプションとマージ
    let config: RequestInit = {
      ...this.getDefaultOptions(),
      ...options,
      headers: {
        ...this.getDefaultOptions().headers,
        ...options.headers,
      },
    }

    // リクエストインターセプターを適用
    for (const interceptor of this.requestInterceptors) {
      config = await interceptor(config)
    }

    try {
      let response = await fetch(url, config)

      // レスポンスインターセプターを適用
      for (const interceptor of this.responseInterceptors) {
        response = await interceptor(response)
      }

      // レスポンスの解析
      const data = await this.parseResponse(response)

      // エラーレスポンスの処理
      if (!response.ok) {
        throw this.createApiError(response, data)
      }

      return data as T
    } catch (error) {
      // TOKEN_REFRESHED エラーの場合は自動リトライ
      if (error instanceof ApiError && 
          error.code === 'TOKEN_REFRESHED' && 
          retryCount < 1) {
        // 新しいトークンで再試行
        return this.request<T>(endpoint, options, retryCount + 1)
      }
      
      if (error instanceof ApiError) {
        throw error
      }
      
      // ネットワークエラーなど
      if (error instanceof TypeError && error.message.includes('fetch')) {
        throw new ApiError(
          'ネットワークエラーが発生しました。接続を確認してください。',
          0,
          'NETWORK_ERROR'
        )
      }
      
      throw new ApiError(
        error instanceof Error ? error.message : '不明なエラーが発生しました',
        undefined,
        'UNKNOWN_ERROR'
      )
    }
  }

  // GET リクエスト
  async get<T>(endpoint: string, params?: Record<string, any>): Promise<T> {
    let url = endpoint
    
    if (params && Object.keys(params).length > 0) {
      const searchParams = new URLSearchParams()
      
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          if (Array.isArray(value)) {
            value.forEach(v => {
              if (v !== undefined && v !== null && v !== '') {
                searchParams.append(key, v.toString())
              }
            })
          } else {
            searchParams.append(key, value.toString())
          }
        }
      })
      
      const queryString = searchParams.toString()
      if (queryString) {
        url += (url.includes('?') ? '&' : '?') + queryString
      }
    }

    return this.request<T>(url, { method: 'GET' })
  }

  // POST リクエスト
  async post<T>(endpoint: string, data?: any): Promise<T> {
    const options: RequestInit = {
      method: 'POST',
    }

    if (data !== undefined) {
      if (data instanceof FormData) {
        options.body = data
        // FormDataの場合はContent-Typeを自動設定させる
      } else {
        options.body = JSON.stringify(data)
      }
    }

    return this.request<T>(endpoint, options)
  }

  // PUT リクエスト
  async put<T>(endpoint: string, data?: any): Promise<T> {
    const options: RequestInit = {
      method: 'PUT',
    }

    if (data !== undefined) {
      if (data instanceof FormData) {
        options.body = data
      } else {
        options.body = JSON.stringify(data)
      }
    }

    return this.request<T>(endpoint, options)
  }

  // PATCH リクエスト
  async patch<T>(endpoint: string, data?: any): Promise<T> {
    const options: RequestInit = {
      method: 'PATCH',
    }

    if (data !== undefined) {
      if (data instanceof FormData) {
        options.body = data
      } else {
        options.body = JSON.stringify(data)
      }
    }

    return this.request<T>(endpoint, options)
  }

  // DELETE リクエスト
  async delete<T>(endpoint: string, data?: any): Promise<T> {
    const options: RequestInit = {
      method: 'DELETE',
    }

    if (data !== undefined) {
      options.body = JSON.stringify(data)
    }

    return this.request<T>(endpoint, options)
  }

  // HEAD リクエスト
  async head(endpoint: string): Promise<Response> {
    const url = endpoint.startsWith('http') ? endpoint : `${this.baseUrl}${endpoint}`
    return fetch(url, {
      ...this.getDefaultOptions(),
      method: 'HEAD',
    })
  }

  // OPTIONS リクエスト
  async options(endpoint: string): Promise<Response> {
    const url = endpoint.startsWith('http') ? endpoint : `${this.baseUrl}${endpoint}`
    return fetch(url, {
      ...this.getDefaultOptions(),
      method: 'OPTIONS',
    })
  }

  // ファイルアップロード専用メソッド
  async uploadFile<T>(
    endpoint: string, 
    file: File, 
    field: string = 'file',
    additionalData?: Record<string, any>
  ): Promise<T> {
    const formData = new FormData()
    formData.append(field, file)
    
    if (additionalData) {
      Object.entries(additionalData).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          formData.append(key, value.toString())
        }
      })
    }

    return this.post<T>(endpoint, formData)
  }

  // バッチリクエスト（複数のAPIを並列実行）
  async batch<T extends Record<string, any>>(
    requests: Record<keyof T, () => Promise<any>>
  ): Promise<T> {
    const promises = Object.entries(requests).map(async ([key, requestFn]) => {
      try {
        const result = await requestFn()
        return [key, { success: true, data: result, error: null }]
      } catch (error) {
        return [key, { success: false, data: null, error }]
      }
    })

    const results = await Promise.all(promises)
    
    return Object.fromEntries(results) as T
  }

  // ヘルスチェック
  async healthCheck(): Promise<{ status: string; timestamp: string }> {
    try {
      return await this.get('/health')
    } catch (error) {
      throw new ApiError('サーバーに接続できません', 0, 'CONNECTION_ERROR')
    }
  }

  // ベースURLを動的に変更
  setBaseUrl(baseUrl: string) {
    this.baseUrl = baseUrl
  }

  // 現在のベースURLを取得
  getBaseUrl(): string {
    return this.baseUrl
  }
}

// シングルトンインスタンス
export const apiClient = new ApiClient()
export const api = apiClient // デフォルトエクスポート用エイリアス

// Token管理のエクスポート
export { TokenManager }

// 開発環境でのデバッグ用インターセプター
if (process.env.NODE_ENV === 'development') {
  apiClient.addRequestInterceptor((config) => {
    console.log('🚀 API Request:', {
      url: config,
      method: config.method,
      headers: config.headers,
      body: config.body,
    })
    return config
  })

  apiClient.addResponseInterceptor((response) => {
    console.log('📥 API Response:', {
      status: response.status,
      statusText: response.statusText,
      headers: Object.fromEntries(response.headers.entries()),
    })
    return response
  })
}

// ユーティリティ関数
export function handleApiError(error: unknown): string {
  if (error instanceof ApiError) {
    return error.message
  }
  
  if (error instanceof Error) {
    return error.message
  }
  
  return '不明なエラーが発生しました'
}

// レスポンス型チェッカー
export function isSuccessResponse<T>(
  response: any
): response is ApiResponse<T> {
  return response && typeof response === 'object' && response.success === true
}

export function isErrorResponse(
  response: any
): response is ErrorResponse {
  return response && typeof response === 'object' && typeof response.error === 'string'
}

// APIクライアントの型エクスポート
export type { ApiClient, RequestInterceptor, ResponseInterceptor }