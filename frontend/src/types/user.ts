// === User Types ===
export interface User {
  id: string;
  email: string;
  username: string; 
  role: 'user' | 'admin';
  email_verified?: boolean;
  last_login?: string;
  created_at?: string;
  updated_at?: string;
}

// === Auth Types (Token認証統一) ===
export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
}

export interface AuthResponse {
  success: boolean;
  message: string;
  data: {
    access_token: string;
    refresh_token: string;
    token_type: string;
    expires_in?: number;
    user: User;
  };
}

export interface UserResponse {
  success: boolean;
  data: User;
}