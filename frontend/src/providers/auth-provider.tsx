'use client'

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { 
  loginUser, 
  registerUser, 
  logoutUser, 
  getCurrentUser, 
  checkAuthStatus 
} from '@/api/auth'
import { User, AuthState } from '@/types'
import { handleApiError } from '@/lib/utils'

interface AuthContextType extends AuthState {
  login: (email: string, password: string) => Promise<void>
  register: (username: string, email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
  clearError: () => void
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

  // 認証状態をチェック
  const checkAuth = useCallback(async () => {
    try {
      setAuthState(prev => ({ ...prev, isLoading: true }))
      
      const result = await checkAuthStatus()
      
      setAuthState(prev => ({
        ...prev,
        user: result.user || null,
        isAuthenticated: result.isAuthenticated,
        isLoading: false,
        error: null,
      }))
    } catch (error) {
      setAuthState(prev => ({
        ...prev,
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null, // 初期チェック時はエラーを表示しない
      }))
    }
  }, [])

  // 初回認証チェック
  useEffect(() => {
    checkAuth()
  }, [checkAuth])

  // ログイン
  const login = async (email: string, password: string): Promise<void> => {
    try {
      setAuthState(prev => ({ ...prev, isLoading: true, error: null }))
      
      const authResponse = await loginUser({ email, password })
      
      if (authResponse.success) {
        // ユーザー情報を取得
        const userResponse = await getCurrentUser()
        
        setAuthState(prev => ({
          ...prev,
          user: userResponse.data,
          isAuthenticated: true,
          isLoading: false,
          error: null,
        }))
        
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
      
      if (response.success) {
        // 登録後にログインを試行
        await login(email, password)
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
      
      await logoutUser()
      
      setAuthState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      })
      
      router.push('/auth/login')
    } catch (error) {
      // ログアウトエラーでも状態はリセット
      setAuthState({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      })
      
      router.push('/auth/login')
    }
  }

  // ユーザー情報の再取得
  const refreshUser = async (): Promise<void> => {
    try {
      const userResponse = await getCurrentUser()
      
      setAuthState(prev => ({
        ...prev,
        user: userResponse.data,
        error: null,
      }))
    } catch (error) {
      // ユーザー情報取得失敗時は認証状態をリセット
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

  const value: AuthContextType = {
    ...authState,
    login,
    register,
    logout,
    refreshUser,
    clearError,
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