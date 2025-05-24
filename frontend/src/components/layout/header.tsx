import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useSession, signOut } from 'next-auth/react';
import { 
  Bell, 
  Calendar, 
  CheckSquare, 
  Menu, 
  Settings, 
  User, 
  LogOut, 
  X 
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import NotificationDropdown from '@/components/notification/notification-dropdown';

export default function Header() {
  const { data: session } = useSession();
  const pathname = usePathname();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isNotificationOpen, setIsNotificationOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);

  // ナビゲーションリンク
  const navLinks = [
    { href: '/dashboard', label: 'ダッシュボード', icon: <CheckSquare className="h-5 w-5" /> },
    { href: '/tasks', label: 'タスク', icon: <CheckSquare className="h-5 w-5" /> },
    { href: '/calendar', label: 'カレンダー', icon: <Calendar className="h-5 w-5" /> },
  ];

  // 未読通知数の取得
  useEffect(() => {
    if (session) {
      fetchUnreadNotificationsCount();
    }
  }, [session]);

  const fetchUnreadNotificationsCount = async () => {
    try {
      const response = await fetch('/api/notifications?unreadOnly=true&limit=1');
      
      if (response.ok) {
        const data = await response.json();
        setUnreadCount(data.totalCount || 0);
      }
    } catch (error) {
      console.error('未読通知数の取得に失敗しました', error);
    }
  };

  const toggleMenu = () => {
    setIsMenuOpen(!isMenuOpen);
  };

  const toggleNotifications = () => {
    setIsNotificationOpen(!isNotificationOpen);
    if (!isNotificationOpen) {
      // 通知ドロップダウンを開いた時に既読にする処理などを実装可能
    }
  };

  const handleSignOut = () => {
    signOut({ callbackUrl: '/' });
  };

  return (
    <header className="sticky top-0 z-40 border-b bg-background">
      <div className="container flex h-16 items-center justify-between px-4">
        {/* ロゴ部分 */}
        <div className="flex items-center">
          <Link href={session ? '/dashboard' : '/'} className="flex items-center">
            <span className="text-xl font-bold">TaskMaster</span>
          </Link>
        </div>

        {/* デスクトップ用ナビゲーション */}
        {session && (
          <nav className="hidden md:flex items-center space-x-4">
            {navLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className={`flex items-center px-3 py-2 text-sm font-medium rounded-md hover:bg-accent hover:text-accent-foreground transition-colors ${
                  pathname === link.href
                    ? "bg-accent text-accent-foreground"
                    : "text-foreground/60"
                }`}
              >
                {link.icon}
                <span className="ml-2">{link.label}</span>
              </Link>
            ))}
          </nav>
        )}

        {/* 右側のアクション部分 */}
        <div className="flex items-center space-x-1">
          {session ? (
            <>
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
                    <span className="absolute top-0 right-0 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] text-white">
                      {unreadCount > 9 ? '9+' : unreadCount}
                    </span>
                  )}
                </Button>
                {isNotificationOpen && (
                  <NotificationDropdown 
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

              {/* ユーザーアクション */}
              <div className="relative ml-2">
                <Button 
                  variant="ghost" 
                  size="icon" 
                  className="rounded-full h-8 w-8 bg-muted flex items-center justify-center"
                  onClick={toggleMenu}
                >
                  <User className="h-5 w-5" />
                </Button>
                
                {isMenuOpen && (
                  <div className="absolute right-0 mt-2 w-48 py-1 bg-background rounded-md shadow-lg border">
                    <div className="px-4 py-2 border-b">
                      <p className="text-sm font-medium truncate">{session.user?.name || session.user?.email}</p>
                      <p className="text-xs text-muted-foreground truncate">{session.user?.email}</p>
                    </div>
                    <Link 
                      href="/profile" 
                      className="flex items-center px-4 py-2 text-sm hover:bg-accent"
                      onClick={() => setIsMenuOpen(false)}
                    >
                      <User className="h-4 w-4 mr-2" />
                      プロフィール
                    </Link>
                    <button 
                      className="w-full flex items-center px-4 py-2 text-sm text-red-500 hover:bg-accent"
                      onClick={handleSignOut}
                    >
                      <LogOut className="h-4 w-4 mr-2" />
                      ログアウト
                    </button>
                  </div>
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
            </>
          ) : (
            <>
              <Button variant="ghost" asChild>
                <Link href="/login">ログイン</Link>
              </Button>
              <Button asChild>
                <Link href="/register">登録</Link>
              </Button>
            </>
          )}
        </div>
      </div>

      {/* モバイル用ドロップダウンメニュー */}
      {session && isMenuOpen && (
        <div className="md:hidden border-t">
          <div className="container py-2 px-4">
            <nav className="flex flex-col space-y-2">
              {navLinks.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  className={`flex items-center px-3 py-2 text-sm font-medium rounded-md hover:bg-accent hover:text-accent-foreground transition-colors ${
                    pathname === link.href
                      ? "bg-accent text-accent-foreground"
                      : "text-foreground/60"
                  }`}
                  onClick={() => setIsMenuOpen(false)}
                >
                  {link.icon}
                  <span className="ml-2">{link.label}</span>
                </Link>
              ))}
              
              <div className="border-t my-2"></div>
              
              <Link
                href="/profile"
                className="flex items-center px-3 py-2 text-sm font-medium rounded-md hover:bg-accent hover:text-accent-foreground transition-colors"
                onClick={() => setIsMenuOpen(false)}
              >
                <User className="h-5 w-5" />
                <span className="ml-2">プロフィール</span>
              </Link>
              <Link
                href="/settings"
                className="flex items-center px-3 py-2 text-sm font-medium rounded-md hover:bg-accent hover:text-accent-foreground transition-colors"
                onClick={() => setIsMenuOpen(false)}
              >
                <Settings className="h-5 w-5" />
                <span className="ml-2">設定</span>
              </Link>
              <button
                className="flex items-center px-3 py-2 text-sm font-medium rounded-md text-red-500 hover:bg-accent transition-colors w-full text-left"
                onClick={handleSignOut}
              >
                <LogOut className="h-5 w-5" />
                <span className="ml-2">ログアウト</span>
              </button>
            </nav>
          </div>
        </div>
      )}
    </header>
  );
}