import { Task } from '@/types';

export function formatDate(dateString: string): string {
  if (!dateString) return '';
  
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return '';
  
  return new Intl.DateTimeFormat('ja-JP', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(date);
}

export function formatDateTime(dateString: string): string {
  if (!dateString) return '';
  
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return '';
  
  return new Intl.DateTimeFormat('ja-JP', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

export function formatTime(dateString: string): string {
  if (!dateString) return '';
  
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return '';
  
  return new Intl.DateTimeFormat('ja-JP', {
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

export function isToday(dateString: string): boolean {
  if (!dateString) return false;
  
  const date = new Date(dateString);
  const today = new Date();
  
  return date.getDate() === today.getDate() &&
    date.getMonth() === today.getMonth() &&
    date.getFullYear() === today.getFullYear();
}

export function isTomorrow(dateString: string): boolean {
  if (!dateString) return false;
  
  const date = new Date(dateString);
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  
  return date.getDate() === tomorrow.getDate() &&
    date.getMonth() === tomorrow.getMonth() &&
    date.getFullYear() === tomorrow.getFullYear();
}

export function isYesterday(dateString: string): boolean {
  if (!dateString) return false;
  
  const date = new Date(dateString);
  const yesterday = new Date();
  yesterday.setDate(yesterday.getDate() - 1);
  
  return date.getDate() === yesterday.getDate() &&
    date.getMonth() === yesterday.getMonth() &&
    date.getFullYear() === yesterday.getFullYear();
}

export function isOverdue(dateString: string): boolean {
  if (!dateString) return false;
  
  const date = new Date(dateString);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  
  return date < today;
}

export function getDaysUntil(dateString: string): number {
  if (!dateString) return 0;
  
  const date = new Date(dateString);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  
  const diffTime = date.getTime() - today.getTime();
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  
  return diffDays;
}

export function getRelativeDateLabel(dateString: string): string {
  if (!dateString) return '';
  
  if (isToday(dateString)) {
    return '今日';
  } else if (isTomorrow(dateString)) {
    return '明日';
  } else if (isYesterday(dateString)) {
    return '昨日';
  } else if (isOverdue(dateString)) {
    const daysOverdue = Math.abs(getDaysUntil(dateString));
    return `${daysOverdue}日前`;
  } else {
    const days = getDaysUntil(dateString);
    if (days <= 7) {
      return `${days}日後`;
    } else if (days <= 30) {
      const weeks = Math.ceil(days / 7);
      return `${weeks}週間後`;
    } else {
      return formatDate(dateString);
    }
  }
}

export function getStartOfWeek(date: Date): Date {
  const day = date.getDay();
  const diff = date.getDate() - day + (day === 0 ? -6 : 1); // 月曜日を週の開始とする
  const start = new Date(date);
  start.setDate(diff);
  start.setHours(0, 0, 0, 0);
  return start;
}

export function getEndOfWeek(date: Date): Date {
  const start = getStartOfWeek(date);
  const end = new Date(start);
  end.setDate(start.getDate() + 6);
  end.setHours(23, 59, 59, 999);
  return end;
}

export function getStartOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

export function getEndOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth() + 1, 0, 23, 59, 59, 999);
}

export function getMonthWeeks(year: number, month: number): Date[][] {
  const weeks: Date[][] = [];
  const firstDay = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0);
  
  // 月の最初の週の開始日を取得
  const startOfFirstWeek = getStartOfWeek(firstDay);
  
  // 月の最後の週の終了日を取得
  const endOfLastWeek = getEndOfWeek(lastDay);
  
  // 各週をループ
  let currentDate = new Date(startOfFirstWeek);
  
  while (currentDate <= endOfLastWeek) {
    const week: Date[] = [];
    
    // 週の各日をループ
    for (let i = 0; i < 7; i++) {
      week.push(new Date(currentDate));
      currentDate.setDate(currentDate.getDate() + 1);
    }
    
    weeks.push(week);
  }
  
  return weeks;
}

export function isSameDay(date1: Date, date2: Date): boolean {
  return date1.getDate() === date2.getDate() &&
         date1.getMonth() === date2.getMonth() &&
         date1.getFullYear() === date2.getFullYear();
}

export function isSameMonth(date1: Date, date2: Date): boolean {
  return date1.getMonth() === date2.getMonth() &&
         date1.getFullYear() === date2.getFullYear();
}

export function isSameYear(date1: Date, date2: Date): boolean {
  return date1.getFullYear() === date2.getFullYear();
}

export function addDays(date: Date, days: number): Date {
  const result = new Date(date);
  result.setDate(result.getDate() + days);
  return result;
}

export function addWeeks(date: Date, weeks: number): Date {
  return addDays(date, weeks * 7);
}

export function addMonths(date: Date, months: number): Date {
  const result = new Date(date);
  result.setMonth(result.getMonth() + months);
  return result;
}

export function addYears(date: Date, years: number): Date {
  const result = new Date(date);
  result.setFullYear(result.getFullYear() + years);
  return result;
}

export function getWeekdayNames(short: boolean = false): string[] {
  if (short) {
    return ['月', '火', '水', '木', '金', '土', '日'];
  }
  return ['月曜日', '火曜日', '水曜日', '木曜日', '金曜日', '土曜日', '日曜日'];
}

export function getMonthName(month: number, short: boolean = false): string {
  const date = new Date();
  date.setMonth(month);
  return new Intl.DateTimeFormat('ja-JP', { 
    month: short ? 'short' : 'long' 
  }).format(date);
}

export function getMonthNames(short: boolean = false): string[] {
  return Array.from({ length: 12 }, (_, i) => getMonthName(i, short));
}

export function getDateTimeForAPI(date: Date): string {
  return date.toISOString();
}

export function parseAPIDateTime(dateString: string): Date {
  return new Date(dateString);
}

export function getDaysBetween(startDate: Date, endDate: Date): number {
  const oneDay = 24 * 60 * 60 * 1000; // ミリ秒単位での1日
  const start = new Date(startDate);
  const end = new Date(endDate);
  
  // 時間部分をリセット
  start.setHours(0, 0, 0, 0);
  end.setHours(0, 0, 0, 0);
  
  // 日数を計算
  return Math.round(Math.abs((end.getTime() - start.getTime()) / oneDay));
}

export function getDateRangeArray(startDate: Date, endDate: Date): Date[] {
  const dateArray: Date[] = [];
  const currentDate = new Date(startDate);
  
  while (currentDate <= endDate) {
    dateArray.push(new Date(currentDate));
    currentDate.setDate(currentDate.getDate() + 1);
  }
  
  return dateArray;
}

export function getTimeOptions(intervalMinutes: number = 30): string[] {
  const options: string[] = [];
  const minutes = 24 * 60;
  
  for (let i = 0; i < minutes; i += intervalMinutes) {
    const hours = Math.floor(i / 60);
    const mins = i % 60;
    options.push(`${hours.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}`);
  }
  
  return options;
}

export function getQuarter(date: Date): number {
  return Math.floor(date.getMonth() / 3) + 1;
}

export function getStartOfQuarter(date: Date): Date {
  const quarter = getQuarter(date);
  const startMonth = (quarter - 1) * 3;
  return new Date(date.getFullYear(), startMonth, 1);
}

export function getEndOfQuarter(date: Date): Date {
  const quarter = getQuarter(date);
  const endMonth = quarter * 3 - 1;
  const endDate = new Date(date.getFullYear(), endMonth + 1, 0);
  endDate.setHours(23, 59, 59, 999);
  return endDate;
}

// タスク関連の日付ユーティリティ
export function isTaskOverdue(task: Task): boolean {
  return task.due_date ? isOverdue(task.due_date) && task.status !== 'DONE' : false;
}

export function getTaskDueDateLabel(task: Task): string {
  return task.due_date ? getRelativeDateLabel(task.due_date) : '';
}

export function sortTasksByDueDate(tasks: Task[], direction: 'asc' | 'desc' = 'asc'): Task[] {
  return [...tasks].sort((a, b) => {
    // due_dateがないタスクは最後に配置
    if (!a.due_date && !b.due_date) return 0;
    if (!a.due_date) return 1;
    if (!b.due_date) return -1;
    
    const dateA = new Date(a.due_date);
    const dateB = new Date(b.due_date);
    
    return direction === 'asc' 
      ? dateA.getTime() - dateB.getTime()
      : dateB.getTime() - dateA.getTime();
  });
}

export function filterTasksByDateRange(
  tasks: Task[], 
  startDate?: string, 
  endDate?: string
): Task[] {
  return tasks.filter(task => {
    if (!task.due_date) return false;
    
    const taskDate = new Date(task.due_date);
    
    if (startDate && taskDate < new Date(startDate)) return false;
    if (endDate && taskDate > new Date(endDate)) return false;
    
    return true;
  });
}

export function getTasksForDate(tasks: Task[], date: Date): Task[] {
  const targetDateStr = date.toISOString().split('T')[0];
  
  return tasks.filter(task => {
    if (!task.due_date) return false;
    
    const taskDateStr = new Date(task.due_date).toISOString().split('T')[0];
    return taskDateStr === targetDateStr;
  });
}

// 日付の妥当性チェック
export function isValidDate(dateString: string): boolean {
  if (!dateString) return false;
  
  const date = new Date(dateString);
  return !isNaN(date.getTime());
}

// 日付を YYYY-MM-DD 形式にフォーマット
export function formatDateForInput(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  if (isNaN(d.getTime())) return '';
  
  return d.toISOString().split('T')[0];
}

// 現在時刻をISO文字列で取得
export function now(): string {
  return new Date().toISOString();
}

// 今日の日付をYYYY-MM-DD形式で取得
export function today(): string {
  return formatDateForInput(new Date());
}

// 明日の日付をYYYY-MM-DD形式で取得
export function tomorrow(): string {
  return formatDateForInput(addDays(new Date(), 1));
}