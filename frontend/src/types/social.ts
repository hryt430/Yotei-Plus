import { User } from './user';

// === Social Types ===
export type FriendshipStatus = 'PENDING' | 'ACCEPTED' | 'BLOCKED';
export type InvitationType = 'FRIEND' | 'GROUP';
export type InvitationMethod = 'IN_APP' | 'CODE' | 'URL';
export type InvitationStatus = 'PENDING' | 'ACCEPTED' | 'DECLINED' | 'EXPIRED' | 'CANCELED';

export interface Friendship {
  id: string;
  requester_id: string;
  addressee_id: string;
  status: FriendshipStatus;
  created_at: string;
  updated_at: string;
  accepted_at?: string;
  blocked_at?: string;
  requester?: User;
  addressee?: User;
}

export interface Invitation {
  id: string;
  type: InvitationType;
  method: InvitationMethod;
  status: InvitationStatus;
  inviter_id: string;
  invitee_id?: string;
  invitee_email?: string;
  invitee_username?: string;
  invitee_phone?: string;
  target_id?: string; // グループ招待の場合のグループID
  code?: string;
  url?: string;
  message: string;
  metadata?: Record<string, any>;
  expires_at: string;
  created_at: string;
  updated_at: string;
  accepted_at?: string;
  inviter?: User;
  invitee?: User;
}

export interface InviteeInfo {
  email?: string;
  username?: string;
  phone?: string;
}

export interface FriendRequest {
  user_id: string;
  message?: string;
}

export interface InvitationRequest {
  type: InvitationType;
  method: InvitationMethod;
  target_id?: string;
  invitee_email?: string;
  message: string;
  expires_hours: number;
}

export interface SocialStats {
  total_friends: number;
  pending_requests: number;
  sent_requests: number;
  total_invitations_sent: number;
  total_invitations_received: number;
}

// === Social API Response Types ===
export interface FriendshipResponse {
  success: boolean;
  message?: string;
  data: Friendship;
}

export interface FriendshipListResponse {
  success: boolean;
  data: Friendship[];
  pagination?: {
    page: number;
    page_size: number;
    total: number;
  };
}

export interface InvitationResponse {
  success: boolean;
  message?: string;
  data: Invitation;
}

export interface InvitationListResponse {
  success: boolean;
  data: Invitation[];
  pagination?: {
    page: number;
    page_size: number;
    total: number;
  };
}

export interface SocialStatsResponse {
  success: boolean;
  data: SocialStats;
}

export interface InviteURLResponse {
  success: boolean;
  data: {
    url: string;
  };
  message?: string;
}

// === Social State Types ===
export interface SocialState {
  friends: Friendship[];
  pendingRequests: Friendship[];
  sentRequests: Friendship[];
  invitations: Invitation[];
  stats: SocialStats | null;
  isLoading: boolean;
  error: string | null;
}