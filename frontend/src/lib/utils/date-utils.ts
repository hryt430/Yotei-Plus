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

export function getDaysUntil(dateString: string): number {
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

export function getStartOfWeek(date: Date): Date {
  const day = date.getDay();
  const diff = date.getDate() - day + (day === 0 ? -6 : 1); // 日曜日から月曜日にする調整
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

export function addMonths(date: Date, months: number): Date {
  const result = new Date(date);
  result.setMonth(result.getMonth() + months);
  return result;
}

export function getWeekdayNames(): string[] {
  return ['月', '火', '水', '木', '金', '土', '日'];
}

export function getMonthName(month: number): string {
  const date = new Date();
  date.setMonth(month);
  return new Intl.DateTimeFormat('ja-JP', { month: 'long' }).format(date);
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
  return Math.round(Math.abs((end.getTime() - start.getTime()) / oneDay)) + 1;
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