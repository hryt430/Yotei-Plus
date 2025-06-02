"use client"

import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from "react"
import { toast } from "@/components/ui/hooks/use-toast"
import { useWebSocket } from "@/hooks/use-websocket"
import type { NotificationMessage, WebSocketMessage } from "@/types"

interface Notification {
  id: string
  type: 'friend_request' | 'friend_accepted' | 'friend_rejected' | 'task_update' | 'system' | 'info' | 'warning' | 'error'
  title: string
  message: string
  timestamp: Date
  isRead: boolean
  actionData?: {
    friendId?: string
    friendName?: string
    friendEmail?: string
    taskId?: string
    userId?: string
  }
}

interface NotificationContextType {
  notifications: Notification[]
  unreadCount: number
  addNotification: (notification: Omit<Notification, "id" | "timestamp" | "isRead">) => void
  markAsRead: (id: string) => void
  markAllAsRead: () => void
  deleteNotification: (id: string) => void
  clearAllNotifications: () => void
  handleFriendAction: (action: "accept" | "reject", notificationId: string) => void
  isConnected: boolean
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
  enableWebSocket?: boolean
  maxNotifications?: number
}

export function NotificationProvider({ 
  children, 
  enableWebSocket = true,
  maxNotifications = 50 
}: NotificationProviderProps) {
  const [notifications, setNotifications] = useState<Notification[]>([])
  
  // WebSocket接続
  const { subscribe, isConnected } = useWebSocket()

  // WebSocketからの通知を購読
  useEffect(() => {
    if (!enableWebSocket) return

    const unsubscribe = subscribe<NotificationMessage>(
      'notification',
      (message) => {
        const notification: Notification = {
          id: message.data.id,
          type: mapWebSocketTypeToNotificationType(message.data.type),
          title: message.data.title,
          message: message.data.message,
          timestamp: new Date(message.timestamp),
          isRead: message.data.status === 'READ',
          actionData: message.data.metadata
        }
        
        // ステートに追加
        addNotificationInternal(notification)
        
        // Toast表示（friend_request以外は自動表示）
        if (notification.type !== 'friend_request') {
          showToast(notification)
        }
      }
    )

    return unsubscribe
  }, [enableWebSocket, subscribe])

  // WebSocketタイプをNotificationタイプにマッピング
  const mapWebSocketTypeToNotificationType = (type: string): Notification['type'] => {
    switch (type) {
      case 'friend_request': return 'friend_request'
      case 'friend_accepted': return 'friend_accepted'
      case 'friend_rejected': return 'friend_rejected'
      case 'task_update': return 'task_update'
      case 'system': return 'system'
      case 'info': return 'info'
      case 'warning': return 'warning'
      case 'error': return 'error'
      default: return 'info'
    }
  }

  // 内部的な通知追加関数
  const addNotificationInternal = useCallback((notification: Notification) => {
    setNotifications((prev) => {
      const newNotifications = [notification, ...prev]
      // 最大件数制限
      return newNotifications.slice(0, maxNotifications)
    })
  }, [maxNotifications])

  // 外部から呼び出される通知追加関数
  const addNotification = useCallback((notificationData: Omit<Notification, "id" | "timestamp" | "isRead">) => {
    const newNotification: Notification = {
      ...notificationData,
      id: Date.now().toString() + Math.random().toString(36).substr(2, 9),
      timestamp: new Date(),
      isRead: false,
    }

    addNotificationInternal(newNotification)
    
    // Toast表示
    showToast(newNotification)
  }, [addNotificationInternal])

  // Toast表示関数
  const showToast = (notification: Notification) => {
    const variant = getToastVariant(notification.type)
    
    if (notification.type === 'friend_request') {
      // Friend requestは特別なToast
      toast({
        title: notification.title,
        description: notification.message,
        variant,
        action: (
          <div className="flex gap-2">
            <button
              className="inline-flex h-8 shrink-0 items-center justify-center rounded-md border bg-transparent px-3 text-sm font-medium ring-offset-background transition-colors hover:bg-secondary focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
              onClick={() => handleFriendAction('accept', notification.id)}
            >
              Accept
            </button>
            <button
              className="inline-flex h-8 shrink-0 items-center justify-center rounded-md border bg-transparent px-3 text-sm font-medium ring-offset-background transition-colors hover:bg-secondary focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
              onClick={() => handleFriendAction('reject', notification.id)}
            >
              Decline
            </button>
          </div>
        ),
      })
    } else {
      toast({
        title: notification.title,
        description: notification.message,
        variant,
      })
    }
  }

  // Toastのvariantを決定
  const getToastVariant = (type: Notification['type']): 'default' | 'destructive' => {
    switch (type) {
      case 'error':
      case 'friend_rejected':
        return 'destructive'
      default:
        return 'default'
    }
  }

  const markAsRead = useCallback((id: string) => {
    setNotifications((prev) => 
      prev.map((n) => (n.id === id ? { ...n, isRead: true } : n))
    )
  }, [])

  const markAllAsRead = useCallback(() => {
    setNotifications((prev) => prev.map((n) => ({ ...n, isRead: true })))
  }, [])

  const deleteNotification = useCallback((id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id))
  }, [])

  const clearAllNotifications = useCallback(() => {
    setNotifications([])
  }, [])

  const handleFriendAction = useCallback(
    async (action: "accept" | "reject", notificationId: string) => {
      const notification = notifications.find((n) => n.id === notificationId)
      if (!notification || !notification.actionData) return

      try {
        // API呼び出し（実際のAPIエンドポイントに置き換える）
        const response = await fetch(`/api/friends/${action}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            friendId: notification.actionData.friendId,
            notificationId,
          }),
        })

        if (!response.ok) {
          throw new Error(`Failed to ${action} friend request`)
        }

        // 元の通知を削除
        deleteNotification(notificationId)

        // 確認通知を追加
        const confirmationNotification: Omit<Notification, "id" | "timestamp" | "isRead"> = {
          type: action === "accept" ? "friend_accepted" : "friend_rejected",
          title: action === "accept" ? "Friend Request Accepted" : "Friend Request Declined",
          message:
            action === "accept"
              ? `You are now friends with ${notification.actionData.friendName}`
              : `You declined the friend request from ${notification.actionData.friendName}`,
        }

        addNotification(confirmationNotification)

      } catch (error) {
        console.error(`Error ${action}ing friend request:`, error)
        
        // エラートースト表示
        toast({
          title: "Error",
          description: `Failed to ${action} friend request. Please try again.`,
          variant: "destructive",
        })
      }
    },
    [notifications, deleteNotification, addNotification]
  )

  // 未読数を計算
  const unreadCount = notifications.filter((n) => !n.isRead).length

  return (
    <NotificationContext.Provider
      value={{
        notifications,
        unreadCount,
        addNotification,
        markAsRead,
        markAllAsRead,
        deleteNotification,
        clearAllNotifications,
        handleFriendAction,
        isConnected,
      }}
    >
      {children}
    </NotificationContext.Provider>
  )
}