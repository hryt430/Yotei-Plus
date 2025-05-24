import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth/next';
import { authOptions } from '../auth/[...nextauth]/route';

// ユーザー一覧を取得（タスク割り当て等で使用）
export async function GET(req: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    // URLからクエリパラメータを取得
    const { searchParams } = new URL(req.url);
    const search = searchParams.get('search') || '';

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/users?search=${search}`, {
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || 'ユーザー一覧の取得に失敗しました' }, { status: response.status });
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('ユーザー一覧取得エラー:', error);
    return NextResponse.json({ error: 'ユーザー一覧の取得中にエラーが発生しました' }, { status: 500 });
  }
}