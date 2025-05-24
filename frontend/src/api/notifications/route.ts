// src/app/api/notifications/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth/next';
import { authOptions } from '../auth/[...nextauth]/route';

// 通知一覧を取得
export async function GET(req: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    // URLからクエリパラメータを取得
    const { searchParams } = new URL(req.url);
    const page = searchParams.get('page') || '1';
    const limit = searchParams.get('limit') || '10';
    const unreadOnly = searchParams.get('unreadOnly') || 'false';

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/notifications?page=${page}&limit=${limit}&unreadOnly=${unreadOnly}`, {
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || '通知の取得に失敗しました' }, { status: response.status });
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('通知取得エラー:', error);
    return NextResponse.json({ error: '通知の取得中にエラーが発生しました' }, { status: 500 });
  }
}

// 通知を既読にする
export async function PATCH(req: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    const { notificationIds } = await req.json();

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/notifications/mark-read`, {
      method: 'PATCH',
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ notificationIds }),
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || '通知の既読化に失敗しました' }, { status: response.status });
    }

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error('通知既読化エラー:', error);
    return NextResponse.json({ error: '通知の既読化中にエラーが発生しました' }, { status: 500 });
  }
}