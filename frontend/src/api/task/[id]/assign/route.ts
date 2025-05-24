// src/api/tasks/[id]/assign/route.ts
// タスク割り当てのAPI関数

import { TaskResponse } from '@/types'
import { apiClient } from '@/api/client/route'

// タスク割り当てリクエスト
export interface AssignTaskRequest {
  assignee_id: string
}

// タスクをユーザーに割り当てる
export async function assignTask(taskId: string, data: AssignTaskRequest): Promise<TaskResponse> {
  return apiClient.post<TaskResponse>(`/api/tasks/${taskId}/assign`, data)
}