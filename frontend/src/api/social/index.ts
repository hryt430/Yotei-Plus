import { apiClient } from '../client';
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

// フレンドリクエストを送信
export async function sendFriendRequest(request: FriendRequest): Promise<FriendshipResponse> {
  const response = await apiClient.post<FriendshipResponse>('/social/friends/request', request);
  return response;
}

// フレンドリクエストを承認
export async function acceptFriendRequest(friendshipId: string): Promise<FriendshipResponse> {
  const response = await apiClient.put<FriendshipResponse>(`/social/friends/${friendshipId}/accept`);
  return response;
}

// フレンドリクエストを拒否
export async function declineFriendRequest(friendshipId: string): Promise<ApiResponse> {
  const response = await apiClient.put<ApiResponse>(`/social/friends/${friendshipId}/decline`);
  return response;
}

// 友達を削除
export async function removeFriend(userId: string): Promise<ApiResponse> {
  const response = await apiClient.delete<ApiResponse>(`/social/friends/${userId}`);
  return response;
}

// ユーザーをブロック
export async function blockUser(userId: string): Promise<ApiResponse> {
  const response = await apiClient.post<ApiResponse>(`/social/friends/${userId}/block`);
  return response;
}

// ユーザーのブロックを解除
export async function unblockUser(userId: string): Promise<ApiResponse> {
  const response = await apiClient.delete<ApiResponse>(`/social/friends/${userId}/block`);
  return response;
}

// 友達一覧を取得
export async function getFriends(pagination?: Pagination): Promise<FriendshipListResponse> {
  const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
  const response = await apiClient.get<FriendshipListResponse>('/social/friends', params);
  return response;
}

// 受信した友達リクエスト一覧を取得
export async function getPendingRequests(pagination?: Pagination): Promise<FriendshipListResponse> {
  const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
  const response = await apiClient.get<FriendshipListResponse>('/social/friends/pending', params);
  return response;
}

// 送信した友達リクエスト一覧を取得
export async function getSentRequests(pagination?: Pagination): Promise<FriendshipListResponse> {
  const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
  const response = await apiClient.get<FriendshipListResponse>('/social/friends/sent', params);
  return response;
}

// 共通の友達を取得
export async function getMutualFriends(userId: string): Promise<FriendshipListResponse> {
  const response = await apiClient.get<FriendshipListResponse>(`/social/friends/${userId}/mutual`);
  return response;
}

// === Invitations ===

// 招待を作成
export async function createInvitation(request: InvitationRequest): Promise<InvitationResponse> {
  const response = await apiClient.post<InvitationResponse>('/social/invitations', request);
  return response;
}

// 招待情報を取得
export async function getInvitation(invitationId: string): Promise<InvitationResponse> {
  const response = await apiClient.get<InvitationResponse>(`/social/invitations/${invitationId}`);
  return response;
}

// コードで招待情報を取得
export async function getInvitationByCode(code: string): Promise<InvitationResponse> {
  const response = await apiClient.get<InvitationResponse>(`/social/invitations/code/${code}`);
  return response;
}

// 招待を承認
export async function acceptInvitation(code: string): Promise<ApiResponse> {
  const response = await apiClient.post<ApiResponse>(`/social/invitations/${code}/accept`);
  return response;
}

// 招待を拒否
export async function declineInvitation(invitationId: string): Promise<ApiResponse> {
  const response = await apiClient.put<ApiResponse>(`/social/invitations/${invitationId}/decline`);
  return response;
}

// 招待をキャンセル
export async function cancelInvitation(invitationId: string): Promise<ApiResponse> {
  const response = await apiClient.delete<ApiResponse>(`/social/invitations/${invitationId}`);
  return response;
}

// 送信した招待一覧を取得
export async function getSentInvitations(pagination?: Pagination): Promise<InvitationListResponse> {
  const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
  const response = await apiClient.get<InvitationListResponse>('/social/invitations/sent', params);
  return response;
}

// 受信した招待一覧を取得
export async function getReceivedInvitations(pagination?: Pagination): Promise<InvitationListResponse> {
  const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
  const response = await apiClient.get<InvitationListResponse>('/social/invitations/received', params);
  return response;
}

// 招待URLを生成
export async function generateInviteURL(invitationId: string): Promise<InviteURLResponse> {
  const response = await apiClient.get<InviteURLResponse>(`/social/invitations/${invitationId}/url`);
  return response;
}

// === Stats ===

// ソーシャル統計を取得
export async function getSocialStats(): Promise<SocialStatsResponse> {
  const response = await apiClient.get<SocialStatsResponse>('/social/stats');
  return response;
}

// === Search Functions ===

// ユーザーを検索（友達追加用）
export async function searchUsers(
  query: string,
  limit: number = 10
): Promise<ApiResponse<{
  users: Array<{
    id: string;
    username: string;
    email: string;
    role: 'user' | 'admin';
    mutualFriends?: number;
    relationshipStatus: 'none' | 'pending-sent' | 'pending-received' | 'friends';
  }>
}>> {
  const response = await apiClient.get<ApiResponse<{
    users: Array<{
      id: string;
      username: string;
      email: string;
      role: 'user' | 'admin';
      mutualFriends?: number;
      relationshipStatus: 'none' | 'pending-sent' | 'pending-received' | 'friends';
    }>
  }>>('/social/search/users', { q: query, limit: limit.toString() });
  return response;
}

// === Utility Functions ===

// 友達関係のステータスを確認
export async function getFriendshipStatus(userId: string): Promise<ApiResponse<{
  status: 'none' | 'pending-sent' | 'pending-received' | 'friends' | 'blocked';
  friendship_id?: string;
}>> {
  const response = await apiClient.get<ApiResponse<{
    status: 'none' | 'pending-sent' | 'pending-received' | 'friends' | 'blocked';
    friendship_id?: string;
  }>>(`/social/friends/${userId}/status`);
  return response;
}

// 友達の活動状況を取得
export async function getFriendsActivity(
  limit: number = 20
): Promise<ApiResponse<{
  activities: Array<{
    id: string;
    friend_id: string;
    friend_name: string;
    activity_type: 'task_completed' | 'project_joined' | 'achievement_unlocked';
    description: string;
    timestamp: string;
  }>
}>> {
  const response = await apiClient.get<ApiResponse<{
    activities: Array<{
      id: string;
      friend_id: string;
      friend_name: string;
      activity_type: 'task_completed' | 'project_joined' | 'achievement_unlocked';
      description: string;
      timestamp: string;
    }>
  }>>('/social/friends/activity', { limit: limit.toString() });
  return response;
}

// ブロックしたユーザー一覧を取得
export async function getBlockedUsers(pagination?: Pagination): Promise<FriendshipListResponse> {
  const params = pagination ? { page: pagination.page, page_size: pagination.page_size } : {};
  const response = await apiClient.get<FriendshipListResponse>('/social/friends/blocked', params);
  return response;
}

// 友達の推奨リストを取得
export async function getFriendSuggestions(
  limit: number = 10
): Promise<ApiResponse<{
  suggestions: Array<{
    id: string;
    username: string;
    email: string;
    mutualFriends: number;
    commonInterests: string[];
    suggestionScore: number;
  }>
}>> {
  const response = await apiClient.get<ApiResponse<{
    suggestions: Array<{
      id: string;
      username: string;
      email: string;
      mutualFriends: number;
      commonInterests: string[];
      suggestionScore: number;
    }>
  }>>('/social/friends/suggestions', { limit: limit.toString() });
  return response;
}