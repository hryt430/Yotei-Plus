import { useEffect, useRef, useState, useCallback } from 'react'
import { WebSocketMessage, NotificationMessage, TaskUpdateMessage } from '@/types'
import { TokenManager } from '@/api/client'

interface WebSocketConfig {
  url?: string
  autoReconnect?: boolean
  reconnectInterval?: number
  maxReconnectAttempts?: number
  pingInterval?: number
}

interface WebSocketState {
  isConnected: boolean
  isConnecting: boolean
  error: string | null
  reconnectAttempts: number
}

type MessageHandler<T = any> = (message: T) => void

const DEFAULT_CONFIG: Required<WebSocketConfig> = {
  url: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws/notifications',
  autoReconnect: true,
  reconnectInterval: 5000,
  maxReconnectAttempts: 10,
  pingInterval: 30000
}

export function useWebSocket(config: WebSocketConfig = {}) {
  const finalConfig = { ...DEFAULT_CONFIG, ...config }
  
  const [state, setState] = useState<WebSocketState>({
    isConnected: false,
    isConnecting: false,
    error: null,
    reconnectAttempts: 0
  })

  const websocketRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const pingIntervalRef = useRef<NodeJS.Timeout | null>(null)
  const messageHandlersRef = useRef<Map<string, Set<MessageHandler>>>(new Map())

  // メッセージハンドラーを登録
  const subscribe = useCallback(<T = any>(
    messageType: string, 
    handler: MessageHandler<T>
  ): (() => void) => {
    if (!messageHandlersRef.current.has(messageType)) {
      messageHandlersRef.current.set(messageType, new Set())
    }
    
    messageHandlersRef.current.get(messageType)!.add(handler)
    
    // アンサブスクライブ関数を返す
    return () => {
      const handlers = messageHandlersRef.current.get(messageType)
      if (handlers) {
        handlers.delete(handler)
        if (handlers.size === 0) {
          messageHandlersRef.current.delete(messageType)
        }
      }
    }
  }, [])

  // メッセージを送信
  const sendMessage = useCallback((message: any) => {
    if (websocketRef.current?.readyState === WebSocket.OPEN) {
      websocketRef.current.send(JSON.stringify(message))
      return true
    }
    return false
  }, [])

  // Ping送信
  const sendPing = useCallback(() => {
    sendMessage({ type: 'ping', timestamp: new Date().toISOString() })
  }, [sendMessage])

  // WebSocket接続
  const connect = useCallback(() => {
    if (websocketRef.current?.readyState === WebSocket.CONNECTING || 
        websocketRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    setState(prev => ({ ...prev, isConnecting: true, error: null }))

    try {
      // 認証トークンを取得
      const token = TokenManager.getAccessToken()
      if (!token) {
        setState(prev => ({ 
          ...prev, 
          isConnecting: false, 
          error: '認証トークンが見つかりません' 
        }))
        return
      }

      // WebSocket URLにトークンを追加
      const wsUrl = `${finalConfig.url}?token=${encodeURIComponent(token)}`
      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        console.log('WebSocket connected')
        setState(prev => ({ 
          ...prev, 
          isConnected: true, 
          isConnecting: false, 
          error: null,
          reconnectAttempts: 0
        }))

        // Ping開始
        if (pingIntervalRef.current) {
          clearInterval(pingIntervalRef.current)
        }
        pingIntervalRef.current = setInterval(sendPing, finalConfig.pingInterval)
      }

      ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason)
        setState(prev => ({ 
          ...prev, 
          isConnected: false, 
          isConnecting: false 
        }))

        // Ping停止
        if (pingIntervalRef.current) {
          clearInterval(pingIntervalRef.current)
          pingIntervalRef.current = null
        }

        // 自動再接続
        if (finalConfig.autoReconnect && 
            event.code !== 1000 && // 正常終了以外
            state.reconnectAttempts < finalConfig.maxReconnectAttempts) {
          
          setState(prev => ({ 
            ...prev, 
            reconnectAttempts: prev.reconnectAttempts + 1 
          }))

          reconnectTimeoutRef.current = setTimeout(() => {
            connect()
          }, finalConfig.reconnectInterval)
        }
      }

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        setState(prev => ({ 
          ...prev, 
          error: 'WebSocket接続エラーが発生しました',
          isConnecting: false
        }))
      }

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          
          // Todo: Pong応答の処理
        //   if (message.type === 'pong') {
        //     return
        //   }

          // メッセージタイプに応じたハンドラーを実行
          const handlers = messageHandlersRef.current.get(message.type)
          if (handlers) {
            handlers.forEach(handler => {
              try {
                handler(message)
              } catch (error) {
                console.error('Message handler error:', error)
              }
            })
          }

          // 全メッセージハンドラーも実行
          const allHandlers = messageHandlersRef.current.get('*')
          if (allHandlers) {
            allHandlers.forEach(handler => {
              try {
                handler(message)
              } catch (error) {
                console.error('All message handler error:', error)
              }
            })
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      websocketRef.current = ws
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error)
      setState(prev => ({ 
        ...prev, 
        isConnecting: false, 
        error: 'WebSocket接続の作成に失敗しました' 
      }))
    }
  }, [finalConfig, state.reconnectAttempts, sendPing])

  // WebSocket切断
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }

    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current)
      pingIntervalRef.current = null
    }

    if (websocketRef.current) {
      websocketRef.current.close(1000, 'Manual disconnect')
      websocketRef.current = null
    }

    setState({
      isConnected: false,
      isConnecting: false,
      error: null,
      reconnectAttempts: 0
    })
  }, [])

  // 手動再接続
  const reconnect = useCallback(() => {
    disconnect()
    setState(prev => ({ ...prev, reconnectAttempts: 0 }))
    connect()
  }, [disconnect, connect])

  // コンポーネントマウント時に接続
  useEffect(() => {
    connect()
    
    // クリーンアップ
    return () => {
      disconnect()
    }
  }, []) // connectとdisconnectは含めない（無限ループ防止）

  // トークン変更時の再接続
  useEffect(() => {
    const token = TokenManager.getAccessToken()
    if (token && !state.isConnected && !state.isConnecting) {
      connect()
    } else if (!token && state.isConnected) {
      disconnect()
    }
  }, [TokenManager.getAccessToken()]) // トークンの変更を検知

  return {
    // 状態
    isConnected: state.isConnected,
    isConnecting: state.isConnecting,
    error: state.error,
    reconnectAttempts: state.reconnectAttempts,
    
    // アクション
    connect,
    disconnect,
    reconnect,
    sendMessage,
    subscribe
  }
}

