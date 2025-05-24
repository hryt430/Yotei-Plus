// src/api/tasks/[id]/route.ts
// 個別タスクのAPI関数

import { 
  ApiResponse, 
  TaskRequest, 
  TaskResponse
} from '@/types'
import { apiClient } from '@/api/client/route'

// 特定のタスクを取得
export async function getTask(id: string): Promise<TaskResponse> {
  return apiClient.get<TaskResponse>(`/api/tasks/${id}`)
}

// タスクを更新
export async function updateTask(id: string, data: TaskRequest): Promise<TaskResponse> {
  return apiClient.put<TaskResponse>(`/api/tasks/${id}`, data)
}

// タスクを削除
export async function deleteTask(id: string): Promise<ApiResponse<{ success: true }>> {
  return apiClient.delete<ApiResponse<{ success: true }>>(`/api/tasks/${id}`)
}

// タスクのステータスを変更
export async function changeTaskStatus(
  id: string, 
  status: 'TODO' | 'IN_PROGRESS' | 'DONE'
): Promise<TaskResponse> {
  return apiClient.patch<TaskResponse>(`/api/tasks/${id}/status`, { status })
}