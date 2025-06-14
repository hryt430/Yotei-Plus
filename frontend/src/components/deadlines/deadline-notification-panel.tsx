"use client"

import { useState } from "react"
import { Button } from "@/components/ui/forms/button"
import { Badge } from "@/components/ui/data-display/badge"
import { ScrollArea } from "@/components/ui/layout/scroll-area"
import {
  Bell,
  Clock,
  AlertTriangle,
  Calendar,
  Flag,
  X,
  CheckCircle,
  AlarmClockIcon as Snooze,
  Trash2,
  Settings,
} from "lucide-react"
import type { DeadlineNotification } from "@/types"

interface DeadlineNotificationPanelProps {
  notifications: DeadlineNotification[]
  onMarkComplete: (taskId: string) => void
  onSnooze: (taskId: string, minutes: number) => void
  onDismiss: (notificationId: string) => void
  onMarkAsRead: (notificationId: string) => void
  onClearAll: () => void
}

export function DeadlineNotificationPanel({
  notifications,
  onMarkComplete,
  onSnooze,
  onDismiss,
  onMarkAsRead,
  onClearAll,
}: DeadlineNotificationPanelProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [filter, setFilter] = useState<"all" | "overdue" | "due-soon" | "upcoming">("all")

  const unreadCount = notifications.filter((n) => !n.isRead).length
  const overdueCount = notifications.filter((n) => n.urgencyLevel === "overdue").length

  const filteredNotifications = notifications.filter((notification) => {
    if (filter === "all") return true
    return notification.urgencyLevel === filter
  })

  const getUrgencyIcon = (urgencyLevel: string) => {
    switch (urgencyLevel) {
      case "overdue":
        return <AlertTriangle className="w-4 h-4 text-red-600" />
      case "due-soon":
        return <Clock className="w-4 h-4 text-orange-600" />
      case "upcoming":
        return <Calendar className="w-4 h-4 text-blue-600" />
      default:
        return <Clock className="w-4 h-4 text-gray-600" />
    }
  }

  const getUrgencyBadge = (urgencyLevel: string) => {
    switch (urgencyLevel) {
      case "overdue":
        return "bg-red-100 text-red-800 border-red-200"
      case "due-soon":
        return "bg-orange-100 text-orange-800 border-orange-200"
      case "upcoming":
        return "bg-blue-100 text-blue-800 border-blue-200"
      default:
        return "bg-gray-100 text-gray-800 border-gray-200"
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

  return (
    <div className="relative">
      {/* Bell Icon Button */}
      <Button variant="ghost" size="sm" onClick={() => setIsOpen(!isOpen)} className="relative p-2 hover:bg-gray-100">
        <Bell className="w-5 h-5 text-gray-600" />
        {unreadCount > 0 && (
          <Badge
            className={`absolute -top-1 -right-1 h-5 w-5 p-0 text-xs ${
              overdueCount > 0 ? "bg-red-500 hover:bg-red-500" : "bg-blue-500 hover:bg-blue-500"
            }`}
          >
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
              <div className="flex items-center justify-between mb-3">
                <h3 className="font-semibold text-gray-900 flex items-center">
                  <Clock className="w-4 h-4 mr-2" />
                  Deadline Alerts
                </h3>
                <div className="flex items-center space-x-2">
                  <Button variant="ghost" size="sm" onClick={() => setIsOpen(false)}>
                    <X className="w-4 h-4" />
                  </Button>
                </div>
              </div>

              {/* Filter Tabs */}
              <div className="flex space-x-1 bg-gray-100 rounded-lg p-1">
                {[
                  { key: "all", label: "All", count: notifications.length },
                  {
                    key: "overdue",
                    label: "Overdue",
                    count: notifications.filter((n) => n.urgencyLevel === "overdue").length,
                  },
                  {
                    key: "due-soon",
                    label: "Soon",
                    count: notifications.filter((n) => n.urgencyLevel === "due-soon").length,
                  },
                  {
                    key: "upcoming",
                    label: "Later",
                    count: notifications.filter((n) => n.urgencyLevel === "upcoming").length,
                  },
                ].map((tab) => (
                  <button
                    key={tab.key}
                    onClick={() => setFilter(tab.key as any)}
                    className={`flex-1 text-xs py-1 px-2 rounded transition-colors ${
                      filter === tab.key ? "bg-white text-gray-900 shadow-sm" : "text-gray-600 hover:text-gray-900"
                    }`}
                  >
                    {tab.label} {tab.count > 0 && `(${tab.count})`}
                  </button>
                ))}
              </div>
            </div>

            {/* Notifications List */}
            <ScrollArea className="max-h-96">
              {filteredNotifications.length === 0 ? (
                <div className="p-8 text-center">
                  <Clock className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                  <p className="text-gray-500">No deadline alerts</p>
                  <p className="text-sm text-gray-400">You're all caught up!</p>
                </div>
              ) : (
                <div className="divide-y divide-gray-100">
                  {filteredNotifications.map((notification) => (
                    <div
                      key={notification.id}
                      className={`p-4 hover:bg-gray-50 transition-colors ${
                        !notification.isRead ? "bg-blue-50/30" : ""
                      }`}
                      onClick={() => !notification.isRead && onMarkAsRead(notification.id)}
                    >
                      <div className="flex items-start space-x-3">
                        <div className="flex-shrink-0 mt-1">{getUrgencyIcon(notification.urgencyLevel)}</div>

                        <div className="flex-1 min-w-0">
                          <div className="flex items-start justify-between mb-2">
                            <div className="flex-1">
                              <div className="flex items-center space-x-2 mb-1">
                                <Badge
                                  variant="outline"
                                  className={`text-xs ${getUrgencyBadge(notification.urgencyLevel)}`}
                                >
                                  {notification.urgencyLevel === "overdue"
                                    ? "Overdue"
                                    : notification.urgencyLevel === "due-soon"
                                      ? "Due Soon"
                                      : "Upcoming"}
                                </Badge>
                                <Badge
                                  variant="outline"
                                  className={`text-xs ${getPriorityColor(notification.priority)}`}
                                >
                                  <Flag className="w-2 h-2 mr-1" />
                                  {notification.priority}
                                </Badge>
                              </div>
                              <h4
                                className={`text-sm font-medium ${
                                  !notification.isRead ? "text-gray-900" : "text-gray-700"
                                }`}
                              >
                                {notification.taskTitle}
                              </h4>
                              <p className="text-xs text-gray-600 mt-1 line-clamp-2">{notification.taskDescription}</p>
                              <div className="flex items-center mt-2 text-xs text-gray-500">
                                <Calendar className="w-3 h-3 mr-1" />
                                <span>{notification.timeUntilDue}</span>
                              </div>
                            </div>

                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={(e) => {
                                e.stopPropagation()
                                onDismiss(notification.id)
                              }}
                              className="p-1 h-auto text-gray-400 hover:text-red-600"
                            >
                              <Trash2 className="w-3 h-3" />
                            </Button>
                          </div>

                          {/* Action Buttons */}
                          <div className="flex space-x-2 mt-3">
                            <Button
                              size="sm"
                              onClick={(e) => {
                                e.stopPropagation()
                                onMarkComplete(notification.taskId)
                              }}
                              className="bg-green-600 hover:bg-green-700 text-white text-xs px-3 py-1"
                            >
                              <CheckCircle className="w-3 h-3 mr-1" />
                              Complete
                            </Button>
                            {notification.urgencyLevel !== "overdue" && (
                              <>
                                <Button
                                  size="sm"
                                  variant="outline"
                                  onClick={(e) => {
                                    e.stopPropagation()
                                    onSnooze(notification.taskId, 15)
                                  }}
                                  className="text-xs px-2 py-1 border-gray-300 hover:bg-gray-50"
                                >
                                  <Snooze className="w-3 h-3 mr-1" />
                                  15m
                                </Button>
                                <Button
                                  size="sm"
                                  variant="outline"
                                  onClick={(e) => {
                                    e.stopPropagation()
                                    onSnooze(notification.taskId, 60)
                                  }}
                                  className="text-xs px-2 py-1 border-gray-300 hover:bg-gray-50"
                                >
                                  1h
                                </Button>
                              </>
                            )}
                          </div>
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

            {/* Footer */}
            {filteredNotifications.length > 0 && (
              <div className="p-3 border-t border-gray-200 bg-gray-50/30">
                <div className="flex justify-between items-center">
                  <Button variant="ghost" size="sm" className="text-xs">
                    <Settings className="w-3 h-3 mr-1" />
                    Settings
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={onClearAll}
                    className="text-xs text-red-600 hover:text-red-700"
                  >
                    Clear All
                  </Button>
                </div>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}

export default DeadlineNotificationPanel
