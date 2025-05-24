import React, { useState, useEffect } from 'react';
import { Calendar, User, Plus, X } from 'lucide-react';
import { Task, User as UserType, TaskFormData } from '@/types';
import { cn } from '@/lib/utils';

interface TaskFormProps {
  task?: Task;
  users?: UserType[];
  onSubmit: (taskData: TaskFormData) => Promise<void>;
  onCancel: () => void;
  isLoading?: boolean;
}

const TaskForm: React.FC<TaskFormProps> = ({ 
  task, 
  users = [], 
  onSubmit, 
  onCancel,
  isLoading = false 
}) => {
  const [formData, setFormData] = useState<TaskFormData>({
    title: '',
    description: '',
    status: 'TODO',
    priority: 'MEDIUM',
    assignee_id: undefined,
    due_date: undefined,
  });
  
  const [errors, setErrors] = useState<Partial<Record<keyof TaskFormData, string>>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // 編集モードの場合、タスクデータで初期化
  useEffect(() => {
    if (task) {
      setFormData({
        title: task.title,
        description: task.description || '',
        status: task.status,
        priority: task.priority,
        assignee_id: task.assignee_id,
        due_date: task.due_date ? task.due_date.split('T')[0] : undefined, // 日付部分のみ
      });
    }
  }, [task]);
  
  // フォームフィールドの変更ハンドラー
  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    
    setFormData(prev => ({
      ...prev,
      [name]: value === '' ? undefined : value,
    }));
    
    // エラーをクリア
    if (errors[name as keyof TaskFormData]) {
      setErrors(prev => ({
        ...prev,
        [name]: undefined,
      }));
    }
  };
  
  // バリデーション
  const validateForm = (): boolean => {
    const newErrors: typeof errors = {};
    
    if (!formData.title.trim()) {
      newErrors.title = 'タイトルは必須です';
    } else if (formData.title.length > 100) {
      newErrors.title = 'タイトルは100文字以内で入力してください';
    }
    
    if (formData.description && formData.description.length > 1000) {
      newErrors.description = '説明は1000文字以内で入力してください';
    }
    
    if (formData.due_date) {
      const dueDate = new Date(formData.due_date);
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      
      if (dueDate < today) {
        newErrors.due_date = '期限は今日以降の日付を設定してください';
      }
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };
  
  // フォーム送信
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }
    
    setIsSubmitting(true);
    
    try {
      // 空文字列をundefinedに変換
      const submitData: TaskFormData = {
        ...formData,
        title: formData.title.trim(),
        description: formData.description?.trim() || '',
        due_date: formData.due_date || undefined,
        assignee_id: formData.assignee_id || undefined,
      };
      
      await onSubmit(submitData);
    } catch (error) {
      console.error('Form submission error:', error);
    } finally {
      setIsSubmitting(false);
    }
  };
  
  // 今日の日付を取得（YYYY-MM-DD形式）
  const getTodayDate = () => {
    return new Date().toISOString().split('T')[0];
  };
  
  const isEditMode = !!task;
  
  return (
    <div className="bg-white rounded-lg shadow-md p-6 max-w-2xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-900">
          {isEditMode ? 'タスクを編集' : '新しいタスク'}
        </h2>
        <button
          onClick={onCancel}
          className="text-gray-400 hover:text-gray-600 p-1 rounded"
          disabled={isSubmitting}
        >
          <X className="h-5 w-5" />
        </button>
      </div>
      
      <form onSubmit={handleSubmit} className="space-y-6">
        {/* タイトル */}
        <div>
          <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-1">
            タイトル *
          </label>
          <input
            type="text"
            id="title"
            name="title"
            value={formData.title}
            onChange={handleChange}
            className={cn(
              "w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500",
              errors.title ? "border-red-300" : "border-gray-300"
            )}
            placeholder="タスクのタイトルを入力..."
            disabled={isSubmitting}
            maxLength={100}
          />
          {errors.title && (
            <p className="mt-1 text-sm text-red-600">{errors.title}</p>
          )}
        </div>
        
        {/* 説明 */}
        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
            説明
          </label>
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleChange}
            rows={4}
            className={cn(
              "w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-vertical",
              errors.description ? "border-red-300" : "border-gray-300"
            )}
            placeholder="タスクの詳細を入力..."
            disabled={isSubmitting}
            maxLength={1000}
          />
          {errors.description && (
            <p className="mt-1 text-sm text-red-600">{errors.description}</p>
          )}
          {formData.description && (
            <p className="mt-1 text-sm text-gray-500">
              {formData.description.length}/1000 文字
            </p>
          )}
        </div>
        
        {/* ステータスと優先度 */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label htmlFor="status" className="block text-sm font-medium text-gray-700 mb-1">
              ステータス
            </label>
            <select
              id="status"
              name="status"
              value={formData.status}
              onChange={handleChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              disabled={isSubmitting}
            >
              <option value="TODO">未着手</option>
              <option value="IN_PROGRESS">進行中</option>
              <option value="DONE">完了</option>
            </select>
          </div>
          
          <div>
            <label htmlFor="priority" className="block text-sm font-medium text-gray-700 mb-1">
              優先度
            </label>
            <select
              id="priority"
              name="priority"
              value={formData.priority}
              onChange={handleChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              disabled={isSubmitting}
            >
              <option value="LOW">低</option>
              <option value="MEDIUM">中</option>
              <option value="HIGH">高</option>
            </select>
          </div>
        </div>
        
        {/* 期限と担当者 */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label htmlFor="due_date" className="block text-sm font-medium text-gray-700 mb-1">
              期限
            </label>
            <div className="relative">
              <input
                type="date"
                id="due_date"
                name="due_date"
                value={formData.due_date || ''}
                onChange={handleChange}
                min={getTodayDate()}
                className={cn(
                  "w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 pl-10",
                  errors.due_date ? "border-red-300" : "border-gray-300"
                )}
                disabled={isSubmitting}
              />
              <Calendar className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            </div>
            {errors.due_date && (
              <p className="mt-1 text-sm text-red-600">{errors.due_date}</p>
            )}
          </div>
          
          <div>
            <label htmlFor="assignee_id" className="block text-sm font-medium text-gray-700 mb-1">
              担当者
            </label>
            <div className="relative">
              <select
                id="assignee_id"
                name="assignee_id"
                value={formData.assignee_id || ''}
                onChange={handleChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 pl-10"
                disabled={isSubmitting}
              >
                <option value="">未割り当て</option>
                {users.map(user => (
                  <option key={user.id} value={user.id}>
                    {user.username} ({user.email})
                  </option>
                ))}
              </select>
              <User className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            </div>
          </div>
        </div>
        
        {/* アクションボタン */}
        <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200">
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
            disabled={isSubmitting}
          >
            キャンセル
          </button>
          <button
            type="submit"
            className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed flex items-center"
            disabled={isSubmitting || isLoading}
          >
            {isSubmitting || isLoading ? (
              <>
                <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                {isEditMode ? '更新中...' : '作成中...'}
              </>
            ) : (
              <>
                {isEditMode ? (
                  'タスクを更新'
                ) : (
                  <>
                    <Plus className="h-4 w-4 mr-1" />
                    タスクを作成
                  </>
                )}
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
};

export default TaskForm;