'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { 
  Bell, 
  Calendar, 
  CheckSquare, 
  Menu, 
  Settings, 
  User, 
  LogOut, 
  X,
  Home
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import NotificationDropdownProps from '@/components/notification/notification-dropdown';
import { useAuth } from '@/providers/auth-provider';
import { cn } from '@/lib/utils';

export default function Header() {
  const { user, isAuthenticated, logout } = useAuth();
  const pathname = usePathname();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isNotificationOpen, setIsNotificationOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);

  // ナビゲーションリンク
  const navLinks = [
    { 
      href: '/dashboard', 
      label: 'ダッシュボード', 
      icon: <Home className="h-5 w-5" />,
      pattern: '/dashboard'
    },
    { 
      href: '/tasks', 
      label: 'タスク', 
      icon: <CheckSquare className="h-5 w-5" />,
      pattern: '/tasks'
    },
    { 
      href: '/calendar', 
      label: 'カレンダー', 
      icon: <Calendar className="h-5 w-5" />,
      pattern: '/calendar'
    },
  ];

  // 未読通知数の取得（ダミー実装）
  useEffect(() => {
    if (isAuthenticated) {
      fetchUnreadNotificationsCount();
    }
  }, [isAuthenticated]);

  const fetchUnreadNotificationsCount = async () => {
    try {
      // 実際はAPIから取得
      // const response = await fetch('/api/notifications?unreadOnly=true&limit=1');
      // const data = await response.json();
      // setUnreadCount(data.totalCount || 0);
      
      // ダミーデータ
      setUnreadCount(2);
    } catch (error) {
      console.error('未読通知数の取得に失敗しました', error);
    }
  };

  const toggleMenu = () => {
    setIsMenuOpen(!isMenuOpen);
  };

  const toggleNotifications = () => {
    setIsNotificationOpen(!isNotificationOpen);
  };

  const handleSignOut = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  // パスが一致するかチェック
  const isActiveLink = (linkPattern: string) => {
    if (linkPattern === '/dashboard') {
      return pathname === '/dashboard';
    }
    return pathname.startsWith(linkPattern);
  };

  // 認証されていない場合のヘッダー
  if (!isAuthenticated) {
    return (
      <header className="sticky top-0 z-40 border-b bg-white shadow-sm">
        <div className="container flex h-16 items-center justify-between px-4">
          {/* ロゴ */}
          <div className="flex items-center">
            <Link href="/" className="flex items-center space-x-2">
              <CheckSquare className="h-6 w-6 text-blue-600" />
              <span className="text-xl font-bold">TaskMaster</span>
            </Link>
          </div>

          {/* 認証ボタン */}
          <div className="flex items-center space-x-2">
            <Button variant="ghost" asChild>
              <Link href="/auth/login">ログイン</Link>
            </Button>
            <Button asChild>
              <Link href="/auth/register">登録</Link>
            </Button>
          </div>
        </div>
      </header>
    );
  }

  return (
    <header className="sticky top-0 z-40 border-b bg-white shadow-sm">
      <div className="container flex h-16 items-center justify-between px-4">
        {/* ロゴ部分 */}
        <div className="flex items-center">
          <Link href="/dashboard" className="flex items-center space-x-2">
            <CheckSquare className="h-6 w-6 text-blue-600" />
            <span className="text-xl font-bold">TaskMaster</span>
          </Link>
        </div>

        {/* デスクトップ用ナビゲーション */}
        <nav className="hidden md:flex items-center space-x-1">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className={cn(
                "flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors",
                isActiveLink(link.pattern)
                  ? "bg-blue-100 text-blue-700"
                  : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
              )}
            >
              {link.icon}
              <span className="ml-2">{link.label}</span>
            </Link>
          ))}
        </nav>

        {/* 右側のアクション部分 */}
        <div className="flex items-center space-x-1">
          {/* 通知ボタン */}
          <div className="relative">
            <Button 
              variant="ghost" 
              size="icon" 
              onClick={toggleNotifications}
              className="relative"
            >
              <Bell className="h-5 w-5" />
              {unreadCount > 0 && (
                <span className="absolute -top-1 -right-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] text-white">
                  {unreadCount > 9 ? '9+' : unreadCount}
                </span>
              )}
            </Button>
            {isNotificationOpen && (
              <NotificationDropdownProps
                onClose={() => setIsNotificationOpen(false)} 
                updateUnreadCount={setUnreadCount}
              />
            )}
          </div>

          {/* 設定へのリンク */}
          <Button variant="ghost" size="icon" asChild>
            <Link href="/settings">
              <Settings className="h-5 w-5" />
            </Link>
          </Button>

          {/* ユーザーメニュー */}
          <div className="relative ml-2">
            <Button 
              variant="ghost" 
              size="icon" 
              className="rounded-full h-8 w-8 bg-gray-100 flex items-center justify-center hover:bg-gray-200"
              onClick={toggleMenu}
            >
              {user?.username?.charAt(0).toUpperCase() || <User className="h-4 w-4" />}
            </Button>
            
            {isMenuOpen && (
              <>
                {/* オーバーレイ */}
                <div 
                  className="fixed inset-0 z-10"
                  onClick={() => setIsMenuOpen(false)}
                />
                
                {/* ドロップダウンメニュー */}
                <div className="absolute right-0 mt-2 w-56 py-2 bg-white rounded-md shadow-lg border z-20">
                  <div className="px-4 py-3 border-b border-gray-100">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {user?.username}
                    </p>
                    <p className="text-xs text-gray-500 truncate">
                      {user?.email}
                    </p>
                    <p className="text-xs text-gray-400 mt-1">
                      {user?.role}
                    </p>
                  </div>
                  
                  <Link 
                    href="/profile" 
                    className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    <User className="h-4 w-4 mr-3" />
                    プロフィール
                  </Link>
                  
                  <Link 
                    href="/settings" 
                    className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    <Settings className="h-4 w-4 mr-3" />
                    設定
                  </Link>
                  
                  <div className="border-t border-gray-100 mt-2 pt-2">
                    <button 
                      className="w-full flex items-center px-4 py-2 text-sm text-red-600 hover:bg-red-50"
                      onClick={() => {
                        setIsMenuOpen(false);
                        handleSignOut();
                      }}
                    >
                      <LogOut className="h-4 w-4 mr-3" />
                      ログアウト
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>

          {/* モバイル用メニューボタン */}
          <Button
            variant="ghost"
            size="icon"
            className="md:hidden"
            onClick={toggleMenu}
          >
            {isMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </Button>
        </div>
      </div>

      {/* モバイル用ドロップダウンメニュー */}
      {isMenuOpen && (
        <div className="md:hidden border-t bg-white">
          <div className="container py-2 px-4">
            <nav className="flex flex-col space-y-1">
              {navLinks.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  className={cn(
                    "flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors",
                    isActiveLink(link.pattern)
                      ? "bg-blue-100 text-blue-700"
                      : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                  )}
                  onClick={() => setIsMenuOpen(false)}
                >
                  {link.icon}
                  <span className="ml-3">{link.label}</span>
                </Link>
              ))}
              
              <div className="border-t border-gray-200 my-2"></div>
              
              <Link
                href="/profile"
                className="flex items-center px-3 py-2 text-sm font-medium text-gray-600 hover:bg-gray-100 hover:text-gray-900 rounded-md transition-colors"
                onClick={() => setIsMenuOpen(false)}
              >
                <User className="h-5 w-5" />
                <span className="ml-3">プロフィール</span>
              </Link>
              
              <Link
                href="/settings"
                className="flex items-center px-3 py-2 text-sm font-medium text-gray-600 hover:bg-gray-100 hover:text-gray-900 rounded-md transition-colors"
                onClick={() => setIsMenuOpen(false)}
              >
                <Settings className="h-5 w-5" />
                <span className="ml-3">設定</span>
              </Link>
              
              <button
                className="flex items-center px-3 py-2 text-sm font-medium text-red-600 hover:bg-red-50 rounded-md transition-colors w-full text-left"
                onClick={() => {
                  setIsMenuOpen(false);
                  handleSignOut();
                }}
              >
                <LogOut className="h-5 w-5" />
                <span className="ml-3">ログアウト</span>
              </button>
            </nav>
          </div>
        </div>
      )}
    </header>
  );
}