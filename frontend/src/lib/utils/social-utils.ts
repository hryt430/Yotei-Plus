import { FriendshipStatus, InvitationStatus, GroupType, MemberRole } from '@/types';

// ソーシャル関連のユーティリティ関数

export function formatFriendshipStatus(status: FriendshipStatus): string {
  switch (status) {
    case 'PENDING':
      return '保留中';
    case 'ACCEPTED':
      return '承認済み';
    case 'BLOCKED':
      return 'ブロック済み';
    default:
      return '不明';
  }
}

export function formatInvitationStatus(status: InvitationStatus): string {
  switch (status) {
    case 'PENDING':
      return '保留中';
    case 'ACCEPTED':
      return '承認済み';
    case 'DECLINED':
      return '拒否済み';
    case 'EXPIRED':
      return '期限切れ';
    case 'CANCELED':
      return 'キャンセル済み';
    default:
      return '不明';
  }
}

export function getInvitationStatusColor(status: InvitationStatus): string {
  switch (status) {
    case 'PENDING':
      return 'text-yellow-600';
    case 'ACCEPTED':
      return 'text-green-600';
    case 'DECLINED':
      return 'text-red-600';
    case 'EXPIRED':
      return 'text-gray-600';
    case 'CANCELED':
      return 'text-gray-600';
    default:
      return 'text-gray-600';
  }
}

export function getFriendshipStatusColor(status: FriendshipStatus): string {
  switch (status) {
    case 'PENDING':
      return 'text-yellow-600';
    case 'ACCEPTED':
      return 'text-green-600';
    case 'BLOCKED':
      return 'text-red-600';
    default:
      return 'text-gray-600';
  }
}

export function isInvitationExpired(expiresAt: string): boolean {
  return new Date(expiresAt) < new Date();
}

export function getTimeUntilExpiry(expiresAt: string): string {
  const now = new Date();
  const expiry = new Date(expiresAt);
  const diffMs = expiry.getTime() - now.getTime();
  
  if (diffMs <= 0) {
    return '期限切れ';
  }
  
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffHours / 24);
  
  if (diffDays > 0) {
    return `${diffDays}日後`;
  } else if (diffHours > 0) {
    return `${diffHours}時間後`;
  } else {
    const diffMinutes = Math.floor(diffMs / (1000 * 60));
    return `${diffMinutes}分後`;
  }
}

export function generateInviteText(type: 'FRIEND' | 'GROUP', inviterName: string, url: string): string {
  if (type === 'FRIEND') {
    return `${inviterName}さんから友達申請が届いています！\n\n以下のリンクから承認してください：\n${url}`;
  } else {
    return `${inviterName}さんからグループ招待が届いています！\n\n以下のリンクから参加してください：\n${url}`;
  }
}

export function copyInviteLink(url: string): Promise<boolean> {
  if (navigator.clipboard && window.isSecureContext) {
    return navigator.clipboard.writeText(url).then(() => true).catch(() => false);
  } else {
    // Fallback for older browsers
    const textArea = document.createElement('textarea');
    textArea.value = url;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
      document.execCommand('copy');
      textArea.remove();
      return Promise.resolve(true);
    } catch (err) {
      textArea.remove();
      return Promise.resolve(false);
    }
  }
}