import { Group, GroupMember, GroupWithMembers } from './group';
import { ProjectTask, Task } from './task';
import { User } from './user';

// === ProjectView: Group + ProjectTaskの統合ビュー ===
export interface ProjectView {
  // プロジェクト情報 (Group情報を流用)
  group: GroupWithMembers;
  
  // プロジェクトタスク (task_type='PROJECT'のタスク)
  tasks: ProjectTask[];
  
  // プロジェクト統計 (リアルタイム計算)
  stats: ProjectStats;
  
  // プロジェクトメンバー (GroupMemberの拡張)
  members: ProjectMember[];
}

// === プロジェクト統計 ===
export interface ProjectStats {
  // タスク統計
  totalTasks: number;
  completedTasks: number;
  overdueTasks: number;
  tasksInProgress: number;
  todoTasks: number;
  
  // メンバー統計
  totalMembers: number;
  activeMembers: number;
  
  // 進捗統計
  averageProgress: number;      // 全タスクの平均進捗
  completionRate: number;       // 完了率 (completed/total * 100)
  
  // 日時統計
  estimatedCompletion?: Date;   // 推定完了日
  daysRemaining?: number;       // 残り日数
  
  // 予算統計 (将来拡張用)
  budgetUtilization?: number;   // 予算使用率
  timeUtilization?: number;     // 時間使用率
}

// === プロジェクトメンバー (GroupMemberの拡張) ===
export interface ProjectMember extends GroupMember {
  // 追加のプロジェクト固有情報
  tasksAssigned: number;        // 割り当てられたタスク数
  tasksCompleted: number;       // 完了したタスク数
  tasksInProgress: number;      // 進行中のタスク数
  completionRate: number;       // 個人完了率
  lastTaskUpdate?: Date;        // 最後にタスクを更新した日時
  
  // User情報 (GroupMember.userから展開)
  user?: User;
}

// === プロジェクト作成・更新リクエスト ===
export interface CreateProjectRequest {
  // Group作成情報
  name: string;
  description?: string;
  
  // プロジェクト固有情報
  start_date?: string;
  end_date?: string;
  budget?: number;
  tags?: string[];
  
  // Group設定 (PROJECT専用設定)
  settings?: ProjectSettings;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  start_date?: string;
  end_date?: string;
  budget?: number;
  tags?: string[];
  settings?: ProjectSettings;
}

// === プロジェクト設定 (GroupSettingsの拡張) ===
export interface ProjectSettings {
  // Group基本設定
  allow_member_invite?: boolean;
  auto_accept_invites?: boolean;
  max_members?: number;
  
  // プロジェクト固有設定
  enable_gantt_chart?: boolean;     // ガントチャート有効
  enable_task_dependency?: boolean; // タスク依存関係有効
  enable_time_tracking?: boolean;   // 時間管理有効
  enable_budget_tracking?: boolean; // 予算管理有効
  default_task_assignee?: string;   // デフォルトタスク担当者
  auto_progress_calculation?: boolean; // 進捗自動計算
}

// === プロジェクト状態管理 ===
export interface ProjectState {
  projects: ProjectView[];
  currentProject: ProjectView | null;
  isLoading: boolean;
  error: string | null;
  
  // フィルター・ソート
  filters: ProjectFilter;
  sort: ProjectSort;
}

export interface ProjectFilter {
  status?: 'active' | 'completed' | 'on-hold' | 'planning';
  member_id?: string;           // 特定メンバーが参加するプロジェクト
  progress_min?: number;        // 最小進捗率
  progress_max?: number;        // 最大進捗率
  due_date_from?: string;       // 期限開始日
  due_date_to?: string;         // 期限終了日
  search?: string;              // 名前・説明での検索
}

export interface ProjectSort {
  field: 'name' | 'created_at' | 'progress' | 'due_date' | 'member_count';
  direction: 'asc' | 'desc';
}

// === API Response Types ===
export interface ProjectViewResponse {
  success: boolean;
  data: ProjectView;
  message?: string;
}

export interface ProjectListResponse {
  success: boolean;
  data: {
    projects: ProjectView[];
    total: number;
    page: number;
    page_size: number;
  };
}

// === ユーティリティ型 ===

// プロジェクトタスク階層構造
export interface TaskHierarchy {
  task: ProjectTask;
  children: TaskHierarchy[];
  level: number;
  canStart: boolean;
}

// ガントチャート用データ
export interface GanttChartData {
  task: ProjectTask;
  startDate: Date;
  endDate: Date;
  progress: number;
  dependencies: string[];
  critical: boolean;          // クリティカルパス上のタスクか
}

// === 計算関数用のヘルパー型 ===
export interface ProjectCalculationData {
  tasks: ProjectTask[];
  members: GroupMember[];
  group: Group;
}

// === プロジェクトテンプレート (将来拡張用) ===
export interface ProjectTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  tasks: Omit<ProjectTask, 'id' | 'group_id' | 'created_at' | 'updated_at'>[];
  estimatedDuration: number;  // 推定期間(日)
  requiredMembers: number;    // 推奨メンバー数
  tags: string[];
}

// === 定数 ===
export const PROJECT_STATUSES = ['planning', 'active', 'on-hold', 'completed'] as const;
export type ProjectStatus = typeof PROJECT_STATUSES[number];

export const PROJECT_STATUS_LABELS: Record<ProjectStatus, string> = {
  planning: '企画中',
  active: '進行中', 
  'on-hold': '一時停止',
  completed: '完了'
};

// === 型ガード関数 ===
export function isProjectGroup(group: Group): boolean {
  return group.type === 'PRIVATE'; // プロジェクトはPRIVATEタイプとして扱う
}

export function hasProjectTasks(tasks: Task[]): tasks is ProjectTask[] {
  return tasks.every(task => task.task_type === 'PROJECT');
}