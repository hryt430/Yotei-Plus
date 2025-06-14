import { User } from './user';

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

// === Filter & Pagination Types ===
export interface TaskFilter {
  status?: TaskStatus;
  priority?: TaskPriority;
  category?: TaskCategory;
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
export interface TaskUpdateMessage {
  type: 'task_update';
  data: {
    task: Task;
    action: 'created' | 'updated' | 'deleted' | 'status_changed' | 'assigned';
  };
  timestamp: string;
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