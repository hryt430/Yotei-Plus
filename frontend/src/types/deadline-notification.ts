export interface DeadlineNotification {
  id: string;
  taskId: string;
  taskTitle: string;
  taskDescription: string;
  dueDate: Date;
  priority: "LOW" | "MEDIUM" | "HIGH";
  category: string;
  timeUntilDue: string;
  urgencyLevel: "upcoming" | "due-soon" | "overdue";
  isRead: boolean;
  createdAt: Date;
}