// src/api/notifications/route.ts
// 通知関連のAPI関数

import { ApiResponse, Notification } from '@/types'
import { apiClient } from '@/api/client/route'

// 通知作成の入力データ
export interface CreateNotificationInput {
  user_id: string
  type: 'APP_NOTIFICATION' | 'TASK_ASSIGNED' | 'TASK_COMPLETED' | 'TASK_DUE_SOON' | 'SYSTEM_NOTICE'
  title: string
  message: string
  metadata?: Record<string, string>
  channels: string[] // "app", "line" などのチャネル指定
}

// 通知一覧取得の入力データ
export interface GetNotificationsParams {
  limit?: number
  offset?: number
  unreadOnly?: boolean
}

// 新しい通知を作成
export async function createNotification(data: CreateNotificationInput): Promise<ApiResponse<Notification>> {
  return apiClient.post<ApiResponse<Notification>>('/api/notifications', data)
}

// 特定の通知を取得
export async function getNotification(id: string): Promise<ApiResponse<Notification>> {
  return apiClient.get<ApiResponse<Notification>>(`/api/notifications/${id}`)
}

// ユーザーの通知一覧を取得
export async function getUserNotifications(
  userId: string, 
  params?: GetNotificationsParams
): Promise<ApiResponse<Notification[]>> {
  const queryParams = {
    limit: params?.limit?.toString() || '10',
    offset: params?.offset?.toString() || '0',
    ...(params?.unreadOnly && { unreadOnly: 'true' })
  }
  
  return apiClient.get<ApiResponse<Notification[]>>(`/api/notifications/user/${userId}`, queryParams)
}

// 通知を送信
export async function sendNotification(id: string): Promise<ApiResponse<{ message: string }>> {
  return apiClient.post<ApiResponse<{ message: string }>>(`/api/notifications/${id}/send`)
}

// 通知を既読にする
export async function markNotificationAsRead(id: string): Promise<ApiResponse<{ message: string }>> {
  return apiClient.put<ApiResponse<{ message: string }>>(`/api/notifications/${id}/read`)
}

// 複数の通知を既読にする
export async function markNotificationsAsRead(notificationIds: string[]): Promise<ApiResponse<{ message: string }>> {
  return apiClient.patch<ApiResponse<{ message: string }>>('/api/notifications/mark-read', { notificationIds })
}

// ユーザーの未読通知数を取得
export async function getUnreadNotificationCount(userId: string): Promise<ApiResponse<{ count: number }>> {
  return apiClient.get<ApiResponse<{ count: number }>>(`/api/notifications/user/${userId}/unread/count`)
}