'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';

// タイプ定義
type Task = {
  id: string;
  title: string;
  description: string;
  dueDate: string;
  priority: 'low' | 'medium' | 'high';
  status: 'todo' | 'in-progress' | 'done';
  assignedTo?: string;
  createdAt: string;
  updatedAt: string;
};

type User = {
  id: string;
  name: string;
  email: string;
};

export default function TaskDetail({ params }: { params: { id: string } }) {
  const [task, setTask] = useState<Task | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [editedTask, setEditedTask] = useState<Partial<Task>>({});
  const [assignUserId, setAssignUserId] = useState<string>('');
  const router = useRouter();

  // タスク詳細を取得
  useEffect(() => {
    const fetchTaskDetail = async () => {
      setLoading(true);
      try {
        const response = await fetch(`/api/tasks/${params.id}`, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('token')}`,
          },
        });

        if (!response.ok) {
          throw new Error('タスクの取得に失敗しました');
        }

        const data = await response.json();
        setTask(data);
        setEditedTask(data);
      } catch (err) {
        console.error('Error fetching task:', err);
        setError('タスクの読み込み中にエラーが発生しました');
      } finally {
        setLoading(false);
      }
    };

    const fetchUsers = async () => {
      try {
        const response = await fetch('/api/users', {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('token')}`,
          },
        });

        if (!response.ok) {
          throw new Error('ユーザーの取得に失敗しました');
        }

        const data = await response.json();
        setUsers(data);
      } catch (err) {
        console.error('Error fetching users:', err);
      }
    };

    fetchTaskDetail();
    fetchUsers();
  }, [params.id]);

  // 編集モードの切り替え
  const handleEditClick = () => {
    setIsEditing(true);
  };

  // 編集キャンセル
  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditedTask(task || {});
  };

  // 入力フィールドの変更ハンドラ
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setEditedTask((prev) => ({ ...prev, [name]: value }));
  };

  // タスク更新処理
  const handleSaveEdit = async () => {
    try {
      const response = await fetch(`/api/tasks/${params.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(editedTask),
      });

      if (!response.ok) {
        throw new Error('タスクの更新に失敗しました');
      }

      const updatedTask = await response.json();
      setTask(updatedTask);
      setIsEditing(false);
    } catch (err) {
      console.error('Error updating task:', err);
      setError('タスクの更新中にエラーが発生しました');
    }
  };

  // タスク削除処理
  const handleDeleteTask = async () => {
    if (!confirm('本当にこのタスクを削除しますか？')) {
      return;
    }

    try {
      const response = await fetch(`/api/tasks/${params.id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('タスクの削除に失敗しました');
      }

      router.push('/tasks');
    } catch (err) {
      console.error('Error deleting task:', err);
      setError('タスクの削除中にエラーが発生しました');
    }
  };

  // タスク割り当て処理
  const handleAssignTask = async () => {
    if (!assignUserId) return;

    try {
      const response = await fetch(`/api/tasks/${params.id}/assign`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({ userId: assignUserId }),
      });

      if (!response.ok) {
        throw new Error('タスクの割り当てに失敗しました');
      }

      const updatedTask = await response.json();
      setTask(updatedTask);
      setAssignUserId('');
    } catch (err) {
      console.error('Error assigning task:', err);
      setError('タスクの割り当て中にエラーが発生しました');
    }
  };

  // 優先度に応じた色を取得
  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high':
        return 'bg-red-100 text-red-800';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800';
      case 'low':
        return 'bg-green-100 text-green-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  // ステータスに応じた色を取得
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'done':
        return 'bg-blue-100 text-blue-800';
      case 'in-progress':
        return 'bg-purple-100 text-purple-800';
      case 'todo':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  // 日付のフォーマット
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ja-JP', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-white">
        <div className="text-gray-600">読み込み中...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen bg-white">
        <div className="text-red-600">{error}</div>
      </div>
    );
  }

  if (!task) {
    return (
      <div className="flex items-center justify-center h-screen bg-white">
        <div className="text-gray-600">タスクが見つかりませんでした</div>
      </div>
    );
  }

  const assignedUser = users.find(user => user.id === task.assignedTo);

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center mb-8">
          <Link href="/dashboard" className="text-blue-600 hover:text-blue-800 mr-4">
            ← ダッシュボードに戻る
          </Link>
          <h1 className="text-2xl font-semibold text-gray-900">
            {isEditing ? '編集中: ' : ''}
            {task.title}
          </h1>
        </div>

        {/* タスク詳細情報 */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-8">
          {isEditing ? (
            <div className="space-y-4">
              <div>
                <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-1">
                  タイトル
                </label>
                <input
                  type="text"
                  id="title"
                  name="title"
                  value={editedTask.title || ''}
                  onChange={handleInputChange}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
                  説明
                </label>
                <textarea
                  id="description"
                  name="description"
                  value={editedTask.description || ''}
                  onChange={handleInputChange}
                  rows={4}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                ></textarea>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label htmlFor="dueDate" className="block text-sm font-medium text-gray-700 mb-1">
                    期限
                  </label>
                  <input
                    type="date"
                    id="dueDate"
                    name="dueDate"
                    value={editedTask.dueDate ? new Date(editedTask.dueDate).toISOString().split('T')[0] : ''}
                    onChange={handleInputChange}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <div>
                  <label htmlFor="priority" className="block text-sm font-medium text-gray-700 mb-1">
                    優先度
                  </label>
                  <select
                    id="priority"
                    name="priority"
                    value={editedTask.priority || 'low'}
                    onChange={handleInputChange}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="low">低</option>
                    <option value="medium">中</option>
                    <option value="high">高</option>
                  </select>
                </div>
              </div>

              <div>
                <label htmlFor="status" className="block text-sm font-medium text-gray-700 mb-1">
                  ステータス
                </label>
                <select
                  id="status"
                  name="status"
                  value={editedTask.status || 'todo'}
                  onChange={handleInputChange}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="todo">ToDo</option>
                  <option value="in-progress">進行中</option>
                  <option value="done">完了</option>
                </select>
              </div>

              <div className="flex justify-end space-x-3 pt-4">
                <button
                  onClick={handleCancelEdit}
                  className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  キャンセル
                </button>
                <button
                  onClick={handleSaveEdit}
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                >
                  保存
                </button>
              </div>
            </div>
          ) : (
            <div>
              <div className="flex justify-between items-start mb-6">
                <div>
                  <h2 className="text-xl font-medium text-gray-900">{task.title}</h2>
                  <div className="flex space-x-2 mt-2">
                    <span className={`px-2 py-1 rounded-full text-xs ${getPriorityColor(task.priority)}`}>
                      優先度: {task.priority}
                    </span>
                    <span className={`px-2 py-1 rounded-full text-xs ${getStatusColor(task.status)}`}>
                      ステータス: {task.status}
                    </span>
                  </div>
                </div>
                <div className="flex space-x-3">
                  <button
                    onClick={handleEditClick}
                    className="px-3 py-1 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 transition-colors text-sm"
                  >
                    編集
                  </button>
                  <button
                    onClick={handleDeleteTask}
                    className="px-3 py-1 border border-red-300 rounded-md text-red-700 hover:bg-red-50 transition-colors text-sm"
                  >
                    削除
                  </button>
                </div>
              </div>

              <div className="border-t border-gray-200 pt-4 pb-2">
                <h3 className="text-sm font-medium text-gray-700 mb-2">説明</h3>
                <p className="text-gray-600 whitespace-pre-line">{task.description}</p>
              </div>

              <div className="border-t border-gray-200 pt-4 pb-2">
                <h3 className="text-sm font-medium text-gray-700 mb-2">期限</h3>
                <p className="text-gray-600">{formatDate(task.dueDate)}</p>
              </div>

              <div className="border-t border-gray-200 pt-4 pb-2">
                <h3 className="text-sm font-medium text-gray-700 mb-2">担当者</h3>
                {assignedUser ? (
                  <p className="text-gray-600">{assignedUser.name} ({assignedUser.email})</p>
                ) : (
                  <p className="text-gray-500 italic">未割り当て</p>
                )}
              </div>

              <div className="border-t border-gray-200 pt-4 pb-2">
                <h3 className="text-sm font-medium text-gray-700 mb-2">更新日時</h3>
                <p className="text-gray-600">{formatDate(task.updatedAt)}</p>
              </div>
            </div>
          )}
        </div>

        {/* タスク割り当て */}
        {!isEditing && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">タスク割り当て</h2>
            <div className="flex items-center space-x-2">
              <select
                value={assignUserId}
                onChange={(e) => setAssignUserId(e.target.value)}
                className="flex-1 border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">ユーザーを選択</option>
                {users.map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.name} ({user.email})
                  </option>
                ))}
              </select>
              <button
                onClick={handleAssignTask}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
              >
                割り当て
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}