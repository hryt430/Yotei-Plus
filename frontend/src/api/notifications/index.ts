import { ApiResponse, Notification, NotificationType, NotificationStatus } from '@/types'
import { apiClient } from '@/api/client'

// 通知作成の入力データ
export interface CreateNotificationInput {
  user_id: string
  type: NotificationType
  title: string
  message: string
  metadata?: Record<string, string>
  channels: ('app' | 'line')[] // 送信チャネル指定
}

// 通知一覧取得の入力データ
export interface GetNotificationsParams {
  limit?: number
  offset?: number
  status?: NotificationStatus // ✅ is_readではなくstatusで統一
  unread_only?: boolean // ✅ 未読のみ取得のフラグ
  type?: NotificationType
  created_after?: string
  created_before?: string
  sort_field?: 'created_at' | 'updated_at' | 'sent_at'
  sort_direction?: 'ASC' | 'DESC'
}

// 通知一覧レスポンス
export interface NotificationsListResponse {
  success: boolean
  data: {
    notifications: Notification[]
    total_count: number
    unread_count: number
    page?: number
    page_size?: number
  }
}

// === Basic Notification Operations ===

// 新しい通知を作成
export async function createNotification(data: CreateNotificationInput): Promise<ApiResponse<Notification>> {
  return apiClient.post<ApiResponse<Notification>>('/notifications', data)
}

// 特定の通知を取得
export async function getNotification(id: string): Promise<ApiResponse<Notification>> {
  return apiClient.get<ApiResponse<Notification>>(`/notifications/${id}`)
}

// ユーザーの通知一覧を取得
export async function getUserNotifications(
  userId: string, 
  params?: GetNotificationsParams
): Promise<NotificationsListResponse> {
  const queryParams: Record<string, string> = {}
  
  if (params) {
    if (params.limit) queryParams.limit = params.limit.toString()
    if (params.offset) queryParams.offset = params.offset.toString()
    if (params.status) queryParams.status = params.status
    if (params.unread_only) queryParams.unread_only = 'true'
    if (params.type) queryParams.type = params.type
    if (params.created_after) queryParams.created_after = params.created_after
    if (params.created_before) queryParams.created_before = params.created_before
    if (params.sort_field) queryParams.sort_field = params.sort_field
    if (params.sort_direction) queryParams.sort_direction = params.sort_direction
  }
  
  return apiClient.get<NotificationsListResponse>(`/notifications/user/${userId}`, queryParams)
}

// 現在のユーザーの通知一覧を取得
export async function getMyNotifications(
  params?: GetNotificationsParams
): Promise<NotificationsListResponse> {
  const queryParams: Record<string, string> = {}
  
  if (params) {
    if (params.limit) queryParams.limit = params.limit.toString()
    if (params.offset) queryParams.offset = params.offset.toString()
    if (params.status) queryParams.status = params.status
    if (params.unread_only) queryParams.unread_only = 'true'
    if (params.type) queryParams.type = params.type
    if (params.created_after) queryParams.created_after = params.created_after
    if (params.created_before) queryParams.created_before = params.created_before
    if (params.sort_field) queryParams.sort_field = params.sort_field
    if (params.sort_direction) queryParams.sort_direction = params.sort_direction
  }
  
  return apiClient.get<NotificationsListResponse>('/notifications/me', queryParams)
}

// === Notification Actions ===

// 通知を送信
export async function sendNotification(id: string): Promise<ApiResponse<{ message: string }>> {
  return apiClient.post<ApiResponse<{ message: string }>>(`/notifications/${id}/send`)
}

// 通知を既読にする（statusを'READ'に変更）
export async function markNotificationAsRead(id: string): Promise<ApiResponse<Notification>> {
  return apiClient.put<ApiResponse<Notification>>(`/notifications/${id}/read`)
}

