"use client"

import { useEffect, useState } from "react"
import { X, Clock, AlertTriangle, Calendar, Flag } from "lucide-react"
import { Button } from "@/components/ui/forms/button"
import { Badge } from "@/components/ui/data-display/badge"
import type { DeadlineNotification } from "@/types"

interface DeadlineNotificationToastProps {
  notification: DeadlineNotification
  onClose: () => void
  onMarkComplete?: (taskId: string) => void
  onSnooze?: (taskId: string, minutes: number) => void
}

export function DeadlineNotificationToast({
  notification,
  onClose,
  onMarkComplete,
  onSnooze,
}: DeadlineNotificationToastProps) {
  const [isVisible, setIsVisible] = useState(false)
  const [isLeaving, setIsLeaving] = useState(false)

  useEffect(() => {
    // Slide in animation
    setTimeout(() => setIsVisible(true), 100)

    // Auto-dismiss after 8 seconds for non-overdue notifications
    if (notification.urgencyLevel !== "overdue") {
      const timer = setTimeout(() => {
        handleClose()
      }, 8000)
      return () => clearTimeout(timer)
    }
  }, [notification.urgencyLevel])

  const handleClose = () => {
    setIsLeaving(true)
    setTimeout(() => {
      onClose()
    }, 300)
  }

  const handleMarkComplete = () => {
    if (onMarkComplete) {
      onMarkComplete(notification.taskId)
    }
    handleClose()
  }

  const handleSnooze = (minutes: number) => {
    if (onSnooze) {
      onSnooze(notification.taskId, minutes)
    }
    handleClose()
  }

  const getUrgencyStyles = () => {
    switch (notification.urgencyLevel) {
      case "overdue":
        return {
          border: "border-l-red-500",
          bg: "bg-red-50/90",
          icon: <AlertTriangle className="w-5 h-5 text-red-600" />,
          title: "Task Overdue!",
          titleColor: "text-red-800",
        }
      case "due-soon":
        return {
          border: "border-l-orange-500",
          bg: "bg-orange-50/90",
          icon: <Clock className="w-5 h-5 text-orange-600" />,
          title: "Due Soon",
          titleColor: "text-orange-800",
        }
      case "upcoming":
        return {
          border: "border-l-blue-500",
          bg: "bg-blue-50/90",
          icon: <Calendar className="w-5 h-5 text-blue-600" />,
          title: "Upcoming Deadline",
          titleColor: "text-blue-800",
        }
      default:
        return {
          border: "border-l-gray-500",
          bg: "bg-gray-50/90",
          icon: <Clock className="w-5 h-5 text-gray-600" />,
          title: "Deadline Reminder",
          titleColor: "text-gray-800",
        }
    }
  }

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case "HIGH":
        return "bg-red-100 text-red-800 border-red-200"
      case "MEDIUM":
        return "bg-yellow-100 text-yellow-800 border-yellow-200"
      case "LOW":
        return "bg-green-100 text-green-800 border-green-200"
      default:
        return "bg-gray-100 text-gray-800 border-gray-200"
    }
  }

  const styles = getUrgencyStyles()

  return (
    <div
      className={`fixed top-4 right-4 z-50 w-96 bg-white rounded-lg shadow-2xl border border-gray-200 ${styles.border} border-l-4 transition-all duration-300 ${
        isVisible && !isLeaving ? "translate-x-0 opacity-100" : "translate-x-full opacity-0"
      } ${styles.bg} backdrop-blur-sm`}
    >
      <div className="p-4">
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center space-x-3">
            {styles.icon}
            <div>
              <h4 className={`font-semibold text-sm ${styles.titleColor}`}>{styles.title}</h4>
              <p className="text-xs text-gray-500 mt-0.5">{notification.timeUntilDue}</p>
            </div>
          </div>
          <Button variant="ghost" size="sm" onClick={handleClose} className="p-1 h-auto">
            <X className="w-4 h-4" />
          </Button>
        </div>

        <div className="mb-3">
          <div className="flex items-center justify-between mb-2">
            <h3 className="font-medium text-gray-900 text-sm">{notification.taskTitle}</h3>
            <div className="flex items-center space-x-2">
              <Badge variant="outline" className={`text-xs ${getPriorityColor(notification.priority)}`}>
                <Flag className="w-2 h-2 mr-1" />
                {notification.priority}
              </Badge>
            </div>
          </div>
          <p className="text-xs text-gray-600 line-clamp-2">{notification.taskDescription}</p>
          <div className="flex items-center mt-2 text-xs text-gray-500">
            <Calendar className="w-3 h-3 mr-1" />
            <span>
              Due: {notification.dueDate.toLocaleDateString()} at{" "}
              {notification.dueDate.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
            </span>
          </div>
        </div>

        <div className="flex space-x-2">
          {onMarkComplete && (
            <Button
              size="sm"
              onClick={handleMarkComplete}
              className="flex-1 bg-green-600 hover:bg-green-700 text-white text-xs"
            >
              Mark Complete
            </Button>
          )}
          {onSnooze && notification.urgencyLevel !== "overdue" && (
            <div className="flex space-x-1">
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleSnooze(15)}
                className="text-xs px-2 border-gray-300 hover:bg-gray-50"
              >
                15m
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleSnooze(60)}
                className="text-xs px-2 border-gray-300 hover:bg-gray-50"
              >
                1h
              </Button>
            </div>
          )}
        </div>

        <div className="text-xs text-gray-400 mt-2 text-center">
          {notification.createdAt.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
        </div>
      </div>
    </div>
  )
}

export default DeadlineNotificationToast
