import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth/next';
import { authOptions } from '../../auth/[...nextauth]/route';

// 現在のユーザー情報を取得
export async function GET() {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/users/me`, {
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || 'ユーザー情報の取得に失敗しました' }, { status: response.status });
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('ユーザー情報取得エラー:', error);
    return NextResponse.json({ error: 'ユーザー情報の取得中にエラーが発生しました' }, { status: 500 });
  }
}

// ユーザー情報を更新
export async function PUT(req: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    const userData = await req.json();

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/users/me`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(userData),
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || 'ユーザー情報の更新に失敗しました' }, { status: response.status });
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('ユーザー情報更新エラー:', error);
    return NextResponse.json({ error: 'ユーザー情報の更新中にエラーが発生しました' }, { status: 500 });
  }
}