'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';

import { User, TaskFormData } from '@/types';
import { useAuth } from '@/providers/auth-provider';
import TaskForm from '@/components/tasks/TaskForm';
import { createTask } from '@/api/task';
import { getUsers } from '@/api/auth';
import { handleApiError } from '@/lib/utils';
import { success, error as showError } from '@/hooks/use-toast';

export default function NewTaskPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  
  const router = useRouter();
  const { isAuthenticated, isLoading: authLoading } = useAuth();

  // 認証チェック
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/auth/login');
    }
  }, [isAuthenticated, authLoading, router]);

  // ユーザー一覧を取得
  useEffect(() => {
    const fetchUsers = async () => {
      if (!isAuthenticated) return;
      
      try {
        const response = await getUsers();
        if (response.success && response.data) {
          setUsers(response.data.map(u => ({
            id: u.id,
            username: u.username,
            email: u.email,
            role: u.role
          })));
        }
      } catch (err) {
        console.error('Error fetching users:', err);
        showError('ユーザー取得エラー', handleApiError(err));
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, [isAuthenticated]);

  // タスク作成
  const handleCreateTask = async (formData: TaskFormData) => {
    setIsCreating(true);
    
    try {
      const response = await createTask({
        title: formData.title,
        description: formData.description,
        priority: formData.priority,
        due_date: formData.due_date,
      });

      if (response.success && response.data) {
        success('タスクを作成しました');
        
        // 作成したタスクの詳細ページに遷移
        router.push(`/tasks/${response.data.id}`);
      }
    } catch (err) {
      const errorMessage = handleApiError(err);
      showError('タスク作成エラー', errorMessage);
      throw err; // TaskFormでエラーハンドリングするため再throw
    } finally {
      setIsCreating(false);
    }
  };

  // キャンセル
  const handleCancel = () => {
    router.back();
  };

  // ローディング中
  if (authLoading || loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-50">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  // 認証されていない場合
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* ナビゲーション */}
        <div className="mb-8">
          <Link 
            href="/tasks" 
            className="inline-flex items-center text-blue-600 hover:text-blue-800 mb-4"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            タスク一覧に戻る
          </Link>
          
          <div>
            <h1 className="text-3xl font-semibold text-gray-900">新しいタスクを作成</h1>
            <p className="text-gray-600 mt-1">
              タスクの詳細を入力してください
            </p>
          </div>
        </div>

        {/* タスク作成フォーム */}
        <TaskForm
          users={users}
          onSubmit={handleCreateTask}
          onCancel={handleCancel}
          isLoading={isCreating}
        />

        {/* ヘルプテキスト */}
        <div className="mt-8 bg-blue-50 border border-blue-200 rounded-lg p-4">
          <h3 className="text-sm font-medium text-blue-900 mb-2">💡 タスク作成のヒント</h3>
          <ul className="text-sm text-blue-800 space-y-1">
            <li>• タイトルは具体的で分かりやすいものにしましょう</li>
            <li>• 期限を設定すると、カレンダーや期限切れ通知に表示されます</li>
            <li>• 優先度を適切に設定することで、重要なタスクを見つけやすくなります</li>
            <li>• 担当者を設定すると、該当ユーザーに通知が送信されます</li>
          </ul>
        </div>
      </div>
    </div>
  );
}