'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import Calendar from 'react-calendar';
import 'react-calendar/dist/Calendar.css';

import { Task, TaskListResponse } from '@/types';
import { useAuth } from '@/providers/auth-provider';
import { getTasks, getTaskStats } from '@/api/task';
import { 
  formatDate, 
  getStatusColor, 
  getPriorityColor, 
  getStatusLabel, 
  getPriorityLabel,
  handleApiError 
} from '@/lib/utils';

// カレンダーのコンポーネント型定義
type CalendarValue = Date | null;

export default function Dashboard() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedDate, setSelectedDate] = useState<Date>(new Date());
  const [taskStats, setTaskStats] = useState({
    total: 0,
    todo: 0,
    in_progress: 0,
    done: 0,
    overdue: 0
  });
  
  const router = useRouter();
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();

  // 認証チェック
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/auth/login');
    }
  }, [isAuthenticated, authLoading, router]);

  // 日付に基づいたタスクをフィルタリングする関数
  const getTasksForDate = (date: Date): Task[] => {
    return tasks.filter(task => {
      if (!task.due_date) return false;
      
      const taskDate = new Date(task.due_date);
      return (
        taskDate.getDate() === date.getDate() &&
        taskDate.getMonth() === date.getMonth() &&
        taskDate.getFullYear() === date.getFullYear()
      );
    });
  };

  // APIからタスクを取得
  useEffect(() => {
    const fetchDashboardData = async () => {
      if (!isAuthenticated) return;
      
      setLoading(true);
      try {
        // タスク一覧を取得
        const tasksResponse: TaskListResponse = await getTasks({
          page: 1,
          page_size: 50, // ダッシュボード用に多めに取得
          sort_field: 'due_date',
          sort_direction: 'ASC'
        });

        if (tasksResponse.success && tasksResponse.data) {
          setTasks(tasksResponse.data.tasks);
        }

        // タスク統計を取得（カスタムエンドポイントがある場合）
        try {
          const statsResponse = await getTaskStats();
          if (statsResponse.success && statsResponse.data) {
            setTaskStats(statsResponse.data);
          }
        } catch (statsError) {
          // 統計APIがない場合は、取得したタスクから計算
          const todoCount = tasksResponse.data.tasks.filter(t => t.status === 'TODO').length;
          const inProgressCount = tasksResponse.data.tasks.filter(t => t.status === 'IN_PROGRESS').length;
          const doneCount = tasksResponse.data.tasks.filter(t => t.status === 'DONE').length;
          const overdueCount = tasksResponse.data.tasks.filter(t => 
            t.due_date && new Date(t.due_date) < new Date() && t.status !== 'DONE'
          ).length;

          setTaskStats({
            total: tasksResponse.data.tasks.length,
            todo: todoCount,
            in_progress: inProgressCount,
            done: doneCount,
            overdue: overdueCount
          });
        }

      } catch (err) {
        console.error('Error fetching dashboard data:', err);
        setError(handleApiError(err));
      } finally {
        setLoading(false);
      }
    };

    fetchDashboardData();
  }, [isAuthenticated]);

  // 日付変更ハンドラー
  const handleDateChange = (value: CalendarValue) => {
    if (value instanceof Date) {
      setSelectedDate(value);
    }
  };

  // タスク作成ページへ移動
  const handleCreateTask = () => {
    router.push('/tasks/new');
  };

  // 選択日のタスク
  const tasksForSelectedDate = getTasksForDate(selectedDate);

  // カレンダーに日付装飾を追加するための関数
  const tileContent = ({ date, view }: { date: Date; view: string }) => {
    if (view === 'month') {
      const tasksForDay = getTasksForDate(date);
      return tasksForDay.length > 0 ? (
        <div className="absolute bottom-0 left-0 right-0 h-1 bg-blue-500 rounded"></div>
      ) : null;
    }
    return null;
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
      <div className="flex items-center justify-center h-screen bg-white">
        <div className="text-center">
          <div className="text-red-600 mb-4">{error}</div>
          <button
            onClick={() => window.location.reload()}
            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
          >
            再読み込み
          </button>
        </div>
      </div>
    );
  }

  // 認証されていない場合
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-semibold text-gray-900">ダッシュボード</h1>
            {user && (
              <p className="text-gray-600 mt-1">
                おかえりなさい、{user.username}さん
              </p>
            )}
          </div>
          <button
            onClick={handleCreateTask}
            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
          >
            新規タスク作成
          </button>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* 左側: カレンダー */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
              <h2 className="text-lg font-medium text-gray-900 mb-4">カレンダー</h2>
              <div className="calendar-wrapper">
                <Calendar
                  onChange={handleDateChange}
                  value={selectedDate}
                  tileContent={tileContent}
                  className="w-full rounded-md"
                />
              </div>
            </div>
          </div>

          {/* 右側: 選択した日付のタスクとタスク概要 */}
          <div className="lg:col-span-2">
            {/* 選択した日付のタスク */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
              <h2 className="text-lg font-medium text-gray-900 mb-4">
                {selectedDate.toLocaleDateString('ja-JP', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric'
                })}のタスク
              </h2>
              
              {tasksForSelectedDate.length > 0 ? (
                <div className="space-y-3">
                  {tasksForSelectedDate.map((task) => (
                    <Link href={`/tasks/${task.id}`} key={task.id} className="block">
                      <div className="border border-gray-200 rounded-md p-4 hover:bg-gray-50 transition-colors">
                        <div className="flex justify-between items-start">
                          <div className="flex-1">
                            <h3 className="text-base font-medium text-gray-900">{task.title}</h3>
                            {task.description && (
                              <p className="text-sm text-gray-500 mt-1 line-clamp-2">{task.description}</p>
                            )}
                          </div>
                          <div className="flex space-x-2 ml-4">
                            <span className={`px-2 py-1 rounded-full text-xs border ${getPriorityColor(task.priority)}`}>
                              {getPriorityLabel(task.priority)}
                            </span>
                            <span className={`px-2 py-1 rounded-full text-xs border ${getStatusColor(task.status)}`}>
                              {getStatusLabel(task.status)}
                            </span>
                          </div>
                        </div>
                        {task.assignee_id && (
                          <div className="mt-2 text-xs text-gray-500">
                            担当者: {task.assignee_id}
                          </div>
                        )}
                      </div>
                    </Link>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8">
                  <div className="text-gray-400 mb-2">
                    <svg className="mx-auto h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                    </svg>
                  </div>
                  <p className="text-gray-500">この日のタスクはありません</p>
                  <button
                    onClick={handleCreateTask}
                    className="mt-2 text-blue-600 hover:text-blue-800 text-sm"
                  >
                    新しいタスクを作成
                  </button>
                </div>
              )}
            </div>

            {/* タスク概要 */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-lg font-medium text-gray-900">タスク概要</h2>
                <Link 
                  href="/tasks"
                  className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                >
                  すべて見る →
                </Link>
              </div>
              
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="bg-gray-50 p-4 rounded-md text-center">
                  <h3 className="text-sm font-medium text-gray-900 mb-2">ToDo</h3>
                  <p className="text-2xl font-bold text-gray-700">
                    {taskStats.todo}
                  </p>
                </div>
                
                <div className="bg-blue-50 p-4 rounded-md text-center">
                  <h3 className="text-sm font-medium text-gray-900 mb-2">進行中</h3>
                  <p className="text-2xl font-bold text-blue-700">
                    {taskStats.in_progress}
                  </p>
                </div>
                
                <div className="bg-green-50 p-4 rounded-md text-center">
                  <h3 className="text-sm font-medium text-gray-900 mb-2">完了</h3>
                  <p className="text-2xl font-bold text-green-700">
                    {taskStats.done}
                  </p>
                </div>

                <div className="bg-red-50 p-4 rounded-md text-center">
                  <h3 className="text-sm font-medium text-gray-900 mb-2">期限切れ</h3>
                  <p className="text-2xl font-bold text-red-700">
                    {taskStats.overdue}
                  </p>
                </div>
              </div>

              {/* 進捗率 */}
              {taskStats.total > 0 && (
                <div className="mt-6">
                  <div className="flex justify-between text-sm text-gray-600 mb-2">
                    <span>全体の進捗</span>
                    <span>{Math.round((taskStats.done / taskStats.total) * 100)}%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-green-600 h-2 rounded-full transition-all duration-300"
                      style={{ width: `${(taskStats.done / taskStats.total) * 100}%` }}
                    ></div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}