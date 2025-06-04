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

// === Task Types ===
export type TaskStatus = 'TODO' | 'IN_PROGRESS' | 'DONE';
export type TaskPriority = 'LOW' | 'MEDIUM' | 'HIGH';
export type TaskCategory = 'WORK' | 'PERSONAL' | 'STUDY' | 'HEALTH' | 'SHOPPING' | 'OTHER';

export interface Task {
  id: string;
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  category: TaskCategory;
  assignee_id?: string;
  created_by: string;
  due_date?: string;
  is_overdue?: boolean; 
  created_at: string;
  updated_at: string;
}

export interface TaskRequest {
  title?: string;
  description?: string;
  status?: TaskStatus;
  priority?: TaskPriority;
  category?: TaskCategory;
  assignee_id?: string;
  due_date?: string;
}

// === Notification Types ===
export type NotificationType = 
  | 'APP_NOTIFICATION' 
  | 'TASK_ASSIGNED' 
  | 'TASK_COMPLETED' 
  | 'TASK_DUE_SOON' 
  | 'SYSTEM_NOTICE';

export type NotificationStatus = 'PENDING' | 'SENT' | 'READ' | 'FAILED';

export interface Notification {
  id: string;
  user_id: string;
  type: NotificationType;
  title: string;
  message: string;
  status: NotificationStatus; 
  metadata?: Record<string, string>;
  created_at: string;
  updated_at: string;
  sent_at?: string;
}

// === Statistics Types (バックエンド完全準拠) ===
export type ProgressColor = 
  | '#22c55e' // ColorDarkGreen (100%)
  | '#84cc16' // ColorGreen (80-99%)
  | '#eab308' // ColorYellow (60-79%)
  | '#f97316' // ColorOrange (40-59%)
  | '#ef4444' // ColorLightRed (20-39%)
  | '#dc2626' // ColorRed (1-19%)
  | '#9ca3af'; // ColorGray (0%)

export interface DailyStats {
  date: string;
  total_tasks: number;
  completed_tasks: number;
  in_progress_tasks: number;
  todo_tasks: number;
  overdue_tasks: number;
  completion_rate: number; // 0-100の範囲
}

export interface DailyPreview {
  date: string;
  task_count: number;
  has_overdue: boolean;
}

export interface WeeklyPreview {
  week_start: string;
  week_end: string;
  total_tasks: number;
  daily_preview: Record<string, DailyPreview>; // "Monday", "Tuesday", etc.
}

export interface WeeklyStats {
  week_start: string;
  week_end: string;
  total_tasks: number;
  completed_tasks: number;
  completion_rate: number;
  daily_stats: Record<string, DailyStats>; // "Monday", "Tuesday", etc.
}

export interface ProgressLevel {
  percentage: number;
  color: ProgressColor;
  label: string; // "完了", "優秀", "良好", "普通", "要改善", "低調", "未着手"
}

export interface DashboardStats {
  today_stats: DailyStats;
  weekly_overview: WeeklyStats;
  upcoming_week_tasks: WeeklyPreview;
  category_breakdown: Record<TaskCategory, number>;
  priority_breakdown: Record<TaskPriority, number>;
  recent_completions: Task[];
  overdue_tasks_count: number;
}

// === API Response Types ===
export interface ApiResponse<T = any> {
  success: boolean;
  message?: string;
  data: T;
}

export interface ErrorResponse {
  error: string;
  code?: string;
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

// === Task API Response Types ===
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

// === Filter & Pagination Types ===
export interface TaskFilter {
  status?: TaskStatus;
  priority?: TaskPriority;
  category?: TaskCategory; // ✅ category追加
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

// === Frontend State Types ===
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

// === Display Types ===
export interface TaskDisplayData extends Omit<Task, 'assignee_id' | 'created_by' | 'due_date' | 'created_at' | 'updated_at'> {
  assignee?: User;
  creator?: User;
  dueDate?: Date;
  createdAt: Date;
  updatedAt: Date;
  dueDateLabel?: string;
  priorityLabel: string;
  statusLabel: string;
  categoryLabel: string; 
}

// === Form Types ===
export interface TaskFormData {
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  category: TaskCategory; 
  assignee_id?: string;
  due_date?: string;
}

// === WebSocket Types ===
export interface WebSocketMessage {
  type: 'notification' | 'task_update' | 'user_update';
  data: any;
  timestamp: string;
}

export interface NotificationMessage extends WebSocketMessage {
  type: 'notification';
  data: Notification;
}

export interface TaskUpdateMessage extends WebSocketMessage {
  type: 'task_update';
  data: {
    task: Task;
    action: 'created' | 'updated' | 'deleted' | 'status_changed' | 'assigned';
  };
}

// === Utility Types ===
export interface ApiError extends Error {
  status?: number;
  code?: string;
  response?: any;
}

// === Constants ===
export const TASK_STATUSES: TaskStatus[] = ['TODO', 'IN_PROGRESS', 'DONE'];
export const TASK_PRIORITIES: TaskPriority[] = ['LOW', 'MEDIUM', 'HIGH'];
export const TASK_CATEGORIES: TaskCategory[] = ['WORK', 'PERSONAL', 'STUDY', 'HEALTH', 'SHOPPING', 'OTHER'];

export const CATEGORY_LABELS: Record<TaskCategory, string> = {
  WORK: '仕事',
  PERSONAL: '個人',
  STUDY: '学習',
  HEALTH: '健康',
  SHOPPING: '買い物',
  OTHER: 'その他'
};

export const PRIORITY_LABELS: Record<TaskPriority, string> = {
  HIGH: '高',
  MEDIUM: '中',
  LOW: '低'
};

export const STATUS_LABELS: Record<TaskStatus, string> = {
  TODO: '未着手',
  IN_PROGRESS: '進行中',
  DONE: '完了'
};