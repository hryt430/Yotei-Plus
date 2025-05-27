"use client"

import { createContext, useContext, useState, useCallback, type ReactNode } from "react"
import { ToastNotification } from "@/components/toast-notification"
import type { Notification } from "@/types/notification"

interface NotificationContextType {
  notifications: Notification[]
  addNotification: (notification: Omit<Notification, "id" | "timestamp" | "isRead">) => void
  markAsRead: (id: string) => void
  markAllAsRead: () => void
  deleteNotification: (id: string) => void
  handleFriendAction: (action: "accept" | "reject", notificationId: string) => void
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined)

export function useNotifications() {
  const context = useContext(NotificationContext)
  if (!context) {
    throw new Error("useNotifications must be used within a NotificationProvider")
  }
  return context
}

interface NotificationProviderProps {
  children: ReactNode
}

export function NotificationProvider({ children }: NotificationProviderProps) {
  const [notifications, setNotifications] = useState<Notification[]>([
    // Sample notifications for demo
    {
      id: "1",
      type: "friend_request",
      title: "Friend Request",
      message: "Tanaka-san sent you a friend request",
      timestamp: new Date(Date.now() - 300000), // 5 minutes ago
      isRead: false,
      actionData: {
        friendId: "user_123",
        friendName: "Tanaka-san",
        friendEmail: "tanaka@example.com",
      },
    },
    {
      id: "2",
      type: "friend_accepted",
      title: "Friend Request Accepted",
      message: "Sarah Johnson accepted your friend request",
      timestamp: new Date(Date.now() - 3600000), // 1 hour ago
      isRead: false,
    },
  ])
  const [toastNotifications, setToastNotifications] = useState<Notification[]>([])

  const addNotification = useCallback((notificationData: Omit<Notification, "id" | "timestamp" | "isRead">) => {
    const newNotification: Notification = {
      ...notificationData,
      id: Date.now().toString(),
      timestamp: new Date(),
      isRead: false,
    }

    setNotifications((prev) => [newNotification, ...prev])

    // Show toast notification
    setToastNotifications((prev) => [...prev, newNotification])
  }, [])

  const markAsRead = useCallback((id: string) => {
    setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, isRead: true } : n)))
  }, [])

  const markAllAsRead = useCallback(() => {
    setNotifications((prev) => prev.map((n) => ({ ...n, isRead: true })))
  }, [])

  const deleteNotification = useCallback((id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id))
  }, [])

  const handleFriendAction = useCallback(
    (action: "accept" | "reject", notificationId: string) => {
      const notification = notifications.find((n) => n.id === notificationId)
      if (!notification || !notification.actionData) return

      // Remove the original friend request notification
      setNotifications((prev) => prev.filter((n) => n.id !== notificationId))

      // Add a confirmation notification
      const confirmationNotification: Notification = {
        id: Date.now().toString(),
        type: action === "accept" ? "friend_accepted" : "friend_rejected",
        title: action === "accept" ? "Friend Request Accepted" : "Friend Request Declined",
        message:
          action === "accept"
            ? `You are now friends with ${notification.actionData.friendName}`
            : `You declined the friend request from ${notification.actionData.friendName}`,
        timestamp: new Date(),
        isRead: false,
      }

      setNotifications((prev) => [confirmationNotification, ...prev])

      // Here you would typically make an API call to the backend
      console.log(`${action} friend request from ${notification.actionData.friendName}`)
    },
    [notifications],
  )

  const removeToastNotification = useCallback((id: string) => {
    setToastNotifications((prev) => prev.filter((n) => n.id !== id))
  }, [])

  const handleToastAction = useCallback(
    (action: "accept" | "reject", notificationId: string) => {
      handleFriendAction(action, notificationId)
      removeToastNotification(notificationId)
    },
    [handleFriendAction, removeToastNotification],
  )

  return (
    <NotificationContext.Provider
      value={{
        notifications,
        addNotification,
        markAsRead,
        markAllAsRead,
        deleteNotification,
        handleFriendAction,
      }}
    >
      {children}

      {/* Toast Notifications */}
      <div className="fixed top-4 right-4 z-50 space-y-2">
        {toastNotifications.map((notification) => (
          <ToastNotification
            key={notification.id}
            notification={notification}
            onClose={() => removeToastNotification(notification.id)}
            onAction={handleToastAction}
          />
        ))}
      </div>
    </NotificationContext.Provider>
  )
}
