'use client'

import React, { useState, useEffect } from 'react'
import { Bell, X, CheckCircle, AlertCircle, Info, Calendar } from 'lucide-react'
import { Notification } from '@/types'
import { formatDateTime } from '@/lib/utils'
import { cn } from '@/lib/utils'

interface NotificationDropdownProps {
  onClose: () => void
  updateUnreadCount: (count: number) => void
}

// ダミーデータ（実際はAPIから取得）
const dummyNotifications: Notification[] = [
  {
    id: '1',
    user_id: '1',
    type: 'TASK_ASSIGNED',
    title: 'タスクが割り当てられました',
    message: '「新機能の実装」が割り当てられました',
    status: 'SENT',
    created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '2',
    user_id: '1',
    type: 'TASK_DUE_SOON',
    title: 'タスクの期限が近づいています',
    message: '「バグ修正」の期限まで2時間です',
    status: 'SENT',
    created_at: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '3',
    user_id: '1',
    type: 'SYSTEM_NOTICE',
    title: 'システムメンテナンス',
    message: '明日午前2時からメンテナンスを実施します',
    status: 'READ',
    created_at: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
  },
]

export default function NotificationDropdown({ onClose, updateUnreadCount }: NotificationDropdownProps) {
  const [notifications, setNotifications] = useState<Notification[]>(dummyNotifications)
  const [loading, setLoading] = useState(false)

  // 通知アイコンを取得
  const getNotificationIcon = (type: Notification['type']) => {
    switch (type) {
      case 'TASK_ASSIGNED':
        return <CheckCircle className="h-5 w-5 text-blue-500" />
      case 'TASK_DUE_SOON':
        return <AlertCircle className="h-5 w-5 text-orange-500" />
      case 'TASK_COMPLETED':
        return <CheckCircle className="h-5 w-5 text-green-500" />
      case 'SYSTEM_NOTICE':
        return <Info className="h-5 w-5 text-purple-500" />
      default:
        return <Bell className="h-5 w-5 text-gray-500" />
    }
  }

  // 通知を既読にする
  const markAsRead = async (notificationId: string) => {
    try {
      // 実際はAPIを呼び出す
      // await markNotificationAsRead(notificationId)
      
      setNotifications(prev => 
        prev.map(notification => 
          notification.id === notificationId 
            ? { ...notification, status: 'READ' as const }
            : notification
        )
      )
      
      // 未読数を更新
      const unreadCount = notifications.filter(n => n.status !== 'READ').length - 1
      updateUnreadCount(Math.max(0, unreadCount))
    } catch (error) {
      console.error('Failed to mark notification as read:', error)
    }
  }

  // 全ての通知を既読にする
  const markAllAsRead = async () => {
    try {
      setLoading(true)
      
      // 実際はAPIを呼び出す
      // await markAllNotificationsAsRead()
      
      setNotifications(prev => 
        prev.map(notification => ({ ...notification, status: 'READ' as const }))
      )
      
      updateUnreadCount(0)
    } catch (error) {
      console.error('Failed to mark all notifications as read:', error)
    } finally {
      setLoading(false)
    }
  }

  // 通知削除
  const deleteNotification = async (notificationId: string) => {
    try {
      // 実際はAPIを呼び出す
      // await deleteNotification(notificationId)
      
      setNotifications(prev => prev.filter(n => n.id !== notificationId))
      
      // 未読数を更新
      const deletedNotification = notifications.find(n => n.id === notificationId)
      if (deletedNotification && deletedNotification.status !== 'READ') {
        const unreadCount = notifications.filter(n => n.status !== 'READ' && n.id !== notificationId).length
        updateUnreadCount(unreadCount)
      }
    } catch (error) {
      console.error('Failed to delete notification:', error)
    }
  }

  // 外側クリックで閉じる
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element
      if (!target.closest('[data-notification-dropdown]')) {
        onClose()
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [onClose])

  const unreadNotifications = notifications.filter(n => n.status !== 'READ')

  return (
    <div 
      data-notification-dropdown
      className="absolute right-0 mt-2 w-96 bg-white rounded-lg shadow-lg border border-gray-200 z-50 max-h-96 overflow-hidden"
    >
      {/* ヘッダー */}
      <div className="flex items-center justify-between p-4 border-b border-gray-200">
        <div className="flex items-center space-x-2">
          <Bell className="h-5 w-5 text-gray-600" />
          <h3 className="font-medium text-gray-900">通知</h3>
          {unreadNotifications.length > 0 && (
            <span className="bg-red-500 text-white text-xs rounded-full px-2 py-1">
              {unreadNotifications.length}
            </span>
          )}
        </div>
        <div className="flex items-center space-x-2">
          {unreadNotifications.length > 0 && (
            <button
              onClick={markAllAsRead}
              disabled={loading}
              className="text-sm text-blue-600 hover:text-blue-800 disabled:opacity-50"
            >
              すべて既読
            </button>
          )}
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
      </div>

      {/* 通知リスト */}
      <div className="max-h-80 overflow-y-auto">
        {notifications.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            <Bell className="h-12 w-12 mx-auto mb-4 text-gray-300" />
            <p>通知はありません</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-100">
            {notifications.map((notification) => (
              <div
                key={notification.id}
                className={cn(
                  "p-4 hover:bg-gray-50 transition-colors cursor-pointer relative group",
                  notification.status !== 'READ' && "bg-blue-50"
                )}
                onClick={() => {
                  if (notification.status !== 'READ') {
                    markAsRead(notification.id)
                  }
                }}
              >
                <div className="flex items-start space-x-3">
                  <div className="flex-shrink-0 mt-1">
                    {getNotificationIcon(notification.type)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <p className={cn(
                          "text-sm font-medium text-gray-900",
                          notification.status !== 'READ' && "font-semibold"
                        )}>
                          {notification.title}
                        </p>
                        <p className="text-sm text-gray-600 mt-1">
                          {notification.message}
                        </p>
                        <p className="text-xs text-gray-400 mt-2">
                          {formatDateTime(notification.created_at)}
                        </p>
                      </div>
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          deleteNotification(notification.id)
                        }}
                        className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-red-600 transition-opacity ml-2"
                      >
                        <X className="h-4 w-4" />
                      </button>
                    </div>
                    {notification.status !== 'READ' && (
                      <div className="absolute top-4 right-4 w-2 h-2 bg-blue-500 rounded-full"></div>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* フッター */}
      {notifications.length > 0 && (
        <div className="p-3 border-t border-gray-200 text-center">
          <button
            onClick={() => {
              onClose()
              // 通知ページへ遷移
              // router.push('/notifications')
            }}
            className="text-sm text-blue-600 hover:text-blue-800"
          >
            すべての通知を見る
          </button>
        </div>
      )}
    </div>
  )
}