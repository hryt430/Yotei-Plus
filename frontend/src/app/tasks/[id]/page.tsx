'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Edit, Trash2, User, Calendar, Clock, CheckCircle2, AlertCircle, ArrowLeft } from 'lucide-react';

import { Task, User as UserType, TaskFormData } from '@/types';
import { useAuth } from '@/providers/auth-provider';
import TaskForm from '@/components/tasks/TaskForm';
import { 
  getTaskById, 
  updateTask, 
  deleteTask, 
  assignTask, 
  changeTaskStatus 
} from '@/api/task';
import { getUsers } from '@/api/auth';
import { 
  formatDate, 
  formatDateTime,
  getRelativeDateLabel, 
  getStatusColor, 
  getPriorityColor, 
  getStatusLabel, 
  getPriorityLabel,
  handleApiError,
  cn
} from '@/lib/utils';
import { success, error as showError } from '@/hooks/use-toast';

interface TaskDetailPageProps {
  params: { id: string };
}

export default function TaskDetailPage({ params }: TaskDetailPageProps) {
  const [task, setTask] = useState<Task | null>(null);
  const [users, setUsers] = useState<UserType[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState<string>('');
  
  const router = useRouter();
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();

  // 認証チェック
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/auth/login');
    }
  }, [isAuthenticated, authLoading, router]);

  // タスク詳細とユーザー一覧を取得
  useEffect(() => {
    const fetchData = async () => {
      if (!isAuthenticated || !params.id) return;
      
      setLoading(true);
      try {
        // タスク詳細とユーザー一覧を並行取得
        const [taskResponse, usersResponse] = await Promise.all([
          getTaskById(params.id),
          getUsers()
        ]);

        if (taskResponse.success && taskResponse.data) {
          setTask(taskResponse.data);
        } else {
          throw new Error('タスクが見つかりません');
        }

        if (usersResponse.success && usersResponse.data) {
          setUsers(usersResponse.data.map(u => ({
            id: u.id,
            username: u.username,
            email: u.email,
            role: u.role
          })));
        }

      } catch (err) {
        console.error('Error fetching task:', err);
        setError(handleApiError(err));
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [isAuthenticated, params.id]);

  // タスク更新
  const handleUpdateTask = async (formData: TaskFormData) => {
    if (!task) return;
    
    setIsUpdating(true);
    try {
      const response = await updateTask(task.id, {
        title: formData.title,
        description: formData.description,
        status: formData.status,
        priority: formData.priority,
        due_date: formData.due_date,
      });

      if (response.success && response.data) {
        setTask(response.data);
        setIsEditing(false);
        success('タスクを更新しました');
      }
    } catch (err) {
      const errorMessage = handleApiError(err);
      showError('更新エラー', errorMessage);
      throw err;
    } finally {
      setIsUpdating(false);
    }
  };

  // タスク削除
  const handleDeleteTask = async () => {
    if (!task || !confirm('このタスクを削除しますか？この操作は取り消せません。')) {
      return;
    }

    try {
      await deleteTask(task.id);
      success('タスクを削除しました');
      router.push('/tasks');
    } catch (err) {
      const errorMessage = handleApiError(err);
      showError('削除エラー', errorMessage);
    }
  };

  // タスク割り当て
  const handleAssignTask = async () => {
    if (!task || !selectedUserId) return;

    try {
      const response = await assignTask(task.id, selectedUserId);
      
      if (response.success && response.data) {
        setTask(response.data);
        setSelectedUserId('');
        success('タスクを割り当てました');
      }
    } catch (err) {
      const errorMessage = handleApiError(err);
      showError('割り当てエラー', errorMessage);
    }
  };

  // ステータス変更
  const handleStatusChange = async (newStatus: Task['status']) => {
    if (!task) return;

    try {
      const response = await changeTaskStatus(task.id, newStatus);
      
      if (response.success && response.data) {
        setTask(response.data);
        success('ステータスを更新しました');
      }
    } catch (err) {
      const errorMessage = handleApiError(err);
      showError('ステータス更新エラー', errorMessage);
    }
  };

  // 編集キャンセル
  const handleCancelEdit = () => {
    setIsEditing(false);
  };

  // ローディング中
  if (authLoading || loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-white">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  // エラー表示
  if (error) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-600 mb-4">{error}</div>
          <div className="space-x-4">
            <Link
              href="/tasks"
              className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
            >
              タスク一覧に戻る
            </Link>
            <button
              onClick={() => window.location.reload()}
              className="bg-gray-600 text-white px-4 py-2 rounded-md hover:bg-gray-700"
            >
              再読み込み
            </button>
          </div>
        </div>
      </div>
    );
  }

  // 認証されていない場合
  if (!isAuthenticated || !task) {
    return null;
  }

  const assignedUser = task.assignee_id 
    ? users.find(u => u.id === task.assignee_id)
    : undefined;
    
  const creator = users.find(u => u.id === task.created_by);
  const isOverdue = task.due_date && 
    new Date(task.due_date) < new Date() && 
    task.status !== 'DONE';

  // 編集モード
  if (isEditing) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="mb-6">
            <button
              onClick={handleCancelEdit}
              className="flex items-center text-blue-600 hover:text-blue-800"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              詳細に戻る
            </button>
          </div>
          
          <TaskForm
            task={task}
            users={users}
            onSubmit={handleUpdateTask}
            onCancel={handleCancelEdit}
            isLoading={isUpdating}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* ナビゲーション */}
        <div className="flex items-center mb-8">
          <Link href="/tasks" className="flex items-center text-blue-600 hover:text-blue-800 mr-4">
            <ArrowLeft className="h-4 w-4 mr-2" />
            タスク一覧に戻る
          </Link>
        </div>

        {/* タスク詳細カード */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          {/* ヘッダー */}
          <div className="px-6 py-4 border-b border-gray-200">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <h1 className="text-2xl font-semibold text-gray-900 mb-2">
                  {task.title}
                </h1>
                <div className="flex flex-wrap items-center gap-2">
                  <span className={cn("text-sm px-3 py-1 rounded-full border", getStatusColor(task.status))}>
                    {getStatusLabel(task.status)}
                  </span>
                  <span className={cn("text-sm px-3 py-1 rounded-full border", getPriorityColor(task.priority))}>
                    {getPriorityLabel(task.priority)}優先度
                  </span>
                  {isOverdue && (
                    <span className="text-sm px-3 py-1 rounded-full bg-red-100 text-red-800 border border-red-200 flex items-center">
                      <AlertCircle className="h-3 w-3 mr-1" />
                      期限切れ
                    </span>
                  )}
                </div>
              </div>
              
              <div className="flex items-center space-x-2 ml-4">
                <button
                  onClick={() => setIsEditing(true)}
                  className="flex items-center px-3 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
                >
                  <Edit className="h-4 w-4 mr-2" />
                  編集
                </button>
                <button
                  onClick={handleDeleteTask}
                  className="flex items-center px-3 py-2 border border-red-300 rounded-md text-red-700 hover:bg-red-50"
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  削除
                </button>
              </div>
            </div>
          </div>

          {/* メインコンテンツ */}
          <div className="px-6 py-6">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              {/* 左側: タスク詳細 */}
              <div className="lg:col-span-2 space-y-6">
                {/* 説明 */}
                <div>
                  <h3 className="text-sm font-medium text-gray-700 mb-2">説明</h3>
                  <div className="bg-gray-50 rounded-md p-4">
                    {task.description ? (
                      <p className="text-gray-900 whitespace-pre-line">{task.description}</p>
                    ) : (
                      <p className="text-gray-500 italic">説明が設定されていません</p>
                    )}
                  </div>
                </div>

                {/* ステータス変更 */}
                <div>
                  <h3 className="text-sm font-medium text-gray-700 mb-2">ステータス変更</h3>
                  <div className="flex space-x-2">
                    {(['TODO', 'IN_PROGRESS', 'DONE'] as const).map((status) => (
                      <button
                        key={status}
                        onClick={() => handleStatusChange(status)}
                        className={cn(
                          "px-4 py-2 rounded-md border text-sm font-medium transition-colors",
                          task.status === status
                            ? "bg-blue-600 text-white border-blue-600"
                            : "bg-white text-gray-700 border-gray-300 hover:bg-gray-50"
                        )}
                      >
                        {getStatusLabel(status)}
                      </button>
                    ))}
                  </div>
                </div>
              </div>

              {/* 右側: メタデータとアクション */}
              <div className="space-y-6">
                {/* 基本情報 */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h3 className="text-sm font-medium text-gray-700 mb-3">基本情報</h3>
                  <div className="space-y-3">
                    {/* 期限 */}
                    {task.due_date && (
                      <div className="flex items-center">
                        <Calendar className="h-4 w-4 text-gray-400 mr-2" />
                        <div>
                          <p className="text-sm text-gray-900">
                            {formatDate(task.due_date)}
                          </p>
                          <p className={cn(
                            "text-xs",
                            isOverdue ? "text-red-600 font-medium" : "text-gray-500"
                          )}>
                            {getRelativeDateLabel(task.due_date)}
                          </p>
                        </div>
                      </div>
                    )}

                    {/* 担当者 */}
                    <div className="flex items-center">
                      <User className="h-4 w-4 text-gray-400 mr-2" />
                      <div>
                        {assignedUser ? (
                          <div>
                            <p className="text-sm text-gray-900">{assignedUser.username}</p>
                            <p className="text-xs text-gray-500">{assignedUser.email}</p>
                          </div>
                        ) : (
                          <p className="text-sm text-gray-500 italic">未割り当て</p>
                        )}
                      </div>
                    </div>

                    {/* 作成者 */}
                    {creator && (
                      <div className="flex items-center">
                        <CheckCircle2 className="h-4 w-4 text-gray-400 mr-2" />
                        <div>
                          <p className="text-sm text-gray-900">作成者: {creator.username}</p>
                          <p className="text-xs text-gray-500">{creator.email}</p>
                        </div>
                      </div>
                    )}

                    {/* 作成日・更新日 */}
                    <div className="flex items-center">
                      <Clock className="h-4 w-4 text-gray-400 mr-2" />
                      <div>
                        <p className="text-sm text-gray-900">
                          作成: {formatDateTime(task.created_at)}
                        </p>
                        {task.updated_at !== task.created_at && (
                          <p className="text-xs text-gray-500">
                            更新: {formatDateTime(task.updated_at)}
                          </p>
                        )}
                      </div>
                    </div>
                  </div>
                </div>

                {/* タスク割り当て */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h3 className="text-sm font-medium text-gray-700 mb-3">タスク割り当て</h3>
                  <div className="space-y-2">
                    <select
                      value={selectedUserId}
                      onChange={(e) => setSelectedUserId(e.target.value)}
                      className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="">ユーザーを選択</option>
                      {users.map((user) => (
                        <option key={user.id} value={user.id}>
                          {user.username} ({user.email})
                        </option>
                      ))}
                    </select>
                    <button
                      onClick={handleAssignTask}
                      disabled={!selectedUserId}
                      className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-sm"
                    >
                      割り当て
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}