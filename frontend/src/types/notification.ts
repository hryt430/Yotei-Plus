// === Notification Types ===
export type NotificationType = 
  | 'APP_NOTIFICATION' 
  | 'TASK_ASSIGNED' 
  | 'TASK_COMPLETED' 
  | 'TASK_DUE_SOON' 
  | 'SYSTEM_NOTICE'
  | 'FRIEND_REQUEST'
  | 'FRIEND_ACCEPTED'
  | 'GROUP_INVITATION'
  | 'GROUP_MEMBER_ADDED';

export type NotificationStatus = 'PENDING' | 'SENT' | 'READ' | 'FAILED';

export interface Notification {
  id: string;
  user_id: string;
  type: NotificationType;
  title: string;
  message: string;
  status: NotificationStatus; 
  metadata?: Record<string, string>;
  created_at: string;
  updated_at: string;
  sent_at?: string;
}

// === Notification State Types ===
export interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  isLoading: boolean;
  error: string | null;
}

// === WebSocket Types ===
export interface WebSocketMessage {
  type: 'notification' | 'task_update' | 'user_update' | 'social_update' | 'group_update';
  data: any;
  timestamp: string;
}

export interface NotificationMessage extends WebSocketMessage {
  type: 'notification';
  data: Notification;
}