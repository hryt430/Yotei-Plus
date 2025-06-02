"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import type { Task } from "@/types"
import { TaskListWithFilters } from "@/components/tasks/task-list-with-filters"
import { TaskStatsChart } from "@/components/tasks/task-stats-chart"
import { useAuth } from "@/providers/auth-provider"
import { getTasks, updateTask } from "@/api/task"
import { handleApiError } from "@/lib/utils"
import { Loader } from "lucide-react"

export default function DashboardPage() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  
  const { isAuthenticated, isLoading: authLoading } = useAuth()
  const router = useRouter()

  // 認証チェック
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/auth/login')
    }
  }, [isAuthenticated, authLoading, router])

  // タスクデータを取得
  useEffect(() => {
    const fetchTasks = async () => {
      if (!isAuthenticated) return
      
      setLoading(true)
      try {
        const response = await getTasks({ page: 1, page_size: 100 })
        
        if (response.success && response.data) {
          setTasks(response.data.tasks || [])
        }
      } catch (err) {
        console.error('Error fetching tasks:', err)
        setError(handleApiError(err))
      } finally {
        setLoading(false)
      }
    }

    fetchTasks()
  }, [isAuthenticated])

  const updateTaskDate = async (taskId: string, newDate: Date) => {
    const task = tasks.find(t => t.id === taskId)
    if (!task) return

    try {
      const response = await updateTask(taskId, {
        ...task,
        due_date: newDate.toISOString()
      })

      if (response.success && response.data) {
        setTasks(prev => prev.map(t => t.id === taskId ? response.data : t))
      }
    } catch (err) {
      console.error('Error updating task date:', err)
    }
  }

  const updateTaskStatus = async (taskId: string, status: "TODO" | "DONE" | "IN_PROGRESS") => {
    const task = tasks.find(t => t.id === taskId)
    if (!task) return

    try {
      const response = await updateTask(taskId, {
        ...task,
        status
      })

      if (response.success && response.data) {
        setTasks(prev => prev.map(t => t.id === taskId ? response.data : t))
      }
    } catch (err) {
      console.error('Error updating task status:', err)
    }
  }

  // ローディング中
  if (authLoading || loading) {
    return (
      <div className="h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <Loader className="h-12 w-12 animate-spin text-gray-900 mx-auto mb-4" />
          <p className="text-gray-600">Loading Dashboard...</p>
        </div>
      </div>
    )
  }

  // エラー表示
  if (error) {
    return (
      <div className="h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-600 mb-4">{error}</div>
          <button
            onClick={() => window.location.reload()}
            className="bg-gray-900 text-white px-4 py-2 rounded-md hover:bg-gray-800"
          >
            再読み込み
          </button>
        </div>
      </div>
    )
  }

  // 認証されていない場合
  if (!isAuthenticated) {
    return null
  }

  return (
    <div className="h-[calc(100vh-89px)] flex">
      {/* Left Panel - Task List with Filters (3/5) */}
      <div className="w-3/5 border-r border-gray-200">
        <TaskListWithFilters 
          tasks={tasks} 
          onUpdateTaskDate={updateTaskDate} 
          onUpdateTaskStatus={updateTaskStatus} 
        />
      </div>

      {/* Right Panel - Statistics (2/5) */}
      <div className="w-2/5 bg-gray-50/30">
        <TaskStatsChart tasks={tasks} />
      </div>
    </div>
  )
}