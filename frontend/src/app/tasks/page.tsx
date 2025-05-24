'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Plus, Search, Filter, ArrowUpDown } from 'lucide-react';

import { Task, User, TaskFilter } from '@/types';
import { useAuth } from '@/providers/auth-provider';
import TaskCard from '@/components/tasks/TaskCard';
import useTasks from '@/lib/hooks/useTask';
import { getUsers } from '@/api/auth';
import { handleApiError, debounce } from '@/lib/utils';

export default function TasksPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [showFilters, setShowFilters] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  
  const router = useRouter();
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();
  
  const {
    tasks,
    isLoading,
    error,
    pagination,
    filters,
    sort,
    setFilters,
    setSorting,
    setPage,
    refreshTasks,
    editTask,
    removeTask,
    updateTaskStatus,
    searchTasks,
    clearError
  } = useTasks();

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
      }
    };

    fetchUsers();
  }, [isAuthenticated]);

  // フィルター変更ハンドラー
  const handleFilterChange = (newFilters: Partial<TaskFilter>) => {
    setFilters({ ...filters, ...newFilters });
  };

  // ソート変更ハンドラー
  const handleSortChange = (field: string) => {
    const newDirection = sort.field === field && sort.direction === 'asc' ? 'desc' : 'asc';
    setSorting(field, newDirection);
  };

  // 検索（デバウンス付き）
  const debouncedSearch = debounce(async (query: string) => {
    if (query.trim()) {
      try {
        await searchTasks(query);
      } catch (error) {
        console.error('Search failed:', error);
      }
    } else {
      refreshTasks();
    }
  }, 500);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const query = e.target.value;
    setSearchQuery(query);
    debouncedSearch(query);
  };

  // タスク編集
  const handleEditTask = async (task: Task) => {
    router.push(`/tasks/${task.id}/edit`);
  };

  // タスク削除
  const handleDeleteTask = async (taskId: string) => {
    try {
      await removeTask(taskId);
    } catch (error) {
      console.error('Failed to delete task:', error);
    }
  };

  // ステータス変更
  const handleStatusChange = async (taskId: string, status: Task['status']) => {
    try {
      await updateTaskStatus(taskId, status);
    } catch (error) {
      console.error('Failed to update task status:', error);
    }
  };

  // タスク作成ページへ移動
  const handleCreateTask = () => {
    router.push('/tasks/new');
  };

  // ページネーション
  const handlePageChange = (page: number) => {
    setPage(page);
  };

  // エラークリア
  const handleErrorDismiss = () => {
    clearError();
  };

  if (authLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-white">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* ヘッダー */}
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-8">
          <div>
            <div className="flex items-center space-x-2 mb-2">
              <Link href="/dashboard" className="text-blue-600 hover:text-blue-800">
                ← ダッシュボード
              </Link>
            </div>
            <h1 className="text-3xl font-semibold text-gray-900">タスク一覧</h1>
            <p className="text-gray-600 mt-1">
              {pagination.total}件のタスク
            </p>
          </div>
          <button
            onClick={handleCreateTask}
            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors flex items-center"
          >
            <Plus className="h-4 w-4 mr-2" />
            新規タスク作成
          </button>
        </div>

        {/* 検索とフィルター */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
          <div className="flex flex-col sm:flex-row gap-4">
            {/* 検索 */}
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={handleSearchChange}
                placeholder="タスクを検索..."
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {/* フィルター切り替え */}
            <button
              onClick={() => setShowFilters(!showFilters)}
              className="flex items-center px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              <Filter className="h-4 w-4 mr-2" />
              フィルター
            </button>

            {/* ソート */}
            <div className="flex items-center space-x-2">
              <ArrowUpDown className="h-4 w-4 text-gray-400" />
              <select
                value={sort.field}
                onChange={(e) => handleSortChange(e.target.value)}
                className="border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="created_at">作成日</option>
                <option value="updated_at">更新日</option>
                <option value="due_date">期限</option>
                <option value="priority">優先度</option>
                <option value="title">タイトル</option>
              </select>
              <button
                onClick={() => setSorting(sort.field, sort.direction === 'asc' ? 'desc' : 'asc')}
                className="p-2 border border-gray-300 rounded-md hover:bg-gray-50"
                title={`${sort.direction === 'asc' ? '降順' : '昇順'}に変更`}
              >
                {sort.direction === 'asc' ? '↑' : '↓'}
              </button>
            </div>
          </div>

          {/* 展開可能なフィルター */}
          {showFilters && (
            <div className="mt-4 pt-4 border-t border-gray-200">
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {/* ステータスフィルター */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    ステータス
                  </label>
                  <select
                    value={filters.status || ''}
                    onChange={(e) => handleFilterChange({ status: e.target.value as Task['status'] || undefined })}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">すべて</option>
                    <option value="TODO">未着手</option>
                    <option value="IN_PROGRESS">進行中</option>
                    <option value="DONE">完了</option>
                  </select>
                </div>

                {/* 優先度フィルター */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    優先度
                  </label>
                  <select
                    value={filters.priority || ''}
                    onChange={(e) => handleFilterChange({ priority: e.target.value as Task['priority'] || undefined })}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">すべて</option>
                    <option value="LOW">低</option>
                    <option value="MEDIUM">中</option>
                    <option value="HIGH">高</option>
                  </select>
                </div>

                {/* 担当者フィルター */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    担当者
                  </label>
                  <select
                    value={filters.assignee_id || ''}
                    onChange={(e) => handleFilterChange({ assignee_id: e.target.value || undefined })}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">すべて</option>
                    <option value="unassigned">未割り当て</option>
                    {users.map((user) => (
                      <option key={user.id} value={user.id}>
                        {user.username}
                      </option>
                    ))}
                  </select>
                </div>

                {/* 期限フィルター */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    期限
                  </label>
                  <select
                    value=""
                    onChange={(e) => {
                      const value = e.target.value;
                      if (value === 'overdue') {
                        const today = new Date().toISOString().split('T')[0];
                        handleFilterChange({ due_date_to: today });
                      } else if (value === 'this_week') {
                        const today = new Date();
                        const nextWeek = new Date(today.getTime() + 7 * 24 * 60 * 60 * 1000);
                        handleFilterChange({
                          due_date_from: today.toISOString().split('T')[0],
                          due_date_to: nextWeek.toISOString().split('T')[0]
                        });
                      } else {
                        handleFilterChange({ due_date_from: undefined, due_date_to: undefined });
                      }
                    }}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">すべて</option>
                    <option value="overdue">期限切れ</option>
                    <option value="this_week">今週</option>
                  </select>
                </div>
              </div>

              {/* フィルタークリア */}
              <div className="mt-4 flex justify-end">
                <button
                  onClick={() => {
                    setFilters({});
                    setSearchQuery('');
                  }}
                  className="text-sm text-gray-600 hover:text-gray-800"
                >
                  フィルターをクリア
                </button>
              </div>
            </div>
          )}
        </div>

        {/* エラー表示 */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-6">
            <div className="flex justify-between items-center">
              <p className="text-red-800">{error}</p>
              <button
                onClick={handleErrorDismiss}
                className="text-red-600 hover:text-red-800"
              >
                ✕
              </button>
            </div>
          </div>
        )}

        {/* タスクリスト */}
        <div className="space-y-4">
          {isLoading ? (
            <div className="flex justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
            </div>
          ) : tasks.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-gray-400 mb-4">
                <Search className="h-12 w-12 mx-auto" />
              </div>
              <h3 className="text-lg font-medium text-gray-900 mb-1">タスクが見つかりません</h3>
              <p className="text-gray-500 mb-4">
                {searchQuery ? '検索条件を変更してください' : 'タスクを作成してください'}
              </p>
              {!searchQuery && (
                <button
                  onClick={handleCreateTask}
                  className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
                >
                  最初のタスクを作成
                </button>
              )}
            </div>
          ) : (
            <>
              {tasks.map((task) => (
                <TaskCard
                  key={task.id}
                  task={task}
                  users={users}
                  onEdit={handleEditTask}
                  onDelete={handleDeleteTask}
                  onStatusChange={handleStatusChange}
                />
              ))}

              {/* ページネーション */}
              {pagination.total > pagination.limit && (
                <div className="flex justify-center mt-8">
                  <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                    <button
                      onClick={() => handlePageChange(pagination.page - 1)}
                      disabled={pagination.page === 1}
                      className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      前へ
                    </button>
                    
                    <div className="relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium text-gray-700">
                      {pagination.page} / {Math.ceil(pagination.total / pagination.limit)}
                    </div>
                    
                    <button
                      onClick={() => handlePageChange(pagination.page + 1)}
                      disabled={pagination.page >= Math.ceil(pagination.total / pagination.limit)}
                      className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      次へ
                    </button>
                  </nav>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}