import { useState, useEffect, useCallback } from 'react';
import {
  getFriends,
  getPendingRequests,
  getSentRequests,
  getReceivedInvitations,
  getSocialStats,
  sendFriendRequest as apiSendFriendRequest,
  acceptFriendRequest as apiAcceptFriendRequest,
  declineFriendRequest as apiDeclineFriendRequest,
  removeFriend as apiRemoveFriend,
  blockUser as apiBlockUser,
  createInvitation as apiCreateInvitation,
  acceptInvitation as apiAcceptInvitation,
  generateInviteURL as apiGenerateInviteURL
} from '@/api/social';
import {
  SocialState,
  FriendRequest,
  InvitationRequest,
  Friendship,
  Invitation,
  Pagination
} from '@/types';
import { toast } from '@/hooks/use-toast';

export const useSocial = () => {
  const [state, setState] = useState<SocialState>({
    friends: [],
    friendRequests: [], 
    sentRequests: [],
    invitations: [],
    stats: null,
    isLoading: false,
    error: null
  });

  // Load initial data
  const loadData = useCallback(async () => {
    setState(prev => ({ ...prev, isLoading: true, error: null }));
    
    try {
      const [friendsRes, friendRequestsRes, sentRes, invitationsRes, statsRes] = await Promise.all([
        getFriends(),
        getPendingRequests(),
        getSentRequests(),
        getReceivedInvitations(),
        getSocialStats()
      ]);

      setState(prev => ({
        ...prev,
        friends: friendsRes.data,
        friendRequests: friendRequestsRes.data,
        sentRequests: sentRes.data,
        invitations: invitationsRes.data,
        stats: statsRes.data,
        isLoading: false
      }));
    } catch (error: any) {
      setState(prev => ({
        ...prev,
        error: error.message || 'データの読み込みに失敗しました',
        isLoading: false
      }));
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Friend actions
  const sendFriendRequest = useCallback(async (request: FriendRequest) => {
    try {
      const response = await apiSendFriendRequest(request);
      setState(prev => ({
        ...prev,
        sentRequests: [...prev.sentRequests, response.data]
      }));
      toast({
        title: '友達申請を送信しました',
        description: '相手の承認をお待ちください。'
      });
      return response.data;
    } catch (error: any) {
      toast({
        title: '友達申請の送信に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, []);

  const acceptFriendRequest = useCallback(async (friendshipId: string) => {
    try {
      const response = await apiAcceptFriendRequest(friendshipId);
      setState(prev => ({
        ...prev,
        friends: [...prev.friends, response.data],
        friendRequests: prev.friendRequests.filter(req => req.id !== friendshipId)
      }));
      toast({
        title: '友達申請を承認しました',
        description: '新しい友達が追加されました。'
      });
      return response.data;
    } catch (error: any) {
      toast({
        title: '友達申請の承認に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, []);

  const declineFriendRequest = useCallback(async (friendshipId: string) => {
    try {
      await apiDeclineFriendRequest(friendshipId);
      setState(prev => ({
        ...prev,
        friendRequests: prev.friendRequests.filter(req => req.id !== friendshipId)
      }));
      toast({
        title: '友達申請を拒否しました'
      });
    } catch (error: any) {
      toast({
        title: '友達申請の拒否に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, []);

  const removeFriend = useCallback(async (userId: string) => {
    try {
      await apiRemoveFriend(userId);
      setState(prev => ({
        ...prev,
        friends: prev.friends.filter(friend => 
          friend.requester_id !== userId && friend.addressee_id !== userId
        )
      }));
      toast({
        title: '友達を削除しました'
      });
    } catch (error: any) {
      toast({
        title: '友達の削除に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, []);

  const blockUser = useCallback(async (userId: string) => {
    try {
      await apiBlockUser(userId);
      setState(prev => ({
        ...prev,
        friends: prev.friends.filter(friend => 
          friend.requester_id !== userId && friend.addressee_id !== userId
        )
      }));
      toast({
        title: 'ユーザーをブロックしました'
      });
    } catch (error: any) {
      toast({
        title: 'ユーザーのブロックに失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, []);

  // Invitation actions
  const createInvitation = useCallback(async (request: InvitationRequest) => {
    try {
      const response = await apiCreateInvitation(request);
      toast({
        title: '招待を作成しました',
        description: '招待URLを共有してください。'
      });
      return response.data;
    } catch (error: any) {
      toast({
        title: '招待の作成に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, []);

  const acceptInvitation = useCallback(async (code: string) => {
    try {
      const response = await apiAcceptInvitation(code);
      await loadData(); // Refresh data
      toast({
        title: '招待を受諾しました'
      });
      return response;
    } catch (error: any) {
      toast({
        title: '招待の受諾に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, [loadData]);

  const generateInviteURL = useCallback(async (invitationId: string) => {
    try {
      const response = await apiGenerateInviteURL(invitationId);
      return response.data.url;
    } catch (error: any) {
      toast({
        title: '招待URLの生成に失敗しました',
        description: error.message,
        variant: 'destructive'
      });
      throw error;
    }
  }, []);

  return {
    ...state,
    actions: {
      loadData,
      sendFriendRequest,
      acceptFriendRequest,
      declineFriendRequest,
      removeFriend,
      blockUser,
      createInvitation,
      acceptInvitation,
      generateInviteURL
    }
  };
};