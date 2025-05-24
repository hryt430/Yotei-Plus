import { useState, useEffect, useCallback } from 'react';
import { 
  getTasks, 
  getTaskById, 
  createTask, 
  updateTask, 
  deleteTask,
  assignTask
} from '@/api/task';
import { Task, TasksState } from '@/types';

export default function useTasks() {
  const [tasksState, setTasksState] = useState<TasksState>({
    tasks: [],
    isLoading: false,
    error: null,
    pagination: {
      page: 1,
      limit: 10,
      total: 0,
    },
    filters: {},
    sort: {
      field: 'dueDate',
      direction: 'asc',
    },
  });

  const fetchTasks = useCallback(async () => {
    setTasksState(prev => ({ ...prev, isLoading: true }));
    
    try {
      const { filters, pagination, sort } = tasksState;
      const { tasks, pagination: newPagination } = await getTasks({
        ...filters,
        page: pagination.page,
        limit: pagination.limit,
        sortBy: sort.field,
        sortDirection: sort.direction,
      });
      
      setTasksState(prev => ({
        ...prev,
        tasks,
        pagination: newPagination,
        isLoading: false,
        error: null,
      }));
      
      return tasks;
    } catch (error) {
      setTasksState(prev => ({ 
        ...prev, 
        isLoading: false, 
        error: (error as Error).message 
      }));
      throw error;
    }
  }, [tasksState.filters, tasksState.pagination.page, tasksState.pagination.limit, tasksState.sort]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const fetchTaskById = async (id: string) => {
    try {
      return await getTaskById(id);
    } catch (error) {
      throw error;
    }
  };

  const addTask = async (taskData: Partial<Task>) => {
    setTasksState(prev => ({ ...prev, isLoading: true }));
    
    try {
      const newTask = await createTask(taskData);
      
      setTasksState(prev => ({
        ...prev,
        tasks: [newTask, ...prev.tasks],
        isLoading: false,
      }));
      
      return newTask;
    } catch (error) {
      setTasksState(prev => ({ 
        ...prev, 
        isLoading: false, 
        error: (error as Error).message 
      }));
      throw error;
    }
  };

  const editTask = async (id: string, taskData: Partial<Task>) => {
    try {
      const updatedTask = await updateTask(id, taskData);
      
      setTasksState(prev => ({
        ...prev,
        tasks: prev.tasks.map(task => 
          task.id === id ? updatedTask : task
        ),
      }));
      
      return updatedTask;
    } catch (error) {
      throw error;
    }
  };

  const removeTask = async (id: string) => {
    try {
      await deleteTask(id);
      
      setTasksState(prev => ({
        ...prev,
        tasks: prev.tasks.filter(task => task.id !== id),
      }));
    } catch (error) {
      throw error;
    }
  };

  const assignTaskToUser = async (taskId: string, userId: string) => {
    try {
      const updatedTask = await assignTask(taskId, userId);
      
      setTasksState(prev => ({
        ...prev,
        tasks: prev.tasks.map(task => 
          task.id === taskId ? updatedTask : task
        ),
      }));
      
      return updatedTask;
    } catch (error) {
      throw error;
    }
  };

  const setFilters = (filters: TasksState['filters']) => {
    setTasksState(prev => ({
      ...prev,
      filters,
      pagination: {
        ...prev.pagination,
        page: 1, // フィルター変更時は1ページ目に戻す
      },
    }));
  };

  const setSorting = (field: string, direction: 'asc' | 'desc') => {
    setTasksState(prev => ({
      ...prev,
      sort: {
        field,
        direction,
      },
    }));
  };

  const setPage = (page: number) => {
    setTasksState(prev => ({
      ...prev,
      pagination: {
        ...prev.pagination,
        page,
      },
    }));
  };

  return {
    ...tasksState,
    fetchTasks,
    fetchTaskById,
    addTask,
    editTask,
    removeTask,
    assignTaskToUser,
    setFilters,
    setSorting,
    setPage,
  };
}