// src/app/api/tasks/[id]/assign/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth/next';
import { authOptions } from '../../../auth/[...nextauth]/route';

// タスクをユーザーに割り当てる
export async function POST(
  req: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const session = await getServerSession(authOptions);
    
    if (!session || !session.accessToken) {
      return NextResponse.json({ error: '認証が必要です' }, { status: 401 });
    }

    const taskId = params.id;
    const { assigneeId } = await req.json();

    // バックエンドAPIにリクエスト
    const response = await fetch(`${process.env.API_URL}/tasks/${taskId}/assign`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${session.accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ assigneeId }),
    });

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json({ error: error.message || 'タスクの割り当てに失敗しました' }, { status: response.status });
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('タスク割り当てエラー:', error);
    return NextResponse.json({ error: 'タスクの割り当て中にエラーが発生しました' }, { status: 500 });
  }
}