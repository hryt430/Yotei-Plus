import React from 'react';
import { CheckSquare, ArrowLeft, ArrowRight } from 'lucide-react';
import { Task, User } from '@/types';
import TaskCard from './TaskCard';
import { cn } from '@/lib/utils';

interface TaskListProps {
  tasks: Task[];
  users?: User[];
  isLoading?: boolean;
  error?: string | null;
  onEditTask?: (task: Task) => void;
  onDeleteTask?: (taskId: string) => void;
  onStatusChange?: (taskId: string, status: Task['status']) => void;
  pagination?: {
    page: number;
    limit: number;
    total: number;
  };
  onPageChange?: (page: number) => void;
  compact?: boolean;
  showPagination?: boolean;
}

const TaskList: React.FC<TaskListProps> = ({
  tasks,
  users = [],
  isLoading = false,
  error = null,
  onEditTask,
  onDeleteTask,
  onStatusChange,
  pagination,
  onPageChange,
  compact = false,
  showPagination = true,
}) => {
  // ローディング表示
  if (isLoading && tasks.length === 0) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  // エラー表示
  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
        <div className="text-red-600 mb-2">エラーが発生しました</div>
        <p className="text-red-500 text-sm">{error}</p>
      </div>
    );
  }

  // 空の状態
  if (tasks.length === 0) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-8 text-center">
        <div className="mx-auto mb-4 w-12 h-12 flex items-center justify-center rounded-full bg-gray-100">
          <CheckSquare className="h-6 w-6 text-gray-400" />
        </div>
        <h3 className="text-lg font-medium text-gray-900 mb-2">タスクが見つかりません</h3>
        <p className="text-gray-500 mb-4">
          条件に一致するタスクがありません。検索条件を変更するか、新しいタスクを作成してください。
        </p>
      </div>
    );
  }

  // ページ数計算
  const totalPages = pagination ? Math.ceil(pagination.total / pagination.limit) : 1;
  const currentPage = pagination?.page || 1;

  // ページ番号の生成
  const getPageNumbers = (current: number, total: number) => {
    if (total <= 7) {
      return Array.from({ length: total }, (_, i) => i + 1);
    }

    if (current <= 3) {
      return [1, 2, 3, 4, 5, '...', total];
    }

    if (current >= total - 2) {
      return [1, '...', total - 4, total - 3, total - 2, total - 1, total];
    }

    return [1, '...', current - 1, current, current + 1, '...', total];
  };

  return (
    <div className="space-y-4">
      {/* タスクリスト */}
      <div className={cn(
        "space-y-3",
        compact && "space-y-2"
      )}>
        {/* ローディング中のオーバーレイ */}
        {isLoading && (
          <div className="relative">
            <div className="absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center z-10 rounded-lg">
              <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
            </div>
          </div>
        )}
        
        {tasks.map(task => (
          <TaskCard
            key={task.id}
            task={task}
            users={users}
            onEdit={onEditTask}
            onDelete={onDeleteTask}
            onStatusChange={onStatusChange}
            compact={compact}
          />
        ))}
      </div>

      {/* ページネーション */}
      {showPagination && pagination && totalPages > 1 && (
        <div className="flex items-center justify-between pt-4">
          {/* 結果情報 */}
          <div className="text-sm text-gray-500">
            {pagination.total > 0 ? (
              <>
                {(currentPage - 1) * pagination.limit + 1} - {Math.min(currentPage * pagination.limit, pagination.total)} / {pagination.total}件
              </>
            ) : (
              '0件'
            )}
          </div>

          {/* ページネーションコントロール */}
          <nav className="flex items-center space-x-1">
            {/* 前のページ */}
            <button
              onClick={() => onPageChange?.(currentPage - 1)}
              disabled={currentPage === 1}
              className={cn(
                "relative inline-flex items-center px-3 py-2 border text-sm font-medium rounded-l-md",
                currentPage === 1
                  ? "bg-gray-100 text-gray-400 cursor-not-allowed border-gray-300"
                  : "bg-white text-gray-700 border-gray-300 hover:bg-gray-50"
              )}
            >
              <ArrowLeft className="h-4 w-4 mr-1" />
              前へ
            </button>

            {/* ページ番号 */}
            <div className="flex items-center">
              {getPageNumbers(currentPage, totalPages).map((pageNum, idx) => (
                pageNum === '...' ? (
                  <span
                    key={`ellipsis-${idx}`}
                    className="relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium text-gray-700"
                  >
                    ...
                  </span>
                ) : (
                  <button
                    key={`page-${pageNum}`}
                    onClick={() => onPageChange?.(Number(pageNum))}
                    className={cn(
                      "relative inline-flex items-center px-4 py-2 border text-sm font-medium",
                      Number(pageNum) === currentPage
                        ? "z-10 bg-blue-50 border-blue-500 text-blue-600"
                        : "bg-white border-gray-300 text-gray-700 hover:bg-gray-50"
                    )}
                  >
                    {pageNum}
                  </button>
                )
              ))}
            </div>

            {/* 次のページ */}
            <button
              onClick={() => onPageChange?.(currentPage + 1)}
              disabled={currentPage >= totalPages}
              className={cn(
                "relative inline-flex items-center px-3 py-2 border text-sm font-medium rounded-r-md",
                currentPage >= totalPages
                  ? "bg-gray-100 text-gray-400 cursor-not-allowed border-gray-300"
                  : "bg-white text-gray-700 border-gray-300 hover:bg-gray-50"
              )}
            >
              次へ
              <ArrowRight className="h-4 w-4 ml-1" />
            </button>
          </nav>
        </div>
      )}

      {/* 結果サマリー（ページネーションなしの場合） */}
      {!showPagination && (
        <div className="text-center text-sm text-gray-500 pt-4 border-t border-gray-200">
          {tasks.length}件のタスク
        </div>
      )}
    </div>
  );
};

export default TaskList;