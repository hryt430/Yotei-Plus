import { api } from '../client';
import {
  FriendRequest,
  InvitationRequest,
  FriendshipResponse,
  FriendshipListResponse,
  InvitationResponse,
  InvitationListResponse,
  SocialStatsResponse,
  InviteURLResponse,
  ApiResponse,
  Pagination
} from '@/types';

// === Friend Management ===
export const socialApi = {
  // Friends
  sendFriendRequest: async (request: FriendRequest): Promise<FriendshipResponse> => {
    const response = await api.post('/social/friends/request', request);
    return response.data;
  },

  acceptFriendRequest: async (friendshipId: string): Promise<FriendshipResponse> => {
    const response = await api.put(`/social/friends/${friendshipId}/accept`);
    return response.data;
  },

  declineFriendRequest: async (friendshipId: string): Promise<ApiResponse> => {
    const response = await api.put(`/social/friends/${friendshipId}/decline`);
    return response.data;
  },

  removeFriend: async (userId: string): Promise<ApiResponse> => {
    const response = await api.delete(`/social/friends/${userId}`);
    return response.data;
  },

  blockUser: async (userId: string): Promise<ApiResponse> => {
    const response = await api.post(`/social/friends/${userId}/block`);
    return response.data;
  },

  unblockUser: async (userId: string): Promise<ApiResponse> => {
    const response = await api.delete(`/social/friends/${userId}/block`);
    return response.data;
  },

  getFriends: async (pagination?: Pagination): Promise<FriendshipListResponse> => {
    const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
    const response = await api.get('/social/friends', { params });
    return response.data;
  },

  getPendingRequests: async (pagination?: Pagination): Promise<FriendshipListResponse> => {
    const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
    const response = await api.get('/social/friends/pending', { params });
    return response.data;
  },

  getSentRequests: async (pagination?: Pagination): Promise<FriendshipListResponse> => {
    const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
    const response = await api.get('/social/friends/sent', { params });
    return response.data;
  },

  getMutualFriends: async (userId: string): Promise<FriendshipListResponse> => {
    const response = await api.get(`/social/friends/${userId}/mutual`);
    return response.data;
  },

  // Invitations
  createInvitation: async (request: InvitationRequest): Promise<InvitationResponse> => {
    const response = await api.post('/social/invitations', request);
    return response.data;
  },

  getInvitation: async (invitationId: string): Promise<InvitationResponse> => {
    const response = await api.get(`/social/invitations/${invitationId}`);
    return response.data;
  },

  getInvitationByCode: async (code: string): Promise<InvitationResponse> => {
    const response = await api.get(`/social/invitations/code/${code}`);
    return response.data;
  },

  acceptInvitation: async (code: string): Promise<ApiResponse> => {
    const response = await api.post(`/social/invitations/${code}/accept`);
    return response.data;
  },

  declineInvitation: async (invitationId: string): Promise<ApiResponse> => {
    const response = await api.put(`/social/invitations/${invitationId}/decline`);
    return response.data;
  },

  cancelInvitation: async (invitationId: string): Promise<ApiResponse> => {
    const response = await api.delete(`/social/invitations/${invitationId}`);
    return response.data;
  },

  getSentInvitations: async (pagination?: Pagination): Promise<InvitationListResponse> => {
    const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
    const response = await api.get('/social/invitations/sent', { params });
    return response.data;
  },

  getReceivedInvitations: async (pagination?: Pagination): Promise<InvitationListResponse> => {
    const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
    const response = await api.get('/social/invitations/received', { params });
    return response.data;
  },

  generateInviteURL: async (invitationId: string): Promise<InviteURLResponse> => {
    const response = await api.get(`/social/invitations/${invitationId}/url`);
    return response.data;
  },

  // Stats
  getSocialStats: async (): Promise<SocialStatsResponse> => {
    const response = await api.get('/social/stats');
    return response.data;
  }
};