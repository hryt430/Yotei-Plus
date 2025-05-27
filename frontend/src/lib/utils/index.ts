import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"
import { Task, TaskDisplayData, User } from "@/types"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// 日付ユーティリティ
export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('ja-JP', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(date);
}

export function formatDateTime(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('ja-JP', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

export function isToday(dateString: string): boolean {
  const date = new Date(dateString);
  const today = new Date();
  
  return date.getDate() === today.getDate() &&
    date.getMonth() === today.getMonth() &&
    date.getFullYear() === today.getFullYear();
}

export function isTomorrow(dateString: string): boolean {
  const date = new Date(dateString);
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  
  return date.getDate() === tomorrow.getDate() &&
    date.getMonth() === tomorrow.getMonth() &&
    date.getFullYear() === tomorrow.getFullYear();
}

export function isOverdue(dateString: string): boolean {
  const date = new Date(dateString);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  
  return date < today;
}

export function getRelativeDateLabel(dateString: string): string {
  if (!dateString) return '';
  
  if (isToday(dateString)) {
    return '今日';
  } else if (isTomorrow(dateString)) {
    return '明日';
  } else if (isOverdue(dateString)) {
    return '期限切れ';
  } else {
    const days = getDaysUntil(dateString);
    if (days <= 7) {
      return `${days}日後`;
    } else {
      return formatDate(dateString);
    }
  }
}

export function getDaysUntil(dateString: string): number {
  const date = new Date(dateString);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  
  const diffTime = date.getTime() - today.getTime();
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  
  return diffDays;
}

// タスク関連ユーティリティ
export function getStatusLabel(status: Task['status']): string {
  switch (status) {
    case 'TODO':
      return '未着手';
    case 'IN_PROGRESS':
      return '進行中';
    case 'DONE':
      return '完了';
    default:
      return '不明';
  }
}

export function getPriorityLabel(priority: Task['priority']): string {
  switch (priority) {
    case 'LOW':
      return '低';
    case 'MEDIUM':
      return '中';
    case 'HIGH':
      return '高';
    default:
      return '不明';
  }
}

export function getStatusColor(status: Task['status']): string {
  switch (status) {
    case 'TODO':
      return 'bg-gray-100 text-gray-800 border-gray-200';
    case 'IN_PROGRESS':
      return 'bg-blue-100 text-blue-800 border-blue-200';
    case 'DONE':
      return 'bg-green-100 text-green-800 border-green-200';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200';
  }
}

export function getPriorityColor(priority: Task['priority']): string {
  switch (priority) {
    case 'LOW':
      return 'bg-green-100 text-green-800 border-green-200';
    case 'MEDIUM':
      return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    case 'HIGH':
      return 'bg-red-100 text-red-800 border-red-200';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200';
  }
}

// タスクをフロントエンド表示用に変換
export function transformTaskForDisplay(task: Task, users: User[] = []): TaskDisplayData {
  const assignee = task.assignee_id 
    ? users.find(user => user.id === task.assignee_id)
    : undefined;
  
  const creator = users.find(user => user.id === task.created_by);
  
  return {
    ...task,
    assignee,
    creator,
    dueDate: task.due_date ? new Date(task.due_date) : undefined,
    createdAt: new Date(task.created_at),
    updatedAt: new Date(task.updated_at),
    dueDateLabel: task.due_date ? getRelativeDateLabel(task.due_date) : undefined,
    priorityLabel: getPriorityLabel(task.priority),
    statusLabel: getStatusLabel(task.status),
  };
}

// API エラーハンドリング
export function handleApiError(error: any): string {
  if (error?.response?.data?.error) {
    return error.response.data.error;
  }
  if (error?.message) {
    return error.message;
  }
  return 'エラーが発生しました';
}

// URL パラメータの構築
export function buildQueryParams(params: Record<string, any>): string {
  const searchParams = new URLSearchParams();
  
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      if (Array.isArray(value)) {
        value.forEach(v => searchParams.append(key, v.toString()));
      } else {
        searchParams.append(key, value.toString());
      }
    }
  });
  
  return searchParams.toString();
}

// デバウンス関数
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout;
  
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

// ローカルストレージ安全アクセス
export function getLocalStorage(key: string): string | null {
  if (typeof window === 'undefined') return null;
  
  try {
    return localStorage.getItem(key);
  } catch (error) {
    console.warn('localStorage access failed:', error);
    return null;
  }
}

export function setLocalStorage(key: string, value: string): void {
  if (typeof window === 'undefined') return;
  
  try {
    localStorage.setItem(key, value);
  } catch (error) {
    console.warn('localStorage write failed:', error);
  }
}

export function removeLocalStorage(key: string): void {
  if (typeof window === 'undefined') return;
  
  try {
    localStorage.removeItem(key);
  } catch (error) {
    console.warn('localStorage remove failed:', error);
  }
}

// バリデーション
export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

export function isValidPassword(password: string): boolean {
  return password.length >= 8;
}

export function isValidUsername(username: string): boolean {
  return username.length >= 3 && username.length <= 30;
}