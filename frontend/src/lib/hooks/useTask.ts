import { useState, useCallback } from 'react'
import { 
  getTasks, 
  getTask, 
  createTask, 
  updateTask, 
  deleteTask,
  assignTask,
  changeTaskStatus,
  searchTasks,
  GetTasksParams
} from '@/api/task'
import { Task, TaskFilter, TaskRequest, TasksState } from '@/types'
import { handleApiError } from '@/lib/utils'

export default function useTasks() {
  const [state, setState] = useState<TasksState>({
    tasks: [],
    isLoading: false,
    error: null,
    pagination: {
      page: 1,
      limit: 10,
      total: 0,
    },
    filters: {},
    sort: {
      field: 'created_at',
      direction: 'desc',
    },
  })

  // タスク一覧を取得
  const fetchTasks = useCallback(async () => {
    setState(prev => ({ ...prev, isLoading: true, error: null }))
    
    try {
      const params: GetTasksParams = {
        ...state.filters,
        page: state.pagination.page,
        page_size: state.pagination.limit,
        sort_field: state.sort.field,
        sort_direction: state.sort.direction === 'asc' ? 'ASC' : 'DESC',
      }
      
      const response = await getTasks(params)
      
      if (response.success && response.data) {
        setState(prev => ({
          ...prev,
          tasks: response.data.tasks,
          pagination: {
            page: response.data.page,
            limit: response.data.page_size,
            total: response.data.total_count,
          },
          isLoading: false,
          error: null,
        }))
      }
    } catch (error) {
      const errorMessage = handleApiError(error)
      setState(prev => ({ 
        ...prev, 
        isLoading: false, 
        error: errorMessage 
      }))
    }
  }, [state.filters, state.pagination.page, state.pagination.limit, state.sort])

  // タスクリストを最新の状態で再取得
  const refreshTasks = useCallback(async () => {
    await fetchTasks()
  }, [fetchTasks])

  // 特定のタスクを取得
  const fetchTaskById = async (id: string): Promise<Task | null> => {
    try {
      const response = await getTask(id)
      if (response.success && response.data) {
        return response.data
      }
      return null
    } catch (error) {
      throw error
    }
  }

  // 新しいタスクを作成
  const addTask = async (taskData: TaskRequest): Promise<Task> => {
    try {
      const response = await createTask(taskData)
      
      if (response.success && response.data) {
        // リストに新しいタスクを追加
        setState(prev => ({
          ...prev,
          tasks: [response.data, ...prev.tasks],
          pagination: {
            ...prev.pagination,
            total: prev.pagination.total + 1,
          },
        }))
        
        return response.data
      }
      throw new Error('タスクの作成に失敗しました')
    } catch (error) {
      throw error
    }
  }

  // タスクを編集
  const editTask = async (id: string, taskData: TaskRequest): Promise<Task> => {
    try {
      const response = await updateTask(id, taskData)
      
      if (response.success && response.data) {
        // リスト内のタスクを更新
        setState(prev => ({
          ...prev,
          tasks: prev.tasks.map(task => 
            task.id === id ? response.data : task
          ),
        }))
        
        return response.data
      }
      throw new Error('タスクの更新に失敗しました')
    } catch (error) {
      throw error
    }
  }

  // タスクを削除
  const removeTask = async (id: string): Promise<void> => {
    try {
      await deleteTask(id)
      
      // リストからタスクを削除
      setState(prev => ({
        ...prev,
        tasks: prev.tasks.filter(task => task.id !== id),
        pagination: {
          ...prev.pagination,
          total: Math.max(0, prev.pagination.total - 1),
        },
      }))
    } catch (error) {
      throw error
    }
  }

  // タスクのステータスを更新
  const updateTaskStatus = async (taskId: string, status: Task['status']): Promise<Task> => {
    try {
      const response = await changeTaskStatus(taskId, status)
      
      if (response.success && response.data) {
        // リスト内のタスクを更新
        setState(prev => ({
          ...prev,
          tasks: prev.tasks.map(task => 
            task.id === taskId ? response.data : task
          ),
        }))
        
        return response.data
      }
      throw new Error('ステータスの更新に失敗しました')
    } catch (error) {
      throw error
    }
  }

  // タスクを割り当て
  const assignTaskToUser = async (taskId: string, assigneeId: string): Promise<Task> => {
    try {
      const response = await assignTask(taskId, assigneeId)
      
      if (response.success && response.data) {
        // リスト内のタスクを更新
        setState(prev => ({
          ...prev,
          tasks: prev.tasks.map(task => 
            task.id === taskId ? response.data : task
          ),
        }))
        
        return response.data
      }
      throw new Error('タスクの割り当てに失敗しました')
    } catch (error) {
      throw error
    }
  }

  // タスクを検索
  const performSearch = async (query: string): Promise<void> => {
    setState(prev => ({ ...prev, isLoading: true, error: null }))
    
    try {
      const response = await searchTasks({ q: query, limit: state.pagination.limit })
      
      if (response.success && response.data) {
        setState(prev => ({
          ...prev,
          tasks: response.data.tasks,
          pagination: {
            ...prev.pagination,
            total: response.data.count,
            page: 1,
          },
          isLoading: false,
          error: null,
        }))
      }
    } catch (error) {
      const errorMessage = handleApiError(error)
      setState(prev => ({ 
        ...prev, 
        isLoading: false, 
        error: errorMessage 
      }))
    }
  }

  // フィルターを設定
  const setFilters = (filters: TaskFilter): void => {
    setState(prev => ({
      ...prev,
      filters,
      pagination: {
        ...prev.pagination,
        page: 1, // フィルター変更時は1ページ目に戻す
      },
    }))
  }

  // ソートを設定
  const setSorting = (field: string, direction: 'asc' | 'desc'): void => {
    setState(prev => ({
      ...prev,
      sort: { field, direction },
      pagination: {
        ...prev.pagination,
        page: 1, // ソート変更時は1ページ目に戻す
      },
    }))
  }

  // ページを設定
  const setPage = (page: number): void => {
    setState(prev => ({
      ...prev,
      pagination: {
        ...prev.pagination,
        page,
      },
    }))
  }

  // エラーをクリア
  const clearError = (): void => {
    setState(prev => ({ ...prev, error: null }))
  }

  return {
    // 状態
    tasks: state.tasks,
    isLoading: state.isLoading,
    error: state.error,
    pagination: state.pagination,
    filters: state.filters,
    sort: state.sort,
    
    // アクション
    fetchTasks,
    refreshTasks,
    fetchTaskById,
    addTask,
    editTask,
    removeTask,
    updateTaskStatus,
    assignTaskToUser,
    performSearch,
    setFilters,
    setSorting,
    setPage,
    clearError,
  }
}