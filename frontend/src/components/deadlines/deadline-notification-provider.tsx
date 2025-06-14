"use client"

import { createContext, useContext, useState, useCallback, type ReactNode } from "react"
import { DeadlineNotificationToast } from "@/components/deadlines/deadline-notification-toast"
import type { DeadlineNotification } from "@/types"
import type { Task } from "@/types/task"

interface DeadlineNotificationContextType {
  notifications: DeadlineNotification[]
  addNotification: (notification: Omit<DeadlineNotification, "id" | "createdAt" | "isRead">) => void
  markAsRead: (id: string) => void
  dismissNotification: (id: string) => void
  clearAllNotifications: () => void
  markTaskComplete: (taskId: string) => void
  snoozeTask: (taskId: string, minutes: number) => void
  checkTaskDeadlines: (tasks: Task[]) => void
}

const DeadlineNotificationContext = createContext<DeadlineNotificationContextType | undefined>(undefined)

export function useDeadlineNotifications() {
  const context = useContext(DeadlineNotificationContext)
  if (!context) {
    throw new Error("useDeadlineNotifications must be used within a DeadlineNotificationProvider")
  }
  return context
}

interface DeadlineNotificationProviderProps {
  children: ReactNode
}

export function DeadlineNotificationProvider({ children }: DeadlineNotificationProviderProps) {
  const [notifications, setNotifications] = useState<DeadlineNotification[]>([])
  const [toastNotifications, setToastNotifications] = useState<DeadlineNotification[]>([])
  const [snoozedTasks, setSnoozedTasks] = useState<Record<string, Date>>({})

  const addNotification = useCallback((notificationData: Omit<DeadlineNotification, "id" | "createdAt" | "isRead">) => {
    const newNotification: DeadlineNotification = {
      ...notificationData,
      id: Date.now().toString(),
      createdAt: new Date(),
      isRead: false,
    }

    setNotifications((prev) => [newNotification, ...prev])

    // Show toast notification
    setToastNotifications((prev) => [...prev, newNotification])
  }, [])

  const markAsRead = useCallback((id: string) => {
    setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, isRead: true } : n)))
  }, [])

  const dismissNotification = useCallback((id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id))
  }, [])

  const clearAllNotifications = useCallback(() => {
    setNotifications([])
  }, [])

  const markTaskComplete = useCallback((taskId: string) => {
    // Remove all notifications for this task
    setNotifications((prev) => prev.filter((n) => n.taskId !== taskId))
    // Here you would also update the task status in your main task list
    console.log("Mark task complete:", taskId)
  }, [])

  const snoozeTask = useCallback((taskId: string, minutes: number) => {
    const snoozeUntil = new Date(Date.now() + minutes * 60 * 1000)
    setSnoozedTasks((prev) => ({ ...prev, [taskId]: snoozeUntil }))

    // Remove current notifications for this task
    setNotifications((prev) => prev.filter((n) => n.taskId !== taskId))

    console.log(`Task ${taskId} snoozed for ${minutes} minutes until ${snoozeUntil}`)
  }, [])

  const getTimeUntilDue = (
    dueDate: Date,
  ): { timeString: string; urgencyLevel: "upcoming" | "due-soon" | "overdue" } => {
    const now = new Date()
    const timeDiff = dueDate.getTime() - now.getTime()
    const minutesDiff = Math.floor(timeDiff / (1000 * 60))
    const hoursDiff = Math.floor(timeDiff / (1000 * 60 * 60))
    const daysDiff = Math.floor(timeDiff / (1000 * 60 * 60 * 24))

    if (timeDiff < 0) {
      const overdueDays = Math.abs(daysDiff)
      const overdueHours = Math.abs(hoursDiff) % 24
      if (overdueDays > 0) {
        return { timeString: `${overdueDays} day${overdueDays > 1 ? "s" : ""} overdue`, urgencyLevel: "overdue" }
      } else if (overdueHours > 0) {
        return { timeString: `${overdueHours} hour${overdueHours > 1 ? "s" : ""} overdue`, urgencyLevel: "overdue" }
      } else {
        return {
          timeString: `${Math.abs(minutesDiff)} minute${Math.abs(minutesDiff) > 1 ? "s" : ""} overdue`,
          urgencyLevel: "overdue",
        }
      }
    } else if (hoursDiff < 2) {
      return { timeString: `${minutesDiff} minute${minutesDiff > 1 ? "s" : ""} remaining`, urgencyLevel: "due-soon" }
    } else if (hoursDiff < 24) {
      return { timeString: `${hoursDiff} hour${hoursDiff > 1 ? "s" : ""} remaining`, urgencyLevel: "due-soon" }
    } else if (daysDiff < 3) {
      return { timeString: `${daysDiff} day${daysDiff > 1 ? "s" : ""} remaining`, urgencyLevel: "due-soon" }
    } else {
      return { timeString: `${daysDiff} day${daysDiff > 1 ? "s" : ""} remaining`, urgencyLevel: "upcoming" }
    }
  }

  const checkTaskDeadlines = useCallback(
    (tasks: Task[]) => {
      const now = new Date()

      tasks.forEach((task) => {
        // Skip completed tasks
        if (task.status === "DONE") return

        // Skip snoozed tasks
        const snoozeUntil = snoozedTasks[task.id]
        if (snoozeUntil && now < snoozeUntil) return

        // Check if we already have a notification for this task
        const existingNotification = notifications.find((n) => n.taskId === task.id)
        if (existingNotification) return

        // Skip tasks without due date
        if (!task.due_date) return
        
        const dueDate = new Date(task.due_date)
        const timeDiff = dueDate.getTime() - now.getTime()
        const hoursDiff = timeDiff / (1000 * 60 * 60)

        // Create notification based on time remaining
        let shouldNotify = false

        if (timeDiff < 0) {
          // Overdue
          shouldNotify = true
        } else if (hoursDiff <= 2) {
          // Due within 2 hours
          shouldNotify = true
        } else if (hoursDiff <= 24 && task.priority === "HIGH") {
          // High priority tasks due within 24 hours
          shouldNotify = true
        } else if (hoursDiff <= 72 && task.priority === "HIGH") {
          // High priority tasks due within 3 days
          shouldNotify = true
        }

        if (shouldNotify) {
          const { timeString, urgencyLevel } = getTimeUntilDue(dueDate)

          addNotification({
            taskId: task.id,
            taskTitle: task.title,
            taskDescription: task.description,
            dueDate: dueDate,
            priority: task.priority,
            category: task.category,
            timeUntilDue: timeString,
            urgencyLevel,
          })
        }
      })
    },
    [notifications, snoozedTasks, addNotification],
  )

  const removeToastNotification = useCallback((id: string) => {
    setToastNotifications((prev) => prev.filter((n) => n.id !== id))
  }, [])

  const handleToastMarkComplete = useCallback(
    (taskId: string) => {
      markTaskComplete(taskId)
      // Remove toast notifications for this task
      setToastNotifications((prev) => prev.filter((n) => n.taskId !== taskId))
    },
    [markTaskComplete],
  )

  const handleToastSnooze = useCallback(
    (taskId: string, minutes: number) => {
      snoozeTask(taskId, minutes)
      // Remove toast notifications for this task
      setToastNotifications((prev) => prev.filter((n) => n.taskId !== taskId))
    },
    [snoozeTask],
  )

  return (
    <DeadlineNotificationContext.Provider
      value={{
        notifications,
        addNotification,
        markAsRead,
        dismissNotification,
        clearAllNotifications,
        markTaskComplete,
        snoozeTask,
        checkTaskDeadlines,
      }}
    >
      {children}

      {/* Toast Notifications */}
      <div className="fixed top-4 right-4 z-50 space-y-2">
        {toastNotifications.map((notification) => (
          <DeadlineNotificationToast
            key={notification.id}
            notification={notification}
            onClose={() => removeToastNotification(notification.id)}
            onMarkComplete={handleToastMarkComplete}
            onSnooze={handleToastSnooze}
          />
        ))}
      </div>
    </DeadlineNotificationContext.Provider>
  )
}