// 複数の通知を既読にする
export async function markNotificationsAsRead(notificationIds: string[]): Promise<ApiResponse<{
  updated_notifications: Notification[]
  failed_updates: Array<{
    notification_id: string
    error: string
  }>
}>> {
  return apiClient.patch<ApiResponse<{
    updated_notifications: Notification[]
    failed_updates: Array<{
      notification_id: string
      error: string
    }>
  }>>('/notifications/mark-read', { notification_ids: notificationIds })
}

// 現在のユーザーの全未読通知を既読にする
export async function markAllMyNotificationsAsRead(): Promise<ApiResponse<{ 
  updated_count: number 
  message: string 
}>> {
  return apiClient.put<ApiResponse<{ 
    updated_count: number 
    message: string 
  }>>('/notifications/me/mark-all-read')
}

// 通知を削除
export async function deleteNotification(id: string): Promise<ApiResponse<{ success: true }>> {
  return apiClient.delete<ApiResponse<{ success: true }>>(`/notifications/${id}`)
}

// 複数の通知を削除
export async function deleteNotifications(notificationIds: string[]): Promise<ApiResponse<{
  deleted_count: number
  failed_deletions: Array<{
    notification_id: string
    error: string
  }>
}>> {
  return apiClient.delete<ApiResponse<{
    deleted_count: number
    failed_deletions: Array<{
      notification_id: string
      error: string
    }>
  }>>('/notifications/batch', { notification_ids: notificationIds })
}

// === Statistics and Counts ===

// ユーザーの未読通知数を取得
export async function getUnreadNotificationCount(userId: string): Promise<ApiResponse<{ 
  count: number 
  by_type: Record<NotificationType, number>
}>> {
  return apiClient.get<ApiResponse<{ 
    count: number 
    by_type: Record<NotificationType, number>
  }>>(`/notifications/user/${userId}/unread/count`)
}

// 現在のユーザーの未読通知数を取得
export async function getMyUnreadNotificationCount(): Promise<ApiResponse<{ 
  count: number 
  by_type: Record<NotificationType, number>
}>> {
  return apiClient.get<ApiResponse<{ 
    count: number 
    by_type: Record<NotificationType, number>
  }>>('/notifications/me/unread/count')
}

// 通知統計を取得
export async function getNotificationStats(
  userId?: string,
  period?: 'today' | 'week' | 'month'
): Promise<ApiResponse<{
  total_notifications: number
  sent_notifications: number
  read_notifications: number
  pending_notifications: number
  failed_notifications: number
  read_rate: number
  by_type: Record<NotificationType, {
    total: number
    sent: number
    read: number
    read_rate: number
  }>
  daily_breakdown?: Array<{
    date: string
    total: number
    sent: number
    read: number
  }>
}>> {
  const params: Record<string, string> = {}
  if (period) params.period = period
  
  const endpoint = userId 
    ? `/notifications/user/${userId}/stats` 
    : '/notifications/me/stats'
    
  return apiClient.get(endpoint, params)
}

// === Advanced Filtering ===

// 通知を検索
export async function searchNotifications(
  query: string,
  params?: {
    user_id?: string
    type?: NotificationType
    status?: NotificationStatus
    limit?: number
  }
): Promise<ApiResponse<{
  notifications: Notification[]
  total_count: number
  search_query: string
}>> {
  const queryParams: Record<string, string> = { q: query }
  
  if (params) {
    if (params.user_id) queryParams.user_id = params.user_id
    if (params.type) queryParams.type = params.type
    if (params.status) queryParams.status = params.status
    if (params.limit) queryParams.limit = params.limit.toString()
  }

  return apiClient.get<ApiResponse<{
    notifications: Notification[]
    total_count: number
    search_query: string
  }>>('/notifications/search', queryParams)
}

