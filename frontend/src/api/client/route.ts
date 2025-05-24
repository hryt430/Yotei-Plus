// src/api/client.ts
// HTTPOnly Cookie対応のAPIクライアント

import { ApiResponse, ErrorResponse } from '@/types'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

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

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl
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
      credentials: 'include', // HTTPOnly Cookie対応
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

  // 基本リクエストメソッド
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
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