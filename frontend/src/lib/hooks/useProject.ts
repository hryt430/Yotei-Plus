import { useState, useCallback, useMemo } from 'react';
import { useGroup } from '@/lib/hooks/useGroup';
import useTask from '@/lib/hooks/useTask';
import {
  ProjectView,
  ProjectStats,
  ProjectMember,
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectTaskRequest,
  ProjectState,
  ProjectFilter,
  ProjectSort,
  GanttChartData,
  TaskHierarchy
} from '@/types';
import {
  ProjectTask,
  Task,
  TaskType,
  isProjectTask
} from '@/types/task';
import {
  Group,
  GroupWithMembers,
  CreateGroupRequest,
  UpdateGroupRequest
} from '@/types/group';

export const useProject = () => {
  // 既存フックの活用
  const groupHook = useGroup();
  const taskHook = useTask();
  
  // プロジェクト固有の状態
  const [projectState, setProjectState] = useState<ProjectState>({
    projects: [],
    currentProject: null,
    isLoading: false,
    error: null,
    filters: {},
    sort: { field: 'created_at', direction: 'desc' }
  });

  // === プロジェクトビュー構築 ===
  const buildProjectView = useCallback((group: GroupWithMembers): ProjectView => {
    // group_idでフィルタしたプロジェクトタスクを取得
    const projectTasks = taskHook.tasks.filter(
      (task): task is ProjectTask => 
        task.task_type === 'PROJECT' && task.group_id === group.id
    );
    
    // プロジェクト統計を計算
    const stats = calculateProjectStats(projectTasks, group.members);
    
    // プロジェクトメンバー情報を拡張
    const members = enhanceProjectMembers(group.members, projectTasks);
    
    return {
      group,
      tasks: projectTasks,
      stats,
      members,
    };
  }, [taskHook.tasks]);

  // === プロジェクト統計計算 ===
  const calculateProjectStats = useCallback((
    tasks: ProjectTask[],
    members: any[]
  ): ProjectStats => {
    const totalTasks = tasks.length;
    const completedTasks = tasks.filter(t => t.status === 'DONE').length;
    const overdueTasks = tasks.filter(t => t.is_overdue).length;
    const tasksInProgress = tasks.filter(t => t.status === 'IN_PROGRESS').length;
    const todoTasks = tasks.filter(t => t.status === 'TODO').length;
    
    // 進捗計算
    const totalProgress = tasks.reduce((sum, task) => sum + (task.progress || 0), 0);
    const averageProgress = totalTasks > 0 ? totalProgress / totalTasks : 0;
    const completionRate = totalTasks > 0 ? (completedTasks / totalTasks) * 100 : 0;
    
    // 推定完了日計算 (簡易版)
    const estimatedCompletion = calculateEstimatedCompletion(tasks);
    const daysRemaining = estimatedCompletion 
      ? Math.ceil((estimatedCompletion.getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24))
      : undefined;
    
    return {
      totalTasks,
      completedTasks,
      overdueTasks,
      tasksInProgress,
      todoTasks,
      totalMembers: members.length,
      activeMembers: members.filter(m => m.last_active).length,
      averageProgress,
      completionRate,
      estimatedCompletion,
      daysRemaining,
    };
  }, []);

  // === プロジェクトメンバー拡張 ===
  const enhanceProjectMembers = useCallback((
    members: any[],
    tasks: ProjectTask[]
  ): ProjectMember[] => {
    return members.map(member => {
      const memberTasks = tasks.filter(t => t.assignee_id === member.user_id);
      const completedTasks = memberTasks.filter(t => t.status === 'DONE').length;
      const inProgressTasks = memberTasks.filter(t => t.status === 'IN_PROGRESS').length;
      const completionRate = memberTasks.length > 0 
        ? (completedTasks / memberTasks.length) * 100 
        : 0;
      
      // 最後のタスク更新日時
      const lastTaskUpdate = memberTasks
        .map(t => new Date(t.updated_at))
        .sort((a, b) => b.getTime() - a.getTime())[0];
      
      return {
        ...member,
        tasksAssigned: memberTasks.length,
        tasksCompleted: completedTasks,
        tasksInProgress: inProgressTasks,
        completionRate,
        lastTaskUpdate,
      };
    });
  }, []);

  // === プロジェクト操作 ===
  
  // プロジェクト作成
  const createProject = useCallback(async (request: CreateProjectRequest) => {
    setProjectState(prev => ({ ...prev, isLoading: true, error: null }));
    
    try {
      // GroupをPROJECTタイプで作成
      const groupRequest: CreateGroupRequest = {
        name: request.name,
        description: request.description,
        type: 'PRIVATE', // プロジェクトは基本的にプライベート
        settings: {
          allow_member_invite: true,
          auto_accept_invites: false,
          max_members: 50,
          ...request.settings,
        }
      };
      
      const createdGroup = await groupHook.actions.createGroup(groupRequest);
      
      // プロジェクトビューを構築
      const projectView = buildProjectView(createdGroup as GroupWithMembers);
      
      setProjectState(prev => ({
        ...prev,
        projects: [projectView, ...prev.projects],
        isLoading: false,
      }));
      
      return projectView;
    } catch (error: any) {
      setProjectState(prev => ({
        ...prev,
        error: error.message,
        isLoading: false,
      }));
      throw error;
    }
  }, [groupHook.actions.createGroup, buildProjectView]);

  // プロジェクト取得
  const getProject = useCallback(async (projectId: string) => {
    setProjectState(prev => ({ ...prev, isLoading: true, error: null }));
    
    try {
      const group = await groupHook.actions.getGroup(projectId);
      const projectView = buildProjectView(group);
      
      setProjectState(prev => ({
        ...prev,
        currentProject: projectView,
        isLoading: false,
      }));
      
      return projectView;
    } catch (error: any) {
      setProjectState(prev => ({
        ...prev,
        error: error.message,
        isLoading: false,
      }));
      throw error;
    }
  }, [groupHook.actions.getGroup, buildProjectView]);

  // プロジェクトタスク作成
  const createProjectTask = useCallback(async (
    projectId: string,
    taskData: Omit<ProjectTaskRequest, 'group_id' | 'task_type'>
  ) => {
    const projectTaskData = {
      ...taskData,
      task_type: 'PROJECT' as const,
      group_id: projectId,
    };
    
    const createdTask = await taskHook.addTask(projectTaskData);
    
    // 現在のプロジェクトを更新
    if (projectState.currentProject?.group.id === projectId) {
      const updatedProject = buildProjectView(projectState.currentProject.group);
      setProjectState(prev => ({
        ...prev,
        currentProject: updatedProject,
      }));
    }
    
    return createdTask;
  }, [taskHook.addTask, buildProjectView, projectState.currentProject]);

  // === ガントチャート用データ生成 ===
  const generateGanttData = useCallback((tasks: ProjectTask[]): GanttChartData[] => {
    return tasks
      .filter(task => task.start_date && task.end_date)
      .map(task => ({
        task,
        startDate: new Date(task.start_date!),
        endDate: new Date(task.end_date!),
        progress: task.progress || 0,
        dependencies: task.dependencies || [],
        critical: false, // TODO: クリティカルパス計算
      }));
  }, []);

  // === タスク階層構造生成 ===
  const generateTaskHierarchy = useCallback((tasks: ProjectTask[]): TaskHierarchy[] => {
    const rootTasks = tasks.filter(task => !task.parent_task_id);
    
    const buildHierarchy = (parentTask: ProjectTask): TaskHierarchy => {
      const children = tasks
        .filter(task => task.parent_task_id === parentTask.id)
        .map(buildHierarchy);
      
      return {
        task: parentTask,
        children,
        level: parentTask.level || 0,
        canStart: parentTask.can_start || false,
      };
    };
    
    return rootTasks.map(buildHierarchy);
  }, []);

  // === 推定完了日計算 ===
  const calculateEstimatedCompletion = useCallback((tasks: ProjectTask[]): Date | undefined => {
    const tasksWithEndDate = tasks.filter(task => task.end_date);
    if (tasksWithEndDate.length === 0) return undefined;
    
    // 最新の終了日を取得
    const latestEndDate = tasksWithEndDate
      .map(task => new Date(task.end_date!))
      .sort((a, b) => b.getTime() - a.getTime())[0];
    
    return latestEndDate;
  }, []);

  // === 現在のプロジェクト一覧を取得 ===
  const projects = useMemo(() => {
    return groupHook.groups
      .filter(group => group.type === 'PRIVATE') // プロジェクトはPRIVATEタイプ
      .map(group => buildProjectView(group as GroupWithMembers));
  }, [groupHook.groups, buildProjectView]);

  // === フィルター・ソート ===
  const setFilters = useCallback((filters: ProjectFilter) => {
    setProjectState(prev => ({ ...prev, filters }));
  }, []);

  const setSort = useCallback((sort: ProjectSort) => {
    setProjectState(prev => ({ ...prev, sort }));
  }, []);

  return {
    // データ
    projects,
    currentProject: projectState.currentProject,
    isLoading: projectState.isLoading || groupHook.isLoading || taskHook.isLoading,
    error: projectState.error || groupHook.error || taskHook.error,
    
    // プロジェクト操作
    createProject,
    updateProject: groupHook.actions.updateGroup,
    deleteProject: groupHook.actions.deleteGroup,
    getProject,
    
    // プロジェクトタスク操作
    createProjectTask,
    updateProjectTask: taskHook.editTask,
    deleteProjectTask: taskHook.removeTask,
    updateTaskStatus: taskHook.updateTaskStatus,
    
    // メンバー操作
    addMember: groupHook.actions.addMember,
    removeMember: groupHook.actions.removeMember,
    updateMemberRole: groupHook.actions.updateMemberRole,
    
    // ユーティリティ
    buildProjectView,
    generateGanttData,
    generateTaskHierarchy,
    calculateProjectStats,
    
    // フィルター・ソート
    filters: projectState.filters,
    sort: projectState.sort,
    setFilters,
    setSort,
    
    // 状態リセット
    clearError: () => setProjectState(prev => ({ ...prev, error: null })),
  };
};