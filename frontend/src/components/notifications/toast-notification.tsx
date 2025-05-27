"use client"

import { useEffect, useState } from "react"
import { X, Users, Check, UserX } from "lucide-react"
import { Button } from "@/components/ui/button"
import type { Notification } from "@/types/notification"

interface ToastNotificationProps {
  notification: Notification
  onClose: () => void
  onAction?: (action: "accept" | "reject", notificationId: string) => void
}

export function ToastNotification({ notification, onClose, onAction }: ToastNotificationProps) {
  const [isVisible, setIsVisible] = useState(false)
  const [isLeaving, setIsLeaving] = useState(false)

  useEffect(() => {
    // Slide in animation
    setTimeout(() => setIsVisible(true), 100)

    // Auto-dismiss after 5 seconds for non-friend-request notifications
    if (notification.type !== "friend_request") {
      const timer = setTimeout(() => {
        handleClose()
      }, 5000)
      return () => clearTimeout(timer)
    }
  }, [notification.type])

  const handleClose = () => {
    setIsLeaving(true)
    setTimeout(() => {
      onClose()
    }, 300)
  }

  const handleAction = (action: "accept" | "reject") => {
    if (onAction) {
      onAction(action, notification.id)
    }
    handleClose()
  }

  const getIcon = () => {
    switch (notification.type) {
      case "friend_request":
        return <Users className="w-5 h-5 text-blue-600" />
      case "friend_accepted":
        return <Check className="w-5 h-5 text-green-600" />
      case "friend_rejected":
        return <UserX className="w-5 h-5 text-red-600" />
      default:
        return <Users className="w-5 h-5 text-gray-600" />
    }
  }

  const getBorderColor = () => {
    switch (notification.type) {
      case "friend_request":
        return "border-l-blue-500"
      case "friend_accepted":
        return "border-l-green-500"
      case "friend_rejected":
        return "border-l-red-500"
      default:
        return "border-l-gray-500"
    }
  }

  return (
    <div
      className={`fixed top-4 right-4 z-50 w-96 bg-white rounded-lg shadow-2xl border border-gray-200 ${getBorderColor()} border-l-4 transition-all duration-300 ${
        isVisible && !isLeaving ? "translate-x-0 opacity-100" : "translate-x-full opacity-0"
      }`}
    >
      <div className="p-4">
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center space-x-3">
            {getIcon()}
            <div>
              <h4 className="font-semibold text-gray-900 text-sm">{notification.title}</h4>
              <p className="text-sm text-gray-600 mt-1">{notification.message}</p>
            </div>
          </div>
          <Button variant="ghost" size="sm" onClick={handleClose} className="p-1 h-auto">
            <X className="w-4 h-4" />
          </Button>
        </div>

        {notification.type === "friend_request" && (
          <div className="flex space-x-2 mt-3">
            <Button
              size="sm"
              onClick={() => handleAction("accept")}
              className="flex-1 bg-blue-600 hover:bg-blue-700 text-white"
            >
              <Check className="w-3 h-3 mr-1" />
              Accept
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleAction("reject")}
              className="flex-1 border-gray-300 hover:bg-gray-50"
            >
              <X className="w-3 h-3 mr-1" />
              Decline
            </Button>
          </div>
        )}

        <div className="text-xs text-gray-400 mt-2">
          {notification.timestamp.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
        </div>
      </div>
    </div>
  )
}
