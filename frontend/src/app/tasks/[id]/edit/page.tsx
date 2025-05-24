'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Eye } from 'lucide-react';

import { Task, User, TaskFormData } from '@/types';
import { useAuth } from '@/providers/auth-provider';
import TaskForm from '@/components/tasks/TaskForm';
import { getTaskById, updateTask } from '@/api/task';
import { getUsers } from '@/api/auth';
import { handleApiError } from '@/lib/utils';
import { success, error as showError } from '@/hooks/use-toast';

interface TaskEditPageProps {
  params: { id: string };
}

export default function TaskEditPage({ params }: TaskEditPageProps) {
  const [task, setTask] = useState<Task | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const router = useRouter();
  const { isAuthenticated, isLoading: authLoading } = useAuth();

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
        console.error('Error fetching data:', err);
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
        success('タスクを更新しました');
        
        // タスク詳細ページに遷移
        router.push(`/tasks/${task.id}`);
      }
    } catch (err) {
      const errorMessage = handleApiError(err);
      showError('タスク更新エラー', errorMessage);
      throw err; // TaskFormでエラーハンドリングするため再throw
    } finally {
      setIsUpdating(false);
    }
  };

  // キャンセル（詳細ページに戻る）
  const handleCancel = () => {
    router.push(`/tasks/${params.id}`);
  };

  // ローディング中
  if (authLoading || loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-50">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  // エラー表示
  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
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

  // 認証されていない場合またはタスクが存在しない場合
  if (!isAuthenticated || !task) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* ナビゲーション */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <Link 
              href={`/tasks/${task.id}`}
              className="inline-flex items-center text-blue-600 hover:text-blue-800"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              タスク詳細に戻る
            </Link>
            
            <Link
              href={`/tasks/${task.id}`}
              className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50"
            >
              <Eye className="h-4 w-4 mr-2" />
              詳細を表示
            </Link>
          </div>
          
          <div>
            <h1 className="text-3xl font-semibold text-gray-900">タスクを編集</h1>
            <p className="text-gray-600 mt-1">
              「{task.title}」の詳細を編集
            </p>
          </div>
        </div>

        {/* タスク編集フォーム */}
        <TaskForm
          task={task}
          users={users}
          onSubmit={handleUpdateTask}
          onCancel={handleCancel}
          isLoading={isUpdating}
        />

        {/* タスク情報 */}
        <div className="mt-8 bg-gray-100 border border-gray-200 rounded-lg p-4">
          <h3 className="text-sm font-medium text-gray-900 mb-2">📋 タスク情報</h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm text-gray-600">
            <div>
              <span className="font-medium">作成日:</span> {new Date(task.created_at).toLocaleDateString('ja-JP')}
            </div>
            <div>
              <span className="font-medium">最終更新:</span> {new Date(task.updated_at).toLocaleDateString('ja-JP')}
            </div>
            <div>
              <span className="font-medium">作成者:</span> {task.created_by}
            </div>
            {task.assignee_id && (
              <div>
                <span className="font-medium">現在の担当者:</span> {
                  users.find(u => u.id === task.assignee_id)?.username || task.assignee_id
                }
              </div>
            )}
          </div>
        </div>

        {/* ヘルプテキスト */}
        <div className="mt-6 bg-amber-50 border border-amber-200 rounded-lg p-4">
          <h3 className="text-sm font-medium text-amber-900 mb-2">⚠️ 編集時の注意</h3>
          <ul className="text-sm text-amber-800 space-y-1">
            <li>• タスクの割り当ては詳細ページから変更できます</li>
            <li>• ステータスを「完了」に変更すると、関係者に通知が送信されます</li>
            <li>• 期限を変更した場合、新しい期限での通知が設定されます</li>
            <li>• 変更内容は即座に保存され、取り消すことはできません</li>
          </ul>
        </div>
      </div>
    </div>
  );
}