"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { Calendar } from "@/components/calendar"
import { TaskList } from "@/components/task-list"
import { TaskListWithFilters } from "@/components/task-list-with-filters"
import { TaskAnalyticsPanel } from "@/components/task-analytics-panel"
import { TaskCreationModal } from "@/components/task-creation-modal"
import { AddFriendModal } from "@/components/add-friend-modal"
import { Sidebar } from "@/components/sidebar"
import { NotificationPanel } from "@/components/notification-panel"
import { NotificationProvider, useNotifications } from "@/components/notification-provider"
import { useAuth } from "@/providers/auth-provider"
import { getTasks, createTask, updateTask, deleteTask } from "@/api/task"
import { getUsers } from "@/api/auth"
import { handleApiError } from "@/lib/utils"
import { success, error as showError } from "@/hooks/use-toast"
import type { Task, User, TaskFormData } from "@/types"
import { ChevronDown, Loader } from "lucide-react"

function TaskManagementContent() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [users, setUsers] = useState<User[]>([])
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [isFriendModalOpen, setIsFriendModalOpen] = useState(false)
  const [currentPage, setCurrentPage] = useState<"dashboard" | "tasks">("dashboard")
  const [showTaskDemo, setShowTaskDemo] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const router = useRouter()
  const { user, isAuthenticated, isLoading: authLoading } = useAuth()

  const { notifications, markAsRead, markAllAsRead, deleteNotification, handleFriendAction, addNotification } =
    useNotifications()

  // 認証チェック
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/auth/login')
    }
  }, [isAuthenticated, authLoading, router])

  // データ取得
  useEffect(() => {
    const fetchData = async () => {
      if (!isAuthenticated) return
      
      setLoading(true)
      try {
        // タスクとユーザー一覧を並行取得
        const [tasksResponse, usersResponse] = await Promise.all([
          getTasks({ page: 1, page_size: 50 }),
          getUsers()
        ])

        if (tasksResponse.success && tasksResponse.data) {
          setTasks(tasksResponse.data.tasks || [])
        }

        if (usersResponse.success && usersResponse.data) {
          setUsers(usersResponse.data.map(u => ({
            id: u.id,
            username: u.username,
            email: u.email,
            role: u.role
          })))
        }

      } catch (err) {
        console.error('Error fetching data:', err)
        setError(handleApiError(err))
      } finally {
        setLoading(false)
      }
    }

    fetchData()
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
        success('期限を更新しました')
      }
    } catch (err) {
      showError('更新エラー', handleApiError(err))
    }
  }

  const updateTaskStatus = async (taskId: string, status: "TODO" | "IN_PROGRESS" | "DONE") => {
    const task = tasks.find(t => t.id === taskId)
    if (!task) return

    try {
      const response = await updateTask(taskId, {
        ...task,
        status
      })

      if (response.success && response.data) {
        setTasks(prev => prev.map(t => t.id === taskId ? response.data : t))
        success('ステータスを更新しました')
      }
    } catch (err) {
      showError('更新エラー', handleApiError(err))
    }
  }

  const createNewTask = async (formData: TaskFormData) => {
    try {
      const response = await createTask({
        title: formData.title,
        description: formData.description,
        priority: formData.priority,
        due_date: formData.due_date,
      })

      if (response.success && response.data) {
        setTasks(prev => [...prev, response.data])
        setIsModalOpen(false)
        success('タスクを作成しました')
      }
    } catch (err) {
      showError('作成エラー', handleApiError(err))
    }
  }

  // デモ機能（v0-designの機能を保持）
  const simulateFriendRequest = () => {
    addNotification({
      type: "friend_request",
      title: "Friend Request",
      message: "Yamada-san sent you a friend request",
      actionData: {
        friendId: "user_456",
        friendName: "Yamada-san",
        friendEmail: "yamada@example.com",
      },
    })
  }

  const addDemoTasks = (type: "mixed" | "completed" | "pending" | "progress" | "reset") => {
    setShowTaskDemo(false)

    if (type === "reset") {
      // 元のタスクを再取得
      getTasks({ page: 1, page_size: 50 }).then(response => {
        if (response.success && response.data) {
          setTasks(response.data.tasks || [])
        }
      })
      return
    }

    if (type === "completed") {
      setTasks(prev =>
        prev.map(task => {
          if (task.status === "TODO" && Math.random() > 0.5) {
            return { ...task, status: "DONE" as const }
          }
          return task
        })
      )
      return
    }

    if (type === "progress") {
      setTasks(prev =>
        prev.map(task => {
          if (task.status === "TODO" && Math.random() > 0.6) {
            return { ...task, status: "IN_PROGRESS" as const }
          }
          return task
        })
      )
      return
    }

    // その他のデモタスク追加ロジック...
  }

  // ローディング中
  if (authLoading || loading) {
    return (
      <div className="h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <Loader className="h-12 w-12 animate-spin text-gray-900 mx-auto mb-4" />
          <p className="text-gray-600">Loading TaskFlow...</p>
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
    <div className="h-screen bg-white flex overflow-hidden">
      {/* Sidebar */}
      <Sidebar
        currentPage={currentPage}
        onNavigate={setCurrentPage}
        onCreateTask={() => setIsModalOpen(true)}
        onAddFriend={() => setIsFriendModalOpen(true)}
      />

      {/* Main Content */}
      <div className="flex-1 flex flex-col h-full min-h-0">
        {/* Header */}
        <header className="border-b border-gray-200 bg-white flex-shrink-0">
          <div className="px-6 py-3 flex items-center justify-between">
            <div>
              <h1 className="text-xl font-semibold text-gray-900">
                {currentPage === "dashboard" ? "Task Management" : "All Tasks"}
              </h1>
              <p className="text-xs text-gray-600 mt-1">
                {currentPage === "dashboard" ? "Organize your work efficiently" : "Search and manage all your tasks"}
              </p>
            </div>

            {/* Demo Controls and Notification Panel */}
            <div className="flex items-center space-x-3">
              <button
                onClick={simulateFriendRequest}
                className="text-xs bg-blue-100 text-blue-700 px-2 py-1 rounded hover:bg-blue-200 transition-colors"
              >
                Demo: Friend Request
              </button>
              <div className="relative">
                <button
                  onClick={() => setShowTaskDemo(!showTaskDemo)}
                  className="text-xs bg-green-100 text-green-700 px-2 py-1 rounded hover:bg-green-200 transition-colors flex items-center"
                >
                  Demo: Progress <ChevronDown className="w-3 h-3 ml-1" />
                </button>

                {showTaskDemo && (
                  <>
                    <div className="fixed inset-0 z-40" onClick={() => setShowTaskDemo(false)} />
                    <div className="absolute right-0 top-full mt-2 w-56 bg-white rounded-lg shadow-lg border border-gray-200 z-50 p-2">
                      <div className="space-y-1">
                        <button
                          onClick={() => addDemoTasks("mixed")}
                          className="w-full text-left px-2 py-1 text-xs hover:bg-gray-100 rounded"
                        >
                          Add Mixed Tasks (5)
                        </button>
                        <button
                          onClick={() => addDemoTasks("completed")}
                          className="w-full text-left px-2 py-1 text-xs hover:bg-gray-100 rounded"
                        >
                          Complete Random Tasks
                        </button>
                        <button
                          onClick={() => addDemoTasks("pending")}
                          className="w-full text-left px-2 py-1 text-xs hover:bg-gray-100 rounded"
                        >
                          Add Pending Tasks (3)
                        </button>
                        <button
                          onClick={() => addDemoTasks("progress")}
                          className="w-full text-left px-2 py-1 text-xs hover:bg-gray-100 rounded"
                        >
                          Set Tasks In Progress
                        </button>
                        <button
                          onClick={() => addDemoTasks("reset")}
                          className="w-full text-left px-2 py-1 text-xs hover:bg-gray-100 rounded text-red-600"
                        >
                          Reset to Original
                        </button>
                      </div>
                    </div>
                  </>
                )}
              </div>
              <NotificationPanel
                notifications={notifications}
                onAction={handleFriendAction}
                onMarkAsRead={markAsRead}
                onMarkAllAsRead={markAllAsRead}
                onDelete={deleteNotification}
              />
            </div>
          </div>
        </header>

        {/* Page Content */}
        {currentPage === "dashboard" ? (
          <div className="flex flex-1 min-h-0">
            {/* Left Panel - Simple Task List (2/5) */}
            <div className="w-2/5 border-r border-gray-200 bg-gray-50/30">
              <TaskList
                tasks={tasks}
                onUpdateTaskDate={updateTaskDate}
                onUpdateTaskStatus={updateTaskStatus}
                onCreateTask={() => setIsModalOpen(true)}
              />
            </div>

            {/* Right Panel - Calendar (3/5) */}
            <div className="w-3/5 bg-white">
              <Calendar tasks={tasks} onUpdateTaskDate={updateTaskDate} />
            </div>
          </div>
        ) : (
          <div className="flex flex-1 min-h-0">
            {/* Left Panel - Task List with Filters (3/5) */}
            <div className="w-3/5 border-r border-gray-200">
              <TaskListWithFilters
                tasks={tasks}
                onUpdateTaskDate={updateTaskDate}
                onUpdateTaskStatus={updateTaskStatus}
              />
            </div>

            {/* Right Panel - Analytics (2/5) */}
            <div className="w-2/5 min-h-0">
              <TaskAnalyticsPanel tasks={tasks} />
            </div>
          </div>
        )}
      </div>

      {/* Task Creation Modal */}
      <TaskCreationModal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        onCreateTask={createNewTask} 
      />

      {/* Add Friend Modal */}
      <AddFriendModal 
        isOpen={isFriendModalOpen} 
        onClose={() => setIsFriendModalOpen(false)} 
      />
    </div>
  )
}

export default function TaskManagement() {
  return (
    <NotificationProvider>
      <TaskManagementContent />
    </NotificationProvider>
  )
}