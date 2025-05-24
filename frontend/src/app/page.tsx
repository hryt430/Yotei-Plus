import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Calendar, Clock, CheckSquare, BellRing } from 'lucide-react';

export default function Home() {
  return (
    <div className="flex flex-col min-h-screen">
      {/* ヒーローセクション */}
      <section className="flex-1 flex flex-col items-center justify-center px-4 py-16 text-center bg-gradient-to-b from-white to-gray-50">
        <h1 className="text-4xl md:text-6xl font-bold tracking-tight text-gray-900">
          シンプルで効率的な<br />タスク管理アプリ
        </h1>
        <p className="mt-6 text-lg md:text-xl text-gray-600 max-w-2xl">
          期限管理、タスクの整理、チーム連携をスマートに。
          Notionのような美しいインターフェースでタスク管理を効率化します。
        </p>
        <div className="mt-10 flex flex-col sm:flex-row gap-4">
          <Button asChild size="lg" className="px-8">
            <Link href="/login">ログイン</Link>
          </Button>
          <Button asChild variant="outline" size="lg" className="px-8">
            <Link href="/register">新規登録</Link>
          </Button>
        </div>
      </section>

      {/* 特徴セクション */}
      <section className="py-16 px-4 bg-white">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-3xl font-bold text-center mb-12">主な機能</h2>
          <div className="grid md:grid-cols-3 gap-8">
            <FeatureCard
              icon={<CheckSquare className="h-10 w-10 text-blue-500" />}
              title="タスク管理"
              description="タスクの作成、編集、削除、フィルタリング、優先順位設定などを簡単に行えます。"
            />
            <FeatureCard
              icon={<Calendar className="h-10 w-10 text-green-500" />}
              title="カレンダー連携"
              description="期限を視覚的に管理し、予定の把握が容易になります。"
            />
            <FeatureCard
              icon={<BellRing className="h-10 w-10 text-purple-500" />}
              title="通知システム"
              description="重要な期限や更新を見逃さないように通知機能でお知らせします。"
            />
          </div>
        </div>
      </section>

      {/* フッター */}
      <footer className="py-8 px-4 bg-gray-50 border-t border-gray-200">
        <div className="max-w-6xl mx-auto text-center text-gray-600">
          <p>© 2025 タスク管理アプリ. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}

// 特徴カードコンポーネントの型定義を追加
interface FeatureCardProps {
  icon: React.ReactNode;
  title: string;
  description: string;
}


// 特徴カードコンポーネント
function FeatureCard({ icon, title, description }: FeatureCardProps) {
  return (
    <div className="flex flex-col items-center p-6 bg-gray-50 rounded-lg border border-gray-100 hover:shadow-md transition-shadow">
      <div className="mb-4">{icon}</div>
      <h3 className="text-xl font-semibold mb-2">{title}</h3>
      <p className="text-gray-600 text-center">{description}</p>
    </div>
  );
}