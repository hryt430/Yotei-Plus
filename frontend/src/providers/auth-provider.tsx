'use client'

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { 
  loginUser, 
  registerUser, 
  logoutUser, 
  getCurrentUser, 
  checkAuthStatus,
  isTokenValid,
  clearAuthTokens
} from '@/api/auth'
import { User, AuthState } from '@/types'
import { handleApiError } from '@/lib/utils'
import { TokenManager } from '@/api/client'

interface AuthContextType extends AuthState {
  login: (email: string, password: string) => Promise<void>
  register: (username: string, email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
  clearError: () => void
  checkTokenValidity: () => boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    isAuthenticated: false,
    isLoading: true,
    error: null,
  })
  
  const router = useRouter()

  // トークンの有効性をチェック
  const checkTokenValidity = useCallback((): boolean => {
    return isTokenValid()
  }, [])

  // 認証状態をチェック
  const checkAuth = useCallback(async () => {
    try {
      setAuthState(prev => ({ ...prev, isLoading: true }))
      
      // まずローカルのトークンをチェック
      if (!checkTokenValidity()) {
        setAuthState({
          user: null,
          isAuthenticated: false,
          isLoading: false,
          error: null,
        })
        return
      }

      // サーバーサイドでの認証状態確認
      const result = await checkAuthStatus()
      
      setAuthState(prev => ({
        ...prev,
        user: result.user || null,
        isAuthenticated: result.isAuthenticated,
        isLoading: false,
        error: null,
      }))
    } catch (error) {
      console.warn('Auth check failed:', error)
      
      // トークンが無効な場合はクリア
      clearAuthTokens()
      
      setAuthState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null, // 初期チェック時はエラーを表示しない
      })
    }
  }, [checkTokenValidity])

  // 初回認証チェック
  useEffect(() => {
    checkAuth()
  }, [checkAuth])

  // トークンの有効期限を定期的にチェック
  useEffect(() => {
    if (!authState.isAuthenticated) return

    const interval = setInterval(() => {
      if (!checkTokenValidity()) {
        console.log('Token expired, logging out')
        logout()
      }
    }, 60000) // 1分ごとにチェック

    return () => clearInterval(interval)
  }, [authState.isAuthenticated, checkTokenValidity])

  // ログイン
  const login = async (email: string, password: string): Promise<void> => {
    try {
      setAuthState(prev => ({ ...prev, isLoading: true, error: null }))
      
      const authResponse = await loginUser({ email, password })
      
      if (authResponse.success && authResponse.data) {
        // レスポンスにユーザー情報が含まれている場合
        if (authResponse.data.user) {
          setAuthState(prev => ({
            ...prev,
            user: authResponse.data.user,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          }))
        } else {
          // ユーザー情報を別途取得
          const userResponse = await getCurrentUser()
          setAuthState(prev => ({
            ...prev,
            user: userResponse.data,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          }))
        }
        
        // ダッシュボードにリダイレクト
        router.push('/dashboard')
      }
    } catch (error) {
      const errorMessage = handleApiError(error)
      setAuthState(prev => ({
        ...prev,
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: errorMessage,
      }))
      throw error
    }
  }

  // 登録
  const register = async (username: string, email: string, password: string): Promise<void> => {
    try {
      setAuthState(prev => ({ ...prev, isLoading: true, error: null }))
      
      const response = await registerUser({ username, email, password })
      
      if (response.success && response.data) {
        // 登録レスポンスにユーザー情報が含まれている場合
        if (response.data.user) {
          setAuthState(prev => ({
            ...prev,
            user: response.data.user,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          }))
        } else {
          // ユーザー情報を別途取得
          const userResponse = await getCurrentUser()
          setAuthState(prev => ({
            ...prev,
            user: userResponse.data,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          }))
        }
        
        // ダッシュボードにリダイレクト
        router.push('/dashboard')
      }
    } catch (error) {
      const errorMessage = handleApiError(error)
      setAuthState(prev => ({
        ...prev,
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: errorMessage,
      }))
      throw error
    }
  }

  // ログアウト
  const logout = async (): Promise<void> => {
    try {
      setAuthState(prev => ({ ...prev, isLoading: true }))
      
      // サーバーサイドでのログアウト処理
      await logoutUser()
    } catch (error) {
      console.warn('Server logout failed:', error)
      // サーバーサイドのログアウトに失敗してもクライアントサイドはクリア
    } finally {
      // 認証状態をリセット
      setAuthState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      })
      
      // ログインページへリダイレクト
      router.push('/auth/login')
    }
  }

  // ユーザー情報の再取得
  const refreshUser = async (): Promise<void> => {
    try {
      if (!checkTokenValidity()) {
        throw new Error('Invalid token')
      }

      const userResponse = await getCurrentUser()
      
      setAuthState(prev => ({
        ...prev,
        user: userResponse.data,
        error: null,
      }))
    } catch (error) {
      console.warn('User refresh failed:', error)
      // ユーザー情報取得失敗時は認証状態をリセット
      clearAuthTokens()
      setAuthState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      })
    }
  }

  // エラーをクリア
  const clearError = (): void => {
    setAuthState(prev => ({ ...prev, error: null }))
  }

  // ページ離脱時の処理
  useEffect(() => {
    const handleBeforeUnload = () => {
      // 必要に応じて状態保存処理を追加
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    
    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload)
    }
  }, [])

  // ブラウザフォーカス時の認証チェック
  useEffect(() => {
    const handleFocus = () => {
      if (authState.isAuthenticated && !checkTokenValidity()) {
        logout()
      }
    }

    window.addEventListener('focus', handleFocus)
    
    return () => {
      window.removeEventListener('focus', handleFocus)
    }
  }, [authState.isAuthenticated, checkTokenValidity])

  const value: AuthContextType = {
    ...authState,
    login,
    register,
    logout,
    refreshUser,
    clearError,
    checkTokenValidity,
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

// 認証が必要なページ用のHOC
export function withAuth<P extends object>(
  WrappedComponent: React.ComponentType<P>
) {
  const WithAuthComponent = (props: P) => {
    const { isAuthenticated, isLoading } = useAuth()
    const router = useRouter()

    useEffect(() => {
      if (!isLoading && !isAuthenticated) {
        router.push('/auth/login')
      }
    }, [isAuthenticated, isLoading, router])

    if (isLoading) {
      return (
        <div className="flex items-center justify-center min-h-screen">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-gray-900"></div>
        </div>
      )
    }

    if (!isAuthenticated) {
      return null
    }

    return <WrappedComponent {...props} />
  }

  WithAuthComponent.displayName = `withAuth(${WrappedComponent.displayName || WrappedComponent.name})`
  
  return WithAuthComponent
}

// 管理者のみアクセス可能なページ用のHOC
export function withAdminAuth<P extends object>(
  WrappedComponent: React.ComponentType<P>
) {
  const WithAdminAuthComponent = (props: P) => {
    const { user, isAuthenticated, isLoading } = useAuth()
    const router = useRouter()

    useEffect(() => {
      if (!isLoading) {
        if (!isAuthenticated) {
          router.push('/auth/login')
        } else if (user?.role !== 'admin') {
          router.push('/dashboard') // 管理者でない場合はダッシュボードへ
        }
      }
    }, [isAuthenticated, isLoading, user, router])

    if (isLoading) {
      return (
        <div className="flex items-center justify-center min-h-screen">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-gray-900"></div>
        </div>
      )
    }

    if (!isAuthenticated || user?.role !== 'admin') {
      return null
    }

    return <WrappedComponent {...props} />
  }

  WithAdminAuthComponent.displayName = `withAdminAuth(${WrappedComponent.displayName || WrappedComponent.name})`
  
  return WithAdminAuthComponent
}