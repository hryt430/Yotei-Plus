// src/app/api/calendar/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth/next';
import { authOptions } from '../auth/[...nextauth]/route';

// カレンダーイベント（タスク期限）を取得
export async function GET(req: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    // URLからクエリパラメータを取得
    const { searchParams } = new URL(req.url);
    const startDate = searchParams.get('startDate');
    const endDate = searchParams.get('endDate');

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/tasks/calendar?${startDate ? `startDate=${startDate}` : ''}${endDate ? `&endDate=${endDate}` : ''}`, {
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || 'カレンダーイベントの取得に失敗しました' }, { status: response.status });
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('カレンダーイベント取得エラー:', error);
    return NextResponse.json({ error: 'カレンダーイベントの取得中にエラーが発生しました' }, { status: 500 });
  }
}