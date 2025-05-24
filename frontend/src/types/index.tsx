// src/types/index.tsx
// バックエンドの型定義と完全に整合させた型定義

export interface User {
  id: string;
  email: string;
  username: string; // バックエンドは username
  role: string;
  email_verified?: boolean;
  last_login?: string;
  created_at?: string;
  updated_at?: string;
}

export interface Task {
  id: string;
  title: string;
  description: string;
  status: 'TODO' | 'IN_PROGRESS' | 'DONE'; // バックエンドと同じ大文字
  priority: 'LOW' | 'MEDIUM' | 'HIGH'; // バックエンドと同じ大文字
  assignee_id?: string; // バックエンドは assignee_id
  created_by: string;
  due_date?: string; // ISO文字列形式
  created_at: string;
  updated_at: string;
  is_overdue?: boolean; // フロントエンド用計算フィールド
}

export interface Notification {
  id: string;
  user_id: string;
  type: 'APP_NOTIFICATION' | 'TASK_ASSIGNED' | 'TASK_COMPLETED' | 'TASK_DUE_SOON' | 'SYSTEM_NOTICE';
  title: string;
  message: string;
  status: 'PENDING' | 'SENT' | 'READ' | 'FAILED';
  metadata?: Record<string, string>;
  created_at: string;
  updated_at: string;
  sent_at?: string;
}

// API レスポンス型（バックエンドの形式に合わせる）
export interface ApiResponse<T = any> {
  success: boolean;
  message?: string;
  data: T;
}

export interface ErrorResponse {
  error: string;
}

// 認証関連の型
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
  };
}

export interface UserResponse {
  success: boolean;
  data: User
}

// タスク関連の型
export interface TaskRequest {
  title?: string;
  description?: string;
  status?: 'TODO' | 'IN_PROGRESS' | 'DONE';
  priority?: 'LOW' | 'MEDIUM' | 'HIGH';
  assignee_id?: string;
  due_date?: string;
}

export interface TaskListResponse {
  success: boolean;
  data: {
    tasks: Task[];
    total_count: number;
    page: number;
    page_size: number;
  };
}

export interface TaskResponse {
  success: boolean;
  message?: string;
  data: Task;
}

// フィルター・ソート関連
export interface TaskFilter {
  status?: 'TODO' | 'IN_PROGRESS' | 'DONE';
  priority?: 'LOW' | 'MEDIUM' | 'HIGH';
  assignee_id?: string;
  created_by?: string;
  due_date_from?: string;
  due_date_to?: string;
  search?: string;
}

export interface Pagination {
  page: number;
  page_size: number;
}

export interface SortOptions {
  field: string;
  direction: 'ASC' | 'DESC';
}

// フロントエンド用の状態管理型
export interface TasksState {
  tasks: Task[];
  isLoading: boolean;
  error: string | null;
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
  filters: TaskFilter;
  sort: {
    field: string;
    direction: 'asc' | 'desc';
  };
}

export interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  isLoading: boolean;
  error: string | null;
}

// ユーティリティ型
export type TaskStatus = Task['status'];
export type TaskPriority = Task['priority'];
export type NotificationType = Notification['type'];
export type NotificationStatus = Notification['status'];

// 表示用の型変換
export interface TaskDisplayData extends Omit<Task, 'assignee_id' | 'created_by' | 'due_date' | 'created_at' | 'updated_at'> {
  assignee?: User;
  creator?: User;
  dueDate?: Date;
  createdAt: Date;
  updatedAt: Date;
  dueDateLabel?: string;
  priorityLabel: string;
  statusLabel: string;
}

// フォーム用の型
export interface TaskFormData {
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  assignee_id?: string;
  due_date?: string;
}

// API エラー型
export interface ApiError extends Error {
  status?: number;
  code?: string;
}