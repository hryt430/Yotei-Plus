'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';

interface AuthState {
  user: null | {
    id: string;
    name: string;
    email: string;
  };
  isAuthenticated: boolean;
  isLoading: boolean;
}

// 初期認証状態
const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  isLoading: false,
};

export default function useAuth() {
  const [authState, setAuthState] = useState<AuthState>(initialState);
  const router = useRouter();

  // ログイン処理
  const login = async (email: string, password: string) => {
    setAuthState(prev => ({ ...prev, isLoading: true }));
    
    try {
      // APIリクエスト（実際のAPIエンドポイントに変更する）
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'ログインに失敗しました');
      }

      const data = await response.json();
      
      // 認証状態を更新
      setAuthState({
        user: data.user,
        isAuthenticated: true,
        isLoading: false,
      });
      
      // ホームへリダイレクト
      router.push('/home');
      
    } catch (error) {
      setAuthState(prev => ({ ...prev, isLoading: false }));
      throw error;
    }
  };

  // 登録処理
  const register = async (name: string, email: string, password: string) => {
    setAuthState(prev => ({ ...prev, isLoading: true }));
    
    try {
      // APIリクエスト（実際のAPIエンドポイントに変更する）
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, email, password }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'アカウント登録に失敗しました');
      }

      const data = await response.json();
      
      // 認証状態を更新
      setAuthState({
        user: data.user,
        isAuthenticated: true,
        isLoading: false,
      });
      
      // ダッシュボードへリダイレクト
      router.push('/home');
      
    } catch (error) {
      setAuthState(prev => ({ ...prev, isLoading: false }));
      throw error;
    }
  };

  // ログアウト処理
  const logout = async () => {
    setAuthState(prev => ({ ...prev, isLoading: true }));
    
    try {
      // APIリクエスト（実際のAPIエンドポイントに変更する）
      await fetch('/api/auth/logout', {
        method: 'POST',
      });
      
      // 認証状態をリセット
      setAuthState(initialState);
      
      // ログインページへリダイレクト
      router.push('/auth/login');
      
    } catch (error) {
      setAuthState(prev => ({ ...prev, isLoading: false }));
      console.error('ログアウト処理に失敗しました', error);
    }
  };

  // ユーザー情報を確認
  const checkAuth = async () => {
    setAuthState(prev => ({ ...prev, isLoading: true }));
    
    try {
      // APIリクエスト（実際のAPIエンドポイントに変更する）
      const response = await fetch('/api/auth/user');
      
      if (!response.ok) {
        // 未認証状態
        setAuthState(initialState);
        return false;
      }

      const data = await response.json();
      
      // 認証状態を更新
      setAuthState({
        user: data.user,
        isAuthenticated: true,
        isLoading: false,
      });
      
      return true;
      
    } catch (error) {
      setAuthState(initialState);
      console.error('認証確認に失敗しました', error);
      return false;
    }
  };

  return {
    ...authState,
    login,
    register,
    logout,
    checkAuth,
  };
}