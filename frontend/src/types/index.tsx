export interface User {
  id: string;
  name: string;
  email: string;
  avatar?: string;
}

export interface Task {
  id: string;
  title: string;
  description?: string;
  status: 'TODO' | 'IN_PROGRESS' | 'DONE';
  priority: 'LOW' | 'MEDIUM' | 'HIGH';
  dueDate?: string;
  createdAt: string;
  updatedAt: string;
  assignedTo?: User;
  assignedBy?: User;
  tags?: string[];
}

export interface Notification {
  id: string;
  type: 'TASK_ASSIGNED' | 'TASK_UPDATED' | 'TASK_DUE_SOON' | 'SYSTEM';
  message: string;
  read: boolean;
  createdAt: string;
  taskId?: string;
  userId: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

export interface TasksState {
  tasks: Task[];
  isLoading: boolean;
  error: string | null;
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
  filters: {
    status?: string[];
    priority?: string[];
    assignedTo?: string;
    search?: string;
    startDate?: string;
    endDate?: string;
    tags?: string[];
  };
  sort: {
    field: string;
    direction: 'asc' | 'desc';
  };
}

export interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  isLoading: boolean;
  error: string | null;
}