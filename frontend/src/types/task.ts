import { User } from './user';

// === Task Types ===
export type TaskStatus = 'TODO' | 'IN_PROGRESS' | 'DONE';
export type TaskPriority = 'LOW' | 'MEDIUM' | 'HIGH';
export type TaskCategory = 'WORK' | 'PERSONAL' | 'STUDY' | 'HEALTH' | 'SHOPPING' | 'OTHER';
export type TaskType = 'PERSONAL' | 'PROJECT';

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
  task_type: TaskType;
  
  // プロジェクトタスク専用フィールド
  group_id?: string;        // プロジェクト(グループ)ID
  start_date?: string;      // 開始日
  end_date?: string;        // 終了日 (due_dateとは別)
  progress?: number;        // 進捗率 (0-100)
  parent_task_id?: string;  // 親タスクID (サブタスク用)
  level?: number;           // 階層レベル (0=ルート, 1=サブタスク...)
  sort_order?: number;      // 表示順序
  dependencies?: string[];  // 依存タスクID配列
  
  // 計算フィールド
  can_start?: boolean;      // 依存関係が解決済みで開始可能か
  has_subtasks?: boolean;   // サブタスクを持つか
  version?: number;         // 楽観的ロック用
}

export interface ProjectTask extends Task {
  task_type: 'PROJECT';
  group_id: string;         // 必須
}

export interface PersonalTask extends Task {
  task_type: 'PERSONAL';
  group_id?: never;         // 個人タスクにはgroup_idは不要
  start_date?: never;
  end_date?: never;
  progress?: never;
  parent_task_id?: never;
  level?: never;
  sort_order?: never;
  dependencies?: never;
  can_start?: never;
  has_subtasks?: never;
}

export interface TaskRequest {
  title?: string;
  description?: string;
  status?: TaskStatus;
  priority?: TaskPriority;
  category?: TaskCategory;
  assignee_id?: string;
  due_date?: string;
  
  task_type?: TaskType;
  group_id?: string;
  start_date?: string;
  end_date?: string;
  progress?: number;
  parent_task_id?: string;
  level?: number;
  sort_order?: number;
  dependencies?: string[];
}

export interface ProjectTaskRequest {
  title: string;
  description: string;
  group_id: string;         // 必須
  task_type: 'PROJECT';     // 固定
  status?: TaskStatus;
  priority?: TaskPriority;
  assignee_id?: string;
  start_date?: string;
  end_date?: string;
  due_date?: string;
  parent_task_id?: string;
  level?: number;
  dependencies?: string[];
}

// === Filter拡張 ===
export interface TaskFilter {
  status?: TaskStatus;
  priority?: TaskPriority;
  category?: TaskCategory;
  assignee_id?: string;
  created_by?: string;
  due_date_from?: string;
  due_date_to?: string;
  search?: string;
  
  task_type?: TaskType;
  group_id?: string;
  parent_task_id?: string;
  level?: number;
  progress_min?: number;
  progress_max?: number;
  is_overdue?: boolean;
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

export interface ProjectTaskFormData extends TaskFormData {
  task_type: 'PROJECT';
  group_id: string;
  start_date?: string;
  end_date?: string;
  parent_task_id?: string;
  level?: number;
  dependencies?: string[];
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

// === 型ガード関数 ===
export function isProjectTask(task: Task): task is ProjectTask {
  return task.task_type === 'PROJECT';
}

export function isPersonalTask(task: Task): task is PersonalTask {
  return task.task_type === 'PERSONAL';
}

// === 定数とラベル ===
export const TASK_TYPES: TaskType[] = ['PERSONAL', 'PROJECT'];
export const TASK_STATUSES: TaskStatus[] = ['TODO', 'IN_PROGRESS', 'DONE'];
export const TASK_PRIORITIES: TaskPriority[] = ['LOW', 'MEDIUM', 'HIGH'];
export const TASK_CATEGORIES: TaskCategory[] = ['WORK', 'PERSONAL', 'STUDY', 'HEALTH', 'SHOPPING', 'OTHER'];

export const TASK_TYPE_LABELS: Record<TaskType, string> = {
  PERSONAL: '個人タスク',
  PROJECT: 'プロジェクトタスク'
};

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