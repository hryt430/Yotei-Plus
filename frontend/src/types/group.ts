import { User } from './user';

// === Group Types ===
export type GroupType = 'PUBLIC' | 'PRIVATE' | 'SECRET';
export type MemberRole = 'OWNER' | 'ADMIN' | 'MEMBER';

export interface Group {
  id: string;
  name: string;
  description?: string;
  type: GroupType;
  owner_id: string;
  settings?: GroupSettings;
  member_count: number;
  created_at: string;
  updated_at: string;
  owner?: User;
}

export interface GroupMember {
  id: string;
  group_id: string;
  user_id: string;
  role: MemberRole;
  joined_at: string;
  updated_at: string;
  user?: User;
}

export interface GroupWithMembers extends Group {
  members: GroupMember[];
}

export interface GroupSettings {
  allow_member_invite?: boolean;
  auto_accept_invites?: boolean;
  max_members?: number;
  [key: string]: any;
}

export interface GroupStats {
  total_members: number;
  active_members: number;
  total_tasks: number;
  completed_tasks: number;
  completion_rate: number;
}

// === Group Request Types ===
export interface CreateGroupRequest {
  name: string;
  description?: string;
  type: GroupType;
  settings?: GroupSettings;
}

export interface UpdateGroupRequest {
  name?: string;
  description?: string;
  settings?: GroupSettings;
}

export interface AddMemberRequest {
  user_id: string;
  role?: MemberRole;
}

export interface UpdateMemberRoleRequest {
  role: MemberRole;
}

// === Group API Response Types ===
export interface GroupResponse {
  success: boolean;
  message?: string;
  data: Group;
}

export interface GroupWithMembersResponse {
  success: boolean;
  data: GroupWithMembers;
}

export interface GroupListResponse {
  success: boolean;
  data: {
    groups: Group[];
    total: number;
    page: number;
    page_size: number;
  };
}

export interface GroupMemberListResponse {
  success: boolean;
  data: GroupMember[];
}

export interface GroupStatsResponse {
  success: boolean;
  data: GroupStats;
}

export interface SuccessResponse {
  success: boolean;
  message: string;
}

export interface ErrorResponse {
  error: string;
  message: string;
}

// === Group State Types ===
export interface GroupState {
  groups: Group[];
  currentGroup: GroupWithMembers | null;
  isLoading: boolean;
  error: string | null;
  pagination: {
    page: number;
    page_size: number;
    total: number;
  };
}

// === Constants ===
export const GROUP_TYPES: GroupType[] = ['PUBLIC', 'PRIVATE', 'SECRET'];
export const MEMBER_ROLES: MemberRole[] = ['OWNER', 'ADMIN', 'MEMBER'];

export const GROUP_TYPE_LABELS: Record<GroupType, string> = {
  PUBLIC: '公開',
  PRIVATE: '非公開',
  SECRET: '秘密'
};

export const MEMBER_ROLE_LABELS: Record<MemberRole, string> = {
  OWNER: 'オーナー',
  ADMIN: '管理者',
  MEMBER: 'メンバー'
};