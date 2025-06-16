import { apiClient } from '../client';
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
  Pagination,
  ApiResponse
} from '@/types';

// === Group CRUD Operations ===

// グループを作成
export async function createGroup(request: CreateGroupRequest): Promise<GroupResponse> {
  const response = await apiClient.post<GroupResponse>('/groups', request);
  return response;
}

// グループ情報を取得（メンバー情報含む）
export async function getGroup(groupId: string): Promise<GroupWithMembersResponse> {
  const response = await apiClient.get<GroupWithMembersResponse>(`/groups/${groupId}`);
  return response;
}

// グループ情報を更新
export async function updateGroup(groupId: string, request: UpdateGroupRequest): Promise<GroupResponse> {
  const response = await apiClient.put<GroupResponse>(`/groups/${groupId}`, request);
  return response;
}

// グループを削除
export async function deleteGroup(groupId: string): Promise<SuccessResponse> {
  const response = await apiClient.delete<SuccessResponse>(`/groups/${groupId}`);
  return response;
}

// === Group Lists ===

// 自分が参加しているグループ一覧を取得
export async function getMyGroups(
  type?: GroupType,
  pagination?: Pagination
): Promise<GroupListResponse> {
  const params: Record<string, string> = {};
  if (type) params.type = type;
  if (pagination) {
    params.page = pagination.page.toString();
    params.page_size = pagination.page_size.toString();
  }
  const response = await apiClient.get<GroupListResponse>('/groups/my', params);
  return response;
}

// グループを検索
export async function searchGroups(
  query: string,
  type?: GroupType,
  pagination?: Pagination
): Promise<GroupListResponse> {
  const params: Record<string, string> = { q: query };
  if (type) params.type = type;
  if (pagination) {
    params.page = pagination.page.toString();
    params.page_size = pagination.page_size.toString();
  }
  const response = await apiClient.get<GroupListResponse>('/groups/search', params);
  return response;
}

// 公開グループ一覧を取得
export async function getPublicGroups(pagination?: Pagination): Promise<GroupListResponse> {
  const params: Record<string, string> = {};
  if (pagination) {
    params.page = pagination.page.toString();
    params.page_size = pagination.page_size.toString();
  }
  const response = await apiClient.get<GroupListResponse>('/groups/public', params);
  return response;
}

// === Member Management ===

// グループにメンバーを追加
export async function addMember(groupId: string, request: AddMemberRequest): Promise<SuccessResponse> {
  const response = await apiClient.post<SuccessResponse>(`/groups/${groupId}/members`, request);
  return response;
}

// グループからメンバーを削除
export async function removeMember(groupId: string, userId: string): Promise<SuccessResponse> {
  const response = await apiClient.delete<SuccessResponse>(`/groups/${groupId}/members/${userId}`);
  return response;
}

// メンバーの役割を更新
export async function updateMemberRole(
  groupId: string,
  userId: string,
  request: UpdateMemberRoleRequest
): Promise<SuccessResponse> {
  const response = await apiClient.put<SuccessResponse>(`/groups/${groupId}/members/${userId}/role`, request);
  return response;
}

// グループのメンバー一覧を取得
export async function getMembers(groupId: string, pagination?: Pagination): Promise<GroupMemberListResponse> {
  const params: Record<string, string> = {};
  if (pagination) {
    params.page = pagination.page.toString();
    params.page_size = pagination.page_size.toString();
  }
  const response = await apiClient.get<GroupMemberListResponse>(`/groups/${groupId}/members`, params);
  return response;
}

// グループから脱退
export async function leaveGroup(groupId: string): Promise<SuccessResponse> {
  const response = await apiClient.delete<SuccessResponse>(`/groups/${groupId}/leave`);
  return response;
}

// === Group Invitations ===

// グループに招待コードで参加
export async function joinGroupByCode(inviteCode: string): Promise<SuccessResponse> {
  const response = await apiClient.post<SuccessResponse>('/groups/join', { invite_code: inviteCode });
  return response;
}

// グループの招待コードを生成
export async function generateInviteCode(
  groupId: string,
  expiresIn?: number
): Promise<ApiResponse<{ invite_code: string; expires_at: string }>> {
  const data = expiresIn ? { expires_in: expiresIn } : {};
  const response = await apiClient.post<ApiResponse<{ invite_code: string; expires_at: string }>>(
    `/groups/${groupId}/invite-code`, 
    data
  );
  return response;
}

// グループの招待コードを無効化
export async function revokeInviteCode(groupId: string): Promise<SuccessResponse> {
  const response = await apiClient.delete<SuccessResponse>(`/groups/${groupId}/invite-code`);
  return response;
}

// === Stats and Analytics ===

// グループの統計情報を取得
export async function getGroupStats(groupId: string): Promise<GroupStatsResponse> {
  const response = await apiClient.get<GroupStatsResponse>(`/groups/${groupId}/stats`);
  return response;
}

// グループのアクティビティ履歴を取得
export async function getGroupActivity(
  groupId: string,
  limit: number = 20
): Promise<ApiResponse<{
  activities: Array<{
    id: string;
    type: 'member_joined' | 'member_left' | 'task_created' | 'task_completed' | 'role_changed';
    user_id: string;
    user_name: string;
    description: string;
    timestamp: string;
    metadata?: Record<string, any>;
  }>
}>> {
  const response = await apiClient.get<ApiResponse<{
    activities: Array<{
      id: string;
      type: 'member_joined' | 'member_left' | 'task_created' | 'task_completed' | 'role_changed';
      user_id: string;
      user_name: string;
      description: string;
      timestamp: string;
      metadata?: Record<string, any>;
    }>
  }>>(`/groups/${groupId}/activity`, { limit: limit.toString() });
  return response;
}

// === Group Settings ===

// グループ設定を取得
export async function getGroupSettings(groupId: string): Promise<ApiResponse<{
  allow_member_invite: boolean;
  auto_accept_invites: boolean;
  max_members: number;
  visibility: 'public' | 'private';
  join_approval_required: boolean;
}>> {
  const response = await apiClient.get<ApiResponse<{
    allow_member_invite: boolean;
    auto_accept_invites: boolean;
    max_members: number;
    visibility: 'public' | 'private';
    join_approval_required: boolean;
  }>>(`/groups/${groupId}/settings`);
  return response;
}

// グループ設定を更新
export async function updateGroupSettings(
  groupId: string,
  settings: {
    allow_member_invite?: boolean;
    auto_accept_invites?: boolean;
    max_members?: number;
    visibility?: 'public' | 'private';
    join_approval_required?: boolean;
  }
): Promise<SuccessResponse> {
  const response = await apiClient.put<SuccessResponse>(`/groups/${groupId}/settings`, settings);
  return response;
}

// === Advanced Features ===

// グループをお気に入りに追加
export async function favoriteGroup(groupId: string): Promise<SuccessResponse> {
  const response = await apiClient.post<SuccessResponse>(`/groups/${groupId}/favorite`);
  return response;
}

// グループのお気に入りを解除
export async function unfavoriteGroup(groupId: string): Promise<SuccessResponse> {
  const response = await apiClient.delete<SuccessResponse>(`/groups/${groupId}/favorite`);
  return response;
}

// お気に入りのグループ一覧を取得
export async function getFavoriteGroups(pagination?: Pagination): Promise<GroupListResponse> {
  const params: Record<string, string> = {};
  if (pagination) {
    params.page = pagination.page.toString();
    params.page_size = pagination.page_size.toString();
  }
  const response = await apiClient.get<GroupListResponse>('/groups/favorites', params);
  return response;
}

// グループのメンバー推奨リストを取得
export async function getMemberSuggestions(
  groupId: string,
  limit: number = 10
): Promise<ApiResponse<{
  suggestions: Array<{
    id: string;
    username: string;
    email: string;
    mutual_connections: number;
    relevance_score: number;
    reason: string;
  }>
}>> {
  const response = await apiClient.get<ApiResponse<{
    suggestions: Array<{
      id: string;
      username: string;
      email: string;
      mutual_connections: number;
      relevance_score: number;
      reason: string;
    }>
  }>>(`/groups/${groupId}/member-suggestions`, { limit: limit.toString() });
  return response;
}

// === Batch Operations ===

// 複数のグループ情報を一括取得
export async function getGroupsBatch(groupIds: string[]): Promise<ApiResponse<{
  groups: Array<any>; // Group型
  not_found: string[];
}>> {
  const response = await apiClient.post<ApiResponse<{
    groups: Array<any>;
    not_found: string[];
  }>>('/groups/batch', { group_ids: groupIds });
  return response;
}

// === Utility Functions ===

// グループの権限を確認
export async function checkGroupPermissions(
  groupId: string
): Promise<ApiResponse<{
  can_edit: boolean;
  can_delete: boolean;
  can_add_members: boolean;
  can_remove_members: boolean;
  can_manage_roles: boolean;
  role: 'OWNER' | 'ADMIN' | 'MEMBER';
}>> {
  const response = await apiClient.get<ApiResponse<{
    can_edit: boolean;
    can_delete: boolean;
    can_add_members: boolean;
    can_remove_members: boolean;
    can_manage_roles: boolean;
    role: 'OWNER' | 'ADMIN' | 'MEMBER';
  }>>(`/groups/${groupId}/permissions`);
  return response;
}

// グループの参加可能性を確認
export async function checkGroupJoinability(groupId: string): Promise<ApiResponse<{
  can_join: boolean;
  reason?: string;
  requires_approval: boolean;
  is_full: boolean;
  is_member: boolean;
}>> {
  const response = await apiClient.get<ApiResponse<{
    can_join: boolean;
    reason?: string;
    requires_approval: boolean;
    is_full: boolean;
    is_member: boolean;
  }>>(`/groups/${groupId}/joinability`);
  return response;
}