// 期間指定で通知を取得
export async function getNotificationsByDateRange(
  startDate: string,
  endDate: string,
  params?: {
    user_id?: string
    type?: NotificationType
    status?: NotificationStatus
  }
): Promise<ApiResponse<{
  notifications: Notification[]
  total_count: number
  date_range: {
    start: string
    end: string
  }
}>> {
  const queryParams: Record<string, string> = {
    start_date: startDate,
    end_date: endDate
  }
  
  if (params) {
    if (params.user_id) queryParams.user_id = params.user_id
    if (params.type) queryParams.type = params.type
    if (params.status) queryParams.status = params.status
  }

  return apiClient.get<ApiResponse<{
    notifications: Notification[]
    total_count: number
    date_range: {
      start: string
      end: string
    }
  }>>('/notifications/by-date-range', queryParams)
}

// === Notification Preferences ===

// ユーザーの通知設定を取得
export async function getNotificationPreferences(userId?: string): Promise<ApiResponse<{
  channels: {
    app: boolean
    line: boolean
    email?: boolean
  }
  types: Record<NotificationType, {
    enabled: boolean
    channels: ('app' | 'line' | 'email')[]
  }>
  schedule: {
    quiet_hours_start?: string
    quiet_hours_end?: string
    weekend_notifications: boolean
  }
}>> {
  const endpoint = userId 
    ? `/notifications/user/${userId}/preferences` 
    : '/notifications/me/preferences'
    
  return apiClient.get(endpoint)
}

// ユーザーの通知設定を更新
export async function updateNotificationPreferences(
  preferences: {
    channels?: {
      app?: boolean
      line?: boolean
      email?: boolean
    }
    types?: Partial<Record<NotificationType, {
      enabled?: boolean
      channels?: ('app' | 'line' | 'email')[]
    }>>
    schedule?: {
      quiet_hours_start?: string
      quiet_hours_end?: string
      weekend_notifications?: boolean
    }
  },
  userId?: string
): Promise<ApiResponse<{ message: string }>> {
  const endpoint = userId 
    ? `/notifications/user/${userId}/preferences` 
    : '/notifications/me/preferences'
    
  return apiClient.put<ApiResponse<{ message: string }>>(endpoint, preferences)
}

// === Utility Functions ===

// 通知がREADステータスかどうかを判定（ヘルパー関数）
export function isNotificationRead(notification: Notification): boolean {
  return notification.status === 'READ'
}

// 通知が未読かどうかを判定（ヘルパー関数）
export function isNotificationUnread(notification: Notification): boolean {
  return notification.status !== 'READ'
}

// 通知の優先度を取得（タイプベース）
export function getNotificationPriority(type: NotificationType): 'high' | 'medium' | 'low' {
  switch (type) {
    case 'TASK_DUE_SOON':
      return 'high'
    case 'TASK_ASSIGNED':
    case 'TASK_COMPLETED':
      return 'medium'
    case 'APP_NOTIFICATION':
    case 'SYSTEM_NOTICE':
      return 'low'
    default:
      return 'low'
  }
}

// 通知タイプの表示名を取得
export function getNotificationTypeLabel(type: NotificationType): string {
  switch (type) {
    case 'APP_NOTIFICATION':
      return 'アプリ通知'
    case 'TASK_ASSIGNED':
      return 'タスク割り当て'
    case 'TASK_COMPLETED':
      return 'タスク完了'
    case 'TASK_DUE_SOON':
      return 'タスク期限間近'
    case 'SYSTEM_NOTICE':
      return 'システム通知'
    default:
      return '通知'
  }
}

// 通知ステータスの表示名を取得
export function getNotificationStatusLabel(status: NotificationStatus): string {
  switch (status) {
    case 'PENDING':
      return '保留中'
    case 'SENT':
      return '送信済み'
    case 'READ':
      return '既読'
    case 'FAILED':
      return '送信失敗'
    default:
      return '不明'
  }
}

// 通知リストから未読カウントを計算
export function calculateUnreadCount(notifications: Notification[]): number {
  return notifications.filter(isNotificationUnread).length
}