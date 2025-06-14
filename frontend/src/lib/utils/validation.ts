import { GroupType, MemberRole, InvitationType, InvitationMethod } from '@/types';

// バリデーション関数

export function validateGroupName(name: string): string | null {
  if (!name || name.trim().length === 0) {
    return 'グループ名は必須です';
  }
  if (name.length > 100) {
    return 'グループ名は100文字以内で入力してください';
  }
  return null;
}

export function validateGroupDescription(description: string): string | null {
  if (description && description.length > 500) {
    return '説明は500文字以内で入力してください';
  }
  return null;
}

export function validateGroupType(type: string): string | null {
  const validTypes: GroupType[] = ['PUBLIC', 'PRIVATE', 'SECRET'];
  if (!validTypes.includes(type as GroupType)) {
    return '無効なグループタイプです';
  }
  return null;
}

export function validateMemberRole(role: string): string | null {
  const validRoles: MemberRole[] = ['OWNER', 'ADMIN', 'MEMBER'];
  if (!validRoles.includes(role as MemberRole)) {
    return '無効なメンバー権限です';
  }
  return null;
}

export function validateInvitationType(type: string): string | null {
  const validTypes: InvitationType[] = ['FRIEND', 'GROUP'];
  if (!validTypes.includes(type as InvitationType)) {
    return '無効な招待タイプです';
  }
  return null;
}

export function validateInvitationMethod(method: string): string | null {
  const validMethods: InvitationMethod[] = ['IN_APP', 'CODE', 'URL'];
  if (!validMethods.includes(method as InvitationMethod)) {
    return '無効な招待方法です';
  }
  return null;
}

export function validateInvitationMessage(message: string): string | null {
  if (message && message.length > 500) {
    return 'メッセージは500文字以内で入力してください';
  }
  return null;
}

export function validateInvitationExpiry(expiresHours: number): string | null {
  if (expiresHours < 1) {
    return '有効期限は1時間以上で設定してください';
  }
  if (expiresHours > 168) { // 1週間
    return '有効期限は168時間（1週間）以内で設定してください';
  }
  return null;
}

export function validateUserId(userId: string): string | null {
  if (!userId || userId.trim().length === 0) {
    return 'ユーザーIDは必須です';
  }
  // UUID形式の簡易チェック
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
  if (!uuidRegex.test(userId)) {
    return '無効なユーザーIDです';
  }
  return null;
}

export function validateGroupSettings(settings: any): string | null {
  if (!settings || typeof settings !== 'object') {
    return null; // 設定は任意
  }
  
  // max_membersの検証
  if (settings.max_members !== undefined) {
    const maxMembers = Number(settings.max_members);
    if (isNaN(maxMembers) || maxMembers < 1 || maxMembers > 1000) {
      return '最大メンバー数は1〜1000の範囲で設定してください';
    }
  }
  
  return null;
}

// 複合バリデーション
export interface GroupFormData {
  name: string;
  description?: string;
  type: GroupType;
  settings?: any;
}

export function validateGroupForm(data: GroupFormData): Record<string, string> {
  const errors: Record<string, string> = {};
  
  const nameError = validateGroupName(data.name);
  if (nameError) errors.name = nameError;
  
  const descriptionError = validateGroupDescription(data.description || '');
  if (descriptionError) errors.description = descriptionError;
  
  const typeError = validateGroupType(data.type);
  if (typeError) errors.type = typeError;
  
  const settingsError = validateGroupSettings(data.settings);
  if (settingsError) errors.settings = settingsError;
  
  return errors;
}

export interface InvitationFormData {
  type: InvitationType;
  method: InvitationMethod;
  target_id?: string;
  invitee_email?: string;
  message: string;
  expires_hours: number;
}

export function validateInvitationForm(data: InvitationFormData): Record<string, string> {
  const errors: Record<string, string> = {};
  
  const typeError = validateInvitationType(data.type);
  if (typeError) errors.type = typeError;
  
  const methodError = validateInvitationMethod(data.method);
  if (methodError) errors.method = methodError;
  
  if (data.type === 'GROUP' && (!data.target_id || data.target_id.trim().length === 0)) {
    errors.target_id = 'グループ招待にはグループIDが必要です';
  }
  
  if (data.invitee_email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(data.invitee_email)) {
      errors.invitee_email = '有効なメールアドレスを入力してください';
    }
  }
  
  const messageError = validateInvitationMessage(data.message);
  if (messageError) errors.message = messageError;
  
  const expiryError = validateInvitationExpiry(data.expires_hours);
  if (expiryError) errors.expires_hours = expiryError;
  
  return errors;
}