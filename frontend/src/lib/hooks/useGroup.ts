import { useState, useEffect, useCallback } from 'react';
import { groupApi } from '@/api/group';
import {
  GroupState,
  CreateGroupRequest,
  UpdateGroupRequest,
  AddMemberRequest,
  UpdateMemberRoleRequest,
  GroupType,
  Pagination
} from '@/types';
import { toast } from '@/hooks/use-toast';

export const useGroup = () => {
  const [state, setState] = useState<GroupState>({
    groups: [],
    currentGroup: null,
    isLoading: false,
    error: null,
    pagination: {
      page: 1,
      page_size: 10,
      total: 0
    }
  });

  // Load my groups
  const loadMyGroups = useCallback(async (type?: GroupType, pagination?: Pagination) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }));
    
    try {
      const response = await groupApi.getMyGroups(type, pagination);
      setState(prev => ({
        ...prev,
        groups: response.data.groups,
        pagination: {
          page: response.data.page,
          page_size: response.data.page_size,
          total: response.data.total
        },
        isLoading: false
      }));
    } catch (error: any) {
      setState(prev => ({
        ...prev,
        error: error.message || 'グループの読み込みに失敗しました',
        isLoading: false
      }));
    }
  }, []);

  useEffect(() => {
    loadMyGroups();
  }, [loadMyGroups]);

  // Group actions
  const createGroup = useCallback(async (request: CreateGroupRequest) => {
    try {
      const response = await groupApi.createGroup(request);
      setState(prev => ({
        ...prev,
        groups: [response.data, ...prev.groups]
      }));
      toast({
        title: 'グループを作成しました',
        description: `「${response.data.name}」が作成されました。`
      });
      return response.data;
    } catch (error: any) {
      toast({
        title: 'グループの作成に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, []);

  const getGroup = useCallback(async (groupId: string) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }));
    
    try {
      const response = await groupApi.getGroup(groupId);
      setState(prev => ({
        ...prev,
        currentGroup: response.data,
        isLoading: false
      }));
      return response.data;
    } catch (error: any) {
      setState(prev => ({
        ...prev,
        error: error.message || 'グループの読み込みに失敗しました',
        isLoading: false
      }));
      throw error;
    }
  }, []);

  const updateGroup = useCallback(async (groupId: string, request: UpdateGroupRequest) => {
    try {
      const response = await groupApi.updateGroup(groupId, request);
      setState(prev => ({
        ...prev,
        groups: prev.groups.map(group => 
          group.id === groupId ? response.data : group
        ),
        currentGroup: prev.currentGroup?.id === groupId 
          ? { ...prev.currentGroup, ...response.data }
          : prev.currentGroup
      }));
      toast({
        title: 'グループを更新しました'
      });
      return response.data;
    } catch (error: any) {
      toast({
        title: 'グループの更新に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, []);

  const deleteGroup = useCallback(async (groupId: string) => {
    try {
      await groupApi.deleteGroup(groupId);
      setState(prev => ({
        ...prev,
        groups: prev.groups.filter(group => group.id !== groupId),
        currentGroup: prev.currentGroup?.id === groupId ? null : prev.currentGroup
      }));
      toast({
        title: 'グループを削除しました'
      });
    } catch (error: any) {
      toast({
        title: 'グループの削除に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, []);

  // Member actions
  const addMember = useCallback(async (groupId: string, request: AddMemberRequest) => {
    try {
      await groupApi.addMember(groupId, request);
      // Refresh current group if it's the same
      if (state.currentGroup?.id === groupId) {
        await getGroup(groupId);
      }
      toast({
        title: 'メンバーを追加しました'
      });
    } catch (error: any) {
      toast({
        title: 'メンバーの追加に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, [state.currentGroup?.id, getGroup]);

  const removeMember = useCallback(async (groupId: string, userId: string) => {
    try {
      await groupApi.removeMember(groupId, userId);
      // Refresh current group if it's the same
      if (state.currentGroup?.id === groupId) {
        await getGroup(groupId);
      }
      toast({
        title: 'メンバーを削除しました'
      });
    } catch (error: any) {
      toast({
        title: 'メンバーの削除に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, [state.currentGroup?.id, getGroup]);

  const updateMemberRole = useCallback(async (
    groupId: string, 
    userId: string, 
    request: UpdateMemberRoleRequest
  ) => {
    try {
      await groupApi.updateMemberRole(groupId, userId, request);
      // Refresh current group if it's the same
      if (state.currentGroup?.id === groupId) {
        await getGroup(groupId);
      }
      toast({
        title: 'メンバーの権限を更新しました'
      });
    } catch (error: any) {
      toast({
        title: 'メンバー権限の更新に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, [state.currentGroup?.id, getGroup]);

  const searchGroups = useCallback(async (
    query: string, 
    type?: GroupType, 
    pagination?: Pagination
  ) => {
    setState(prev => ({ ...prev, isLoading: true, error: null }));
    
    try {
      const response = await groupApi.searchGroups(query, type, pagination);
      setState(prev => ({
        ...prev,
        groups: response.data.groups,
        pagination: {
          page: response.data.page,
          page_size: response.data.page_size,
          total: response.data.total
        },
        isLoading: false
      }));
      return response.data.groups;
    } catch (error: any) {
      setState(prev => ({
        ...prev,
        error: error.message || 'グループの検索に失敗しました',
        isLoading: false
      }));
      throw error;
    }
  }, []);

  return {
    ...state,
    actions: {
      loadMyGroups,
      createGroup,
      getGroup,
      updateGroup,
      deleteGroup,
      addMember,
      removeMember,
      updateMemberRole,
      searchGroups
    }
  };
};