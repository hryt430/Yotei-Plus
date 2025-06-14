import { GroupType, MemberRole, GroupMember } from '@/types';

// グループ関連のユーティリティ関数

export function formatGroupType(type: GroupType): string {
  switch (type) {
    case 'PUBLIC':
      return '公開';
    case 'PRIVATE':
      return '非公開';
    case 'SECRET':
      return '秘密';
    default:
      return '不明';
  }
}

export function formatMemberRole(role: MemberRole): string {
  switch (role) {
    case 'OWNER':
      return 'オーナー';
    case 'ADMIN':
      return '管理者';
    case 'MEMBER':
      return 'メンバー';
    default:
      return '不明';
  }
}

export function getGroupTypeIcon(type: GroupType): string {
  switch (type) {
    case 'PUBLIC':
      return '🌐';
    case 'PRIVATE':
      return '🔒';
    case 'SECRET':
      return '🔐';
    default:
      return '❓';
  }
}

export function getMemberRoleColor(role: MemberRole): string {
  switch (role) {
    case 'OWNER':
      return 'text-purple-600 bg-purple-100';
    case 'ADMIN':
      return 'text-blue-600 bg-blue-100';
    case 'MEMBER':
      return 'text-green-600 bg-green-100';
    default:
      return 'text-gray-600 bg-gray-100';
  }
}

export function getGroupTypeColor(type: GroupType): string {
  switch (type) {
    case 'PUBLIC':
      return 'text-green-600 bg-green-100';
    case 'PRIVATE':
      return 'text-blue-600 bg-blue-100';
    case 'SECRET':
      return 'text-red-600 bg-red-100';
    default:
      return 'text-gray-600 bg-gray-100';
  }
}

export function canManageMembers(userRole: MemberRole): boolean {
  return userRole === 'OWNER' || userRole === 'ADMIN';
}

export function canUpdateGroup(userRole: MemberRole): boolean {
  return userRole === 'OWNER' || userRole === 'ADMIN';
}

export function canDeleteGroup(userRole: MemberRole): boolean {
  return userRole === 'OWNER';
}

export function canChangeRole(currentUserRole: MemberRole, targetRole: MemberRole): boolean {
  // オーナーは全ての権限を変更可能
  if (currentUserRole === 'OWNER') {
    return true;
  }
  
  // 管理者は自分より下の権限のみ変更可能（オーナーは変更不可）
  if (currentUserRole === 'ADMIN') {
    return targetRole !== 'OWNER';
  }
  
  // メンバーは権限変更不可
  return false;
}

export function getRoleHierarchy(role: MemberRole): number {
  switch (role) {
    case 'OWNER':
      return 3;
    case 'ADMIN':
      return 2;
    case 'MEMBER':
      return 1;
    default:
      return 0;
  }
}

export function sortMembersByRole(members: GroupMember[]): GroupMember[] {
  return [...members].sort((a, b) => {
    const roleA = getRoleHierarchy(a.role);
    const roleB = getRoleHierarchy(b.role);
    
    if (roleA !== roleB) {
      return roleB - roleA; // 降順（オーナー -> 管理者 -> メンバー）
    }
    
    // 同じ権限の場合は参加日順
    return new Date(a.joined_at).getTime() - new Date(b.joined_at).getTime();
  });
}

export function formatMemberCount(count: number): string {
  if (count === 1) {
    return '1人';
  }
  return `${count}人`;
}

export function getGroupDescription(type: GroupType): string {
  switch (type) {
    case 'PUBLIC':
      return '誰でも検索・参加できるグループです';
    case 'PRIVATE':
      return '招待されたメンバーのみ参加できるグループです';
    case 'SECRET':
      return '招待されたメンバーのみ参加でき、検索にも表示されません';
    default:
      return '';
  }
}