// 通知専用フック
export function useNotificationWebSocket() {
  const websocket = useWebSocket()
  const [notifications, setNotifications] = useState<NotificationMessage[]>([])
  const [unreadCount, setUnreadCount] = useState(0)

  // 通知メッセージの購読
  useEffect(() => {
    const unsubscribe = websocket.subscribe<NotificationMessage>(
      'notification',
      (message) => {
        setNotifications(prev => [message, ...prev.slice(0, 49)]) // 最新50件まで保持
        
        // 未読カウントを増加（READステータス以外）
        if (message.data.status !== 'READ') {
          setUnreadCount(prev => prev + 1)
        }
      }
    )

    return unsubscribe
  }, [websocket.subscribe])

  // 通知を既読にする
  const markAsRead = useCallback((notificationId: string) => {
    setNotifications(prev => 
      prev.map(notification => 
        notification.data.id === notificationId
          ? { ...notification, data: { ...notification.data, status: 'READ' as const } }
          : notification
      )
    )
    setUnreadCount(prev => Math.max(0, prev - 1))
  }, [])

  // 全通知をクリア
  const clearNotifications = useCallback(() => {
    setNotifications([])
    setUnreadCount(0)
  }, [])

  return {
    ...websocket,
    notifications,
    unreadCount,
    markAsRead,
    clearNotifications
  }
}

// タスク更新専用フック
export function useTaskWebSocket() {
  const websocket = useWebSocket()
  const [taskUpdates, setTaskUpdates] = useState<TaskUpdateMessage[]>([])

  // タスク更新メッセージの購読
  useEffect(() => {
    const unsubscribe = websocket.subscribe<TaskUpdateMessage>(
      'task_update',
      (message) => {
        setTaskUpdates(prev => [message, ...prev.slice(0, 99)]) // 最新100件まで保持
      }
    )

    return unsubscribe
  }, [websocket.subscribe])

  // タスク更新履歴をクリア
  const clearTaskUpdates = useCallback(() => {
    setTaskUpdates([])
  }, [])

  return {
    ...websocket,
    taskUpdates,
    clearTaskUpdates
  }
}