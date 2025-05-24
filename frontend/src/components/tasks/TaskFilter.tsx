import React, { useState } from 'react';
import { User } from '@/types';

interface TaskFilterProps {
  users?: User[];
  onFilter: (filters: {
    status?: string[];
    priority?: string[];
    assignedTo?: string;
    search?: string;
    startDate?: string;
    endDate?: string;
    tags?: string[];
  }) => void;
  onSort: (field: string, direction: 'asc' | 'desc') => void;
}

const TaskFilter: React.FC<TaskFilterProps> = ({ users = [], onFilter, onSort }) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [filters, setFilters] = useState({
    status: [] as string[],
    priority: [] as string[],
    assignedTo: '',
    search: '',
    startDate: '',
    endDate: '',
    tags: [] as string[],
  });
  const [sortField, setSortField] = useState('dueDate');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
  const [tagInput, setTagInput] = useState('');

  const handleCheckboxChange = (filterType: 'status' | 'priority', value: string) => {
    setFilters(prev => {
      const currentValues = [...prev[filterType]];
      const valueIndex = currentValues.indexOf(value);
      
      if (valueIndex === -1) {
        currentValues.push(value);
      } else {
        currentValues.splice(valueIndex, 1);
      }
      
      return {
        ...prev,
        [filterType]: currentValues,
      };
    });
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFilters(prev => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleTagAdd = () => {
    if (tagInput.trim()) {
      setFilters(prev => ({
        ...prev,
        tags: [...prev.tags, tagInput.trim()],
      }));
      setTagInput('');
    }
  };

  const handleTagRemove = (tag: string) => {
    setFilters(prev => ({
      ...prev,
      tags: prev.tags.filter(t => t !== tag),
    }));
  };

  const handleSortChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setSortField(e.target.value);
    onSort(e.target.value, sortDirection);
  };

  const toggleSortDirection = () => {
    const newDirection = sortDirection === 'asc' ? 'desc' : 'asc';
    setSortDirection(newDirection);
    onSort(sortField, newDirection);
  };

  const applyFilters = () => {
    onFilter(filters);
  };

  const resetFilters = () => {
    setFilters({
      status: [],
      priority: [],
      assignedTo: '',
      search: '',
      startDate: '',
      endDate: '',
      tags: [],
    });
    setSortField('dueDate');
    setSortDirection('asc');
    onFilter({});
    onSort('dueDate', 'asc');
  };

  return (
    <div className="bg-white shadow rounded-lg p-4 mb-4">
      <div className="flex items-center justify-between mb-2">
        <div className="relative w-full md:w-1/3 mr-4">
          <input
            type="text"
            name="search"
            value={filters.search}
            onChange={handleInputChange}
            placeholder="タスクを検索..."
            className="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            onClick={applyFilters}
            className="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-gray-700"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
            </svg>
          </button>
        </div>
        <div className="flex items-center space-x-2">
          <select
            value={sortField}
            onChange={handleSortChange}
            className="px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="dueDate">期日</option>
            <option value="priority">優先度</option>
            <option value="title">タイトル</option>
            <option value="createdAt">作成日</option>
          </select>
          <button
            onClick={toggleSortDirection}
            className="px-3 py-2 border rounded-lg focus:outline-none hover:bg-gray-50"
          >
            {sortDirection === 'asc' ? (
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd" />
              </svg>
            ) : (
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M14.707 12.707a1 1 0 01-1.414 0L10 9.414l-3.293 3.293a1 1 0 01-1.414-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 010 1.414z" clipRule="evenodd" />
              </svg>
            )}
          </button>
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="px-3 py-2 border rounded-lg focus:outline-none hover:bg-gray-50"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M3 3a1 1 0 011-1h12a1 1 0 011 1v3a1 1 0 01-.293.707L12 11.414V15a1 1 0 01-.293.707l-2 2A1 1 0 018 17v-5.586L3.293 6.707A1 1 0 013 6V3z" clipRule="evenodd" />
            </svg>
          </button>
        </div>
      </div>

      {isExpanded && (
        <div className="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <h4 className="font-medium mb-2">ステータス</h4>
            <div className="space-y-1">
              {['未着手', '進行中', '完了', '延期'].map((status) => (
                <div key={status} className="flex items-center">
                  <input
                    type="checkbox"
                    id={`status-${status}`}
                    checked={filters.status.includes(status)}
                    onChange={() => handleCheckboxChange('status', status)}
                    className="mr-2"
                  />
                  <label htmlFor={`status-${status}`}>{status}</label>
                </div>
              ))}
            </div>
          </div>

          <div>
            <h4 className="font-medium mb-2">優先度</h4>
            <div className="space-y-1">
              {['低', '中', '高', '緊急'].map((priority) => (
                <div key={priority} className="flex items-center">
                  <input
                    type="checkbox"
                    id={`priority-${priority}`}
                    checked={filters.priority.includes(priority)}
                    onChange={() => handleCheckboxChange('priority', priority)}
                    className="mr-2"
                  />
                  <label htmlFor={`priority-${priority}`}>{priority}</label>
                </div>
              ))}
            </div>
          </div>

          <div>
            <h4 className="font-medium mb-2">担当者</h4>
            <select
              name="assignedTo"
              value={filters.assignedTo}
              onChange={handleInputChange}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">すべて</option>
              {users.map((user) => (
                <option key={user.id} value={user.id}>
                  {user.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <h4 className="font-medium mb-2">期間</h4>
            <div className="flex items-center space-x-2">
              <input
                type="date"
                name="startDate"
                value={filters.startDate}
                onChange={handleInputChange}
                className="flex-1 px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <span>〜</span>
              <input
                type="date"
                name="endDate"
                value={filters.endDate}
                onChange={handleInputChange}
                className="flex-1 px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          <div className="md:col-span-2">
            <h4 className="font-medium mb-2">タグ</h4>
            <div className="flex flex-wrap items-center gap-2 mb-2">
              {filters.tags.map((tag) => (
                <span
                  key={tag}
                  className="bg-blue-100 text-blue-800 px-2 py-1 rounded-full text-sm flex items-center"
                >
                  {tag}
                  <button
                    onClick={() => handleTagRemove(tag)}
                    className="ml-1 text-blue-500 hover:text-blue-700"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                    </svg>
                  </button>
                </span>
              ))}
            </div>
            <div className="flex items-center">
              <input
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                placeholder="タグを追加..."
                className="flex-1 px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                onKeyPress={(e) => e.key === 'Enter' && handleTagAdd()}
              />
              <button
                onClick={handleTagAdd}
                className="ml-2 px-3 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                追加
              </button>
            </div>
          </div>

          <div className="md:col-span-3 flex justify-end space-x-2 mt-4">
            <button
              onClick={resetFilters}
              className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              リセット
            </button>
            <button
              onClick={applyFilters}
              className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              フィルター適用
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default TaskFilter;