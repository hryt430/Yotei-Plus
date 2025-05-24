import React from 'react';
import Link from 'next/link';
import { Task } from '@/types';
import { formatDate, getRelativeDateLabel, isOverdue } from '@/lib/utils/date-utils';

interface TaskCardProps {
  task: Task;
  onEdit?: (task: Task) => void;
  onDelete?: (taskId: string) => void;
}

const priorityClasses = {
  LOW: 'bg-blue-50 text-blue-600 border-blue-200',
  MEDIUM: 'bg-yellow-50 text-yellow-600 border-yellow-200',
  HIGH: 'bg-red-50 text-red-600 border-red-200',
};

const statusClasses = {
  TODO: 'bg-gray-50 text-gray-600 border-gray-200',
  IN_PROGRESS: 'bg-indigo-50 text-indigo-600 border-indigo-200',
  DONE: 'bg-green-50 text-green-600 border-green-200',
};

const TaskCard: React.FC<TaskCardProps> = ({ task, onEdit, onDelete }) => {
  const dueDateLabel = task.dueDate ? getRelativeDateLabel(task.dueDate) : '';
  const isTaskOverdue = task.dueDate && isOverdue(task.dueDate) && task.status !== 'DONE';

  return (
    <div className="bg-white rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-shadow p-4 mb-3">
      <div className="flex items-start justify-between mb-2">
        <Link href={`/tasks/${task.id}`} className="text-lg font-medium text-gray-900 hover:text-blue-600 transition-colors">
          {task.title}
        </Link>
        
        <div className="flex space-x-2">
          {onEdit && (
            <button
              onClick={() => onEdit(task)}
              className="text-gray-400 hover:text-gray-700"
              aria-label="タスクを編集"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zm-2.207 2.207L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
              </svg>
            </button>
          )}
          
          {onDelete && (
            <button
              onClick={() => onDelete(task.id)}
              className="text-gray-400 hover:text-red-600"
              aria-label="タスクを削除"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
            </button>
          )}
        </div>
      </div>
      
      {task.description && (
        <p className="text-gray-600 text-sm mb-3 line-clamp-2">{task.description}</p>
      )}
      
      <div className="flex flex-wrap items-center gap-2 mb-3">
        <span className={`text-xs px-2 py-1 rounded-full border ${statusClasses[task.status]}`}>
          {task.status === 'TODO' ? '未着手' : task.status === 'IN_PROGRESS' ? '進行中' : '完了'}
        </span>
        
        <span className={`text-xs px-2 py-1 rounded-full border ${priorityClasses[task.priority]}`}>
          {task.priority === 'LOW' ? '低' : task.priority === 'MEDIUM' ? '中' : '高'}優先度
        </span>
        
        {task.tags && task.tags.map(tag => (
          <span key={tag} className="text-xs px-2 py-1 rounded-full bg-gray-100 text-gray-600">
            {tag}
          </span>
        ))}
      </div>
      
      <div className="flex items-center justify-between text-xs text-gray-500">
        <div className="flex items-center">
          {task.assignedTo && (
            <div className="flex items-center mr-3">
              <span className="w-6 h-6 rounded-full bg-gray-200 flex items-center justify-center text-xs font-medium mr-1">
                {task.assignedTo.name.charAt(0)}
              </span>
              <span>{task.assignedTo.name}</span>
            </div>
          )}
          
          <span>作成: {formatDate(task.createdAt)}</span>
        </div>
        
        {task.dueDate && (
          <div className={`flex items-center ${isTaskOverdue ? 'text-red-600 font-medium' : ''}`}>
            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <span>{dueDateLabel}</span>
          </div>
        )}
      </div>
    </div>
  );
};

export default TaskCard;