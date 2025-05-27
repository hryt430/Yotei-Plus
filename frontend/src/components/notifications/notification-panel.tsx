"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Bell, Check, X, Users, UserX, Trash2, BookMarkedIcon as MarkAsRead } from "lucide-react"
import type { Notification } from "@/types/notification"

interface NotificationPanelProps {
  notifications: Notification[]
  onAction: (action: "accept" | "reject", notificationId: string) => void
  onMarkAsRead: (notificationId: string) => void
  onMarkAllAsRead: () => void
  onDelete: (notificationId: string) => void
}

export function NotificationPanel({
  notifications,
  onAction,
  onMarkAsRead,
  onMarkAllAsRead,
  onDelete,
}: NotificationPanelProps) {
  const [isOpen, setIsOpen] = useState(false)

  const unreadCount = notifications.filter((n) => !n.isRead).length

  const getIcon = (type: string) => {
    switch (type) {
      case "friend_request":
        return <Users className="w-4 h-4 text-blue-600" />
      case "friend_accepted":
        return <Check className="w-4 h-4 text-green-600" />
      case "friend_rejected":
        return <UserX className="w-4 h-4 text-red-600" />
      default:
        return <Users className="w-4 h-4 text-gray-600" />
    }
  }

  const formatTime = (date: Date) => {
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(diff / 3600000)
    const days = Math.floor(diff / 86400000)

    if (minutes < 1) return "Just now"
    if (minutes < 60) return `${minutes}m ago`
    if (hours < 24) return `${hours}h ago`
    return `${days}d ago`
  }

  return (
    <div className="relative">
      {/* Bell Icon Button */}
      <Button variant="ghost" size="sm" onClick={() => setIsOpen(!isOpen)} className="relative p-2 hover:bg-gray-100">
        <Bell className="w-5 h-5 text-gray-600" />
        {unreadCount > 0 && (
          <Badge className="absolute -top-1 -right-1 h-5 w-5 p-0 text-xs bg-red-500 hover:bg-red-500">
            {unreadCount > 9 ? "9+" : unreadCount}
          </Badge>
        )}
      </Button>

      {/* Notification Dropdown Panel */}
      {isOpen && (
        <>
          {/* Backdrop */}
          <div className="fixed inset-0 z-40" onClick={() => setIsOpen(false)} />

          {/* Panel */}
          <div className="absolute right-0 top-full mt-2 w-96 bg-white rounded-lg shadow-2xl border border-gray-200 z-50">
            {/* Header */}
            <div className="p-4 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold text-gray-900">Notifications</h3>
                <div className="flex items-center space-x-2">
                  {unreadCount > 0 && (
                    <Button variant="ghost" size="sm" onClick={onMarkAllAsRead} className="text-xs">
                      <MarkAsRead className="w-3 h-3 mr-1" />
                      Mark all read
                    </Button>
                  )}
                  <Button variant="ghost" size="sm" onClick={() => setIsOpen(false)}>
                    <X className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </div>

            {/* Notifications List */}
            <ScrollArea className="max-h-96">
              {notifications.length === 0 ? (
                <div className="p-8 text-center">
                  <Bell className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                  <p className="text-gray-500">No notifications yet</p>
                  <p className="text-sm text-gray-400">You'll see friend requests and updates here</p>
                </div>
              ) : (
                <div className="divide-y divide-gray-100">
                  {notifications.map((notification) => (
                    <div
                      key={notification.id}
                      className={`p-4 hover:bg-gray-50 transition-colors ${
                        !notification.isRead ? "bg-blue-50/30" : ""
                      }`}
                    >
                      <div className="flex items-start space-x-3">
                        <div className="flex-shrink-0 mt-1">{getIcon(notification.type)}</div>

                        <div className="flex-1 min-w-0">
                          <div className="flex items-start justify-between">
                            <div className="flex-1">
                              <h4
                                className={`text-sm font-medium ${
                                  !notification.isRead ? "text-gray-900" : "text-gray-700"
                                }`}
                              >
                                {notification.title}
                              </h4>
                              <p className="text-sm text-gray-600 mt-1">{notification.message}</p>
                              <p className="text-xs text-gray-400 mt-2">{formatTime(notification.timestamp)}</p>
                            </div>

                            <div className="flex items-center space-x-1 ml-2">
                              {!notification.isRead && (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => onMarkAsRead(notification.id)}
                                  className="p-1 h-auto"
                                >
                                  <MarkAsRead className="w-3 h-3" />
                                </Button>
                              )}
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => onDelete(notification.id)}
                                className="p-1 h-auto text-gray-400 hover:text-red-600"
                              >
                                <Trash2 className="w-3 h-3" />
                              </Button>
                            </div>
                          </div>

                          {/* Action Buttons for Friend Requests */}
                          {notification.type === "friend_request" && (
                            <div className="flex space-x-2 mt-3">
                              <Button
                                size="sm"
                                onClick={() => onAction("accept", notification.id)}
                                className="bg-blue-600 hover:bg-blue-700 text-white text-xs px-3 py-1"
                              >
                                <Check className="w-3 h-3 mr-1" />
                                Accept
                              </Button>
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => onAction("reject", notification.id)}
                                className="border-gray-300 hover:bg-gray-50 text-xs px-3 py-1"
                              >
                                <X className="w-3 h-3 mr-1" />
                                Decline
                              </Button>
                            </div>
                          )}
                        </div>
                      </div>

                      {!notification.isRead && (
                        <div className="absolute left-2 top-1/2 transform -translate-y-1/2 w-2 h-2 bg-blue-500 rounded-full"></div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>
          </div>
        </>
      )}
    </div>
  )
}
