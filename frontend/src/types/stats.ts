import { TaskCategory, TaskPriority, Task } from './task';

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