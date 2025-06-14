import { api } from '../client';
import {
  CreateGroupRequest,
  UpdateGroupRequest,
  AddMemberRequest,
  UpdateMemberRoleRequest,
  GroupResponse,
  GroupWithMembersResponse,
  GroupListResponse,
  GroupMemberListResponse,
  GroupStatsResponse,
  SuccessResponse,
  GroupType,
  Pagination
} from '@/types';

export const groupApi = {
  // Group CRUD
  createGroup: async (request: CreateGroupRequest): Promise<GroupResponse> => {
    const response = await api.post('/groups', request);
    return response.data;
  },

  getGroup: async (groupId: string): Promise<GroupWithMembersResponse> => {
    const response = await api.get(`/groups/${groupId}`);
    return response.data;
  },

  updateGroup: async (groupId: string, request: UpdateGroupRequest): Promise<GroupResponse> => {
    const response = await api.put(`/groups/${groupId}`, request);
    return response.data;
  },

  deleteGroup: async (groupId: string): Promise<SuccessResponse> => {
    const response = await api.delete(`/groups/${groupId}`);
    return response.data;
  },

  // Group Lists
  getMyGroups: async (
    type?: GroupType,
    pagination?: Pagination
  ): Promise<GroupListResponse> => {
    const params: any = {};
    if (type) params.type = type;
    if (pagination) {
      params.page = pagination.page;
      params.page_size = pagination.page_size;
    }
    const response = await api.get('/groups/my', { params });
    return response.data;
  },

  searchGroups: async (
    query: string,
    type?: GroupType,
    pagination?: Pagination
  ): Promise<GroupListResponse> => {
    const params: any = { q: query };
    if (type) params.type = type;
    if (pagination) {
      params.page = pagination.page;
      params.page_size = pagination.page_size;
    }
    const response = await api.get('/groups/search', { params });
    return response.data;
  },

  // Member Management
  addMember: async (groupId: string, request: AddMemberRequest): Promise<SuccessResponse> => {
    const response = await api.post(`/groups/${groupId}/members`, request);
    return response.data;
  },

  removeMember: async (groupId: string, userId: string): Promise<SuccessResponse> => {
    const response = await api.delete(`/groups/${groupId}/members/${userId}`);
    return response.data;
  },

  updateMemberRole: async (
    groupId: string,
    userId: string,
    request: UpdateMemberRoleRequest
  ): Promise<SuccessResponse> => {
    const response = await api.put(`/groups/${groupId}/members/${userId}/role`, request);
    return response.data;
  },

  getMembers: async (groupId: string, pagination?: Pagination): Promise<GroupMemberListResponse> => {
    const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
    const response = await api.get(`/groups/${groupId}/members`, { params });
    return response.data;
  },

  // Stats
  getGroupStats: async (groupId: string): Promise<GroupStatsResponse> => {
    const response = await api.get(`/groups/${groupId}/stats`);
    return response.data;
  }
};