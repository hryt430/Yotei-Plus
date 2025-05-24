import React, { useState } from 'react';
import { Search, Filter, X, Calendar, User, ArrowUpDown } from 'lucide-react';
import { User as UserType, TaskFilter as TaskFilterType } from '@/types';
import { cn } from '@/lib/utils';

interface TaskFilterProps {
  users?: UserType[];
  filters: TaskFilterType;
  sort: {
    field: string;
    direction: 'asc' | 'desc';
  };
  onFilterChange: (filters: TaskFilterType) => void;
  onSortChange: (field: string, direction: 'asc' | 'desc') => void;
  onSearch: (query: string) => void;
  searchQuery?: string;
}

const TaskFilter: React.FC<TaskFilterProps> = ({
  users = [],
  filters,
  sort,
  onFilterChange,
  onSortChange,
  onSearch,
  searchQuery = '',
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [localSearchQuery, setLocalSearchQuery] = useState(searchQuery);

  // 検索ハンドラー
  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const query = e.target.value;
    setLocalSearchQuery(query);
    onSearch(query);
  };

  // フィルター変更ハンドラー
  const handleFilterChange = (key: keyof TaskFilterType, value: any) => {
    onFilterChange({
      ...filters,
      [key]: value || undefined,
    });
  };

  // ソート変更ハンドラー
  const handleSortChange = (field: string) => {
    const newDirection = sort.field === field && sort.direction === 'asc' ? 'desc' : 'asc';
    onSortChange(field, newDirection);
  };

  // フィルタークリア
  const handleClearFilters = () => {
    onFilterChange({});
    setLocalSearchQuery('');
    onSearch('');
  };

  // アクティブなフィルター数を計算
  const activeFiltersCount = Object.values(filters).filter(value => 
    value !== undefined && value !== null && value !== ''
  ).length;

  // 日付範囲プリセット
  const getDatePreset = (preset: string) => {
    const today = new Date();
    const todayStr = today.toISOString().split('T')[0];
    
    switch (preset) {
      case 'today':
        return { due_date_from: todayStr, due_date_to: todayStr };
      case 'this_week':
        const startOfWeek = new Date(today);
        startOfWeek.setDate(today.getDate() - today.getDay());
        const endOfWeek = new Date(startOfWeek);
        endOfWeek.setDate(startOfWeek.getDate() + 6);
        return {
          due_date_from: startOfWeek.toISOString().split('T')[0],
          due_date_to: endOfWeek.toISOString().split('T')[0]
        };
      case 'overdue':
        return { due_date_to: todayStr };
      default:
        return {};
    }
  };

  const handleDatePreset = (preset: string) => {
    const dateFilter = getDatePreset(preset);
    onFilterChange({
      ...filters,
      ...dateFilter,
    });
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200">
      {/* メインフィルターバー */}
      <div className="p-4">
        <div className="flex flex-col sm:flex-row gap-4">
          {/* 検索 */}
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <input
              type="text"
              value={localSearchQuery}
              onChange={handleSearchChange}
              placeholder="タスクを検索..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
            {localSearchQuery && (
              <button
                onClick={() => {
                  setLocalSearchQuery('');
                  onSearch('');
                }}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>

          {/* クイックフィルター */}
          <div className="flex items-center space-x-2">
            {/* ステータス */}
            <select
              value={filters.status || ''}
              onChange={(e) => handleFilterChange('status', e.target.value as any)}
              className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">すべて</option>
              <option value="TODO">未着手</option>
              <option value="IN_PROGRESS">進行中</option>
              <option value="DONE">完了</option>
            </select>

            {/* 優先度 */}
            <select
              value={filters.priority || ''}
              onChange={(e) => handleFilterChange('priority', e.target.value as any)}
              className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">すべて</option>
              <option value="LOW">低</option>
              <option value="MEDIUM">中</option>
              <option value="HIGH">高</option>
            </select>

            {/* ソート */}
            <div className="flex items-center space-x-1">
              <select
                value={sort.field}
                onChange={(e) => handleSortChange(e.target.value)}
                className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="created_at">作成日</option>
                <option value="updated_at">更新日</option>
                <option value="due_date">期限</option>
                <option value="priority">優先度</option>
                <option value="title">タイトル</option>
              </select>
              <button
                onClick={() => onSortChange(sort.field, sort.direction === 'asc' ? 'desc' : 'asc')}
                className="p-2 border border-gray-300 rounded-md hover:bg-gray-50 text-sm"
                title={`${sort.direction === 'asc' ? '降順' : '昇順'}に変更`}
              >
                <ArrowUpDown className={cn(
                  "h-4 w-4 transition-transform",
                  sort.direction === 'desc' && "rotate-180"
                )} />
              </button>
            </div>

            {/* フィルター展開ボタン */}
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className={cn(
                "flex items-center px-3 py-2 border rounded-md text-sm transition-colors",
                isExpanded || activeFiltersCount > 0
                  ? "border-blue-500 bg-blue-50 text-blue-700"
                  : "border-gray-300 hover:bg-gray-50"
              )}
            >
              <Filter className="h-4 w-4 mr-1" />
              詳細フィルター
              {activeFiltersCount > 0 && (
                <span className="ml-1 bg-blue-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
                  {activeFiltersCount}
                </span>
              )}
            </button>
          </div>
        </div>
      </div>

      {/* 展開可能な詳細フィルター */}
      {isExpanded && (
        <div className="border-t border-gray-200 p-4">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {/* 担当者フィルター */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                <User className="inline h-4 w-4 mr-1" />
                担当者
              </label>
              <select
                value={filters.assignee_id || ''}
                onChange={(e) => handleFilterChange('assignee_id', e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
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

            {/* 作成者フィルター */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                作成者
              </label>
              <select
                value={filters.created_by || ''}
                onChange={(e) => handleFilterChange('created_by', e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">すべて</option>
                {users.map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.username}
                  </option>
                ))}
              </select>
            </div>

            {/* 期限プリセット */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                <Calendar className="inline h-4 w-4 mr-1" />
                期限
              </label>
              <select
                value=""
                onChange={(e) => e.target.value && handleDatePreset(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">期間を選択</option>
                <option value="today">今日</option>
                <option value="this_week">今週</option>
                <option value="overdue">期限切れ</option>
              </select>
            </div>

            {/* カスタム期限範囲 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                期限開始日
              </label>
              <input
                type="date"
                value={filters.due_date_from || ''}
                onChange={(e) => handleFilterChange('due_date_from', e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                期限終了日
              </label>
              <input
                type="date"
                value={filters.due_date_to || ''}
                onChange={(e) => handleFilterChange('due_date_to', e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          {/* フィルターアクション */}
          <div className="flex items-center justify-between mt-4 pt-4 border-t border-gray-200">
            <div className="text-sm text-gray-500">
              {activeFiltersCount > 0 && (
                <span>{activeFiltersCount}個のフィルターが適用中</span>
              )}
            </div>
            
            <div className="flex items-center space-x-2">
              <button
                onClick={handleClearFilters}
                className="px-3 py-1 text-sm text-gray-600 hover:text-gray-800 border border-gray-300 rounded-md hover:bg-gray-50"
              >
                すべてクリア
              </button>
              <button
                onClick={() => setIsExpanded(false)}
                className="px-3 py-1 text-sm text-blue-600 hover:text-blue-800"
              >
                閉じる
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default TaskFilter;