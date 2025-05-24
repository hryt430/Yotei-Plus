import React from 'react';
import Link from 'next/link';
import { Edit, Trash2, Calendar, User, AlertCircle } from 'lucide-react';
import { Task, User as UserType } from '@/types';
import { 
  formatDate, 
  getRelativeDateLabel, 
  getStatusColor, 
  getPriorityColor, 
  getStatusLabel, 
  getPriorityLabel,
  cn 
} from '@/lib/utils';

interface TaskCardProps {
  task: Task;
  users?: UserType[];
  onEdit?: (task: Task) => void;
  onDelete?: (taskId: string) => void;
  onStatusChange?: (taskId: string, status: Task['status']) => void;
  compact?: boolean;
}

const TaskCard: React.FC<TaskCardProps> = ({ 
  task, 
  users = [],
  onEdit, 
  onDelete, 
  onStatusChange,
  compact = false 
}) => {
  const assignedUser = task.assignee_id 
    ? users.find(user => user.id === task.assignee_id)
    : undefined;
    
  const creator = users.find(user => user.id === task.created_by);
  
  const dueDateLabel = task.due_date ? getRelativeDateLabel(task.due_date) : '';
  const isOverdue = task.due_date && 
    new Date(task.due_date) < new Date() && 
    task.status !== 'DONE';

  const handleStatusChange = (newStatus: Task['status']) => {
    if (onStatusChange) {
      onStatusChange(task.id, newStatus);
    }
  };

  const handleEdit = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (onEdit) {
      onEdit(task);
    }
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (onDelete && confirm('このタスクを削除しますか？')) {
      onDelete(task.id);
    }
  };

  if (compact) {
    return (
      <div className="bg-white border border-gray-200 rounded-lg p-3 hover:shadow-md transition-shadow">
        <Link href={`/tasks/${task.id}`} className="block">
          <div className="flex items-center justify-between">
            <div className="flex-1 min-w-0">
              <h3 className="text-sm font-medium text-gray-900 truncate">
                {task.title}
              </h3>
              <div className="flex items-center space-x-2 mt-1">
                <span className={cn("text-xs px-2 py-1 rounded-full", getStatusColor(task.status))}>
                  {getStatusLabel(task.status)}
                </span>
                <span className={cn("text-xs px-2 py-1 rounded-full", getPriorityColor(task.priority))}>
                  {getPriorityLabel(task.priority)}
                </span>
              </div>
            </div>
            {task.due_date && (
              <div className={cn(
                "text-xs flex items-center ml-2",
                isOverdue ? "text-red-600 font-medium" : "text-gray-500"
              )}>
                <Calendar className="h-3 w-3 mr-1" />
                {dueDateLabel}
              </div>
            )}
          </div>
        </Link>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-shadow p-4 mb-3">
      <div className="flex items-start justify-between mb-3">
        <Link href={`/tasks/${task.id}`} className="flex-1 group">
          <h3 className="text-lg font-medium text-gray-900 group-hover:text-blue-600 transition-colors line-clamp-2">
            {task.title}
          </h3>
        </Link>
        
        <div className="flex items-center space-x-1 ml-3 opacity-0 group-hover:opacity-100 transition-opacity">
          {onEdit && (
            <button
              onClick={handleEdit}
              className="p-1 text-gray-400 hover:text-blue-600 rounded"
              title="編集"
            >
              <Edit className="h-4 w-4" />
            </button>
          )}
          
          {onDelete && (
            <button
              onClick={handleDelete}
              className="p-1 text-gray-400 hover:text-red-600 rounded"
              title="削除"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>
      
      {task.description && (
        <p className="text-gray-600 text-sm mb-3 line-clamp-2">
          {task.description}
        </p>
      )}
      
      {/* ステータスと優先度 */}
      <div className="flex flex-wrap items-center gap-2 mb-3">
        {onStatusChange ? (
          <select
            value={task.status}
            onChange={(e) => handleStatusChange(e.target.value as Task['status'])}
            className={cn(
              "text-xs px-2 py-1 rounded-full border cursor-pointer",
              getStatusColor(task.status)
            )}
          >
            <option value="TODO">未着手</option>
            <option value="IN_PROGRESS">進行中</option>
            <option value="DONE">完了</option>
          </select>
        ) : (
          <span className={cn("text-xs px-2 py-1 rounded-full border", getStatusColor(task.status))}>
            {getStatusLabel(task.status)}
          </span>
        )}
        
        <span className={cn("text-xs px-2 py-1 rounded-full border", getPriorityColor(task.priority))}>
          {getPriorityLabel(task.priority)}優先度
        </span>

        {isOverdue && (
          <span className="text-xs px-2 py-1 rounded-full bg-red-100 text-red-800 border border-red-200 flex items-center">
            <AlertCircle className="h-3 w-3 mr-1" />
            期限切れ
          </span>
        )}
      </div>
      
      {/* メタデータ */}
      <div className="flex items-center justify-between text-xs text-gray-500">
        <div className="flex items-center space-x-4">
          {assignedUser && (
            <div className="flex items-center">
              <User className="h-3 w-3 mr-1" />
              <span className="truncate max-w-24">{assignedUser.username}</span>
            </div>
          )}
          
          {creator && (
            <div className="flex items-center">
              <span>作成者: {creator.username}</span>
            </div>
          )}
        </div>
        
        {task.due_date && (
          <div className={cn(
            "flex items-center",
            isOverdue ? "text-red-600 font-medium" : "text-gray-500"
          )}>
            <Calendar className="h-3 w-3 mr-1" />
            <span>{dueDateLabel}</span>
          </div>
        )}
      </div>

      {/* 作成日・更新日 */}
      <div className="flex items-center justify-between text-xs text-gray-400 mt-2 pt-2 border-t border-gray-100">
        <span>作成: {formatDate(task.created_at)}</span>
        {task.updated_at !== task.created_at && (
          <span>更新: {formatDate(task.updated_at)}</span>
        )}
      </div>
    </div>
  );
};

export default TaskCard;