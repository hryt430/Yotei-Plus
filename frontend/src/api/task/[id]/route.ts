// src/app/api/tasks/[id]/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth/next';
import { authOptions } from '../../auth/[...nextauth]/route';

// 特定のタスクを取得
export async function GET(
  req: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    const taskId = params.id;

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/tasks/${taskId}`, {
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

// タスクを更新
export async function PUT(
  req: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    const taskId = params.id;
    const taskData = await req.json();

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/tasks/${taskId}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(taskData),
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || 'タスクの更新に失敗しました' }, { status: response.status });
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('タスク更新エラー:', error);
    return NextResponse.json({ error: 'タスクの更新中にエラーが発生しました' }, { status: 500 });
  }
}

// タスクを削除
export async function DELETE(
  req: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    const taskId = params.id;

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/tasks/${taskId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || 'タスクの削除に失敗しました' }, { status: response.status });
    }

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error('タスク削除エラー:', error);
    return NextResponse.json({ error: 'タスクの削除中にエラーが発生しました' }, { status: 500 });
  }
}