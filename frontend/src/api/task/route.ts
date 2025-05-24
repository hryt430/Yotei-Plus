// src/app/api/tasks/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth/next';
import { authOptions } from '../auth/[...nextauth]/route';

// タスク一覧を取得
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
    const sort = searchParams.get('sort') || 'dueDate';
    const order = searchParams.get('order') || 'asc';
    const status = searchParams.get('status');
    const priority = searchParams.get('priority');

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/tasks?page=${page}&limit=${limit}&sort=${sort}&order=${order}${status ? `&status=${status}` : ''}${priority ? `&priority=${priority}` : ''}`, {
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || 'タスクの取得に失敗しました' }, { status: response.status });
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('タスク取得エラー:', error);
    return NextResponse.json({ error: 'タスクの取得中にエラーが発生しました' }, { status: 500 });
  }
}

// 新しいタスクを作成
export async function POST(req: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    const taskData = await req.json();

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/tasks`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(taskData),
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || 'タスクの作成に失敗しました' }, { status: response.status });
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('タスク作成エラー:', error);
    return NextResponse.json({ error: 'タスクの作成中にエラーが発生しました' }, { status: 500 });
  }
}