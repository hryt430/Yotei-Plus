"use client"

import type { Task } from "@/types/task"
import { TodayTasks } from "@/components/today-tasks"
import { WeekTasks } from "@/components/week-tasks"

interface TaskListProps {
  tasks: Task[]
  onUpdateTaskDate: (taskId: string, newDate: Date) => void
  onUpdateTaskStatus: (taskId: string, status: "pending" | "completed" | "in-progress") => void
  onCreateTask: () => void
}

export function TaskList({ tasks, onUpdateTaskDate, onUpdateTaskStatus, onCreateTask }: TaskListProps) {
  return (
    <div className="h-full flex flex-col">
      {/* Top Half - Today's Tasks */}
      <div className="h-1/2 border-b border-gray-200">
        <TodayTasks
          tasks={tasks}
          onUpdateTaskDate={onUpdateTaskDate}
          onUpdateTaskStatus={onUpdateTaskStatus}
          onCreateTask={onCreateTask}
        />
      </div>

      {/* Bottom Half - This Week's Tasks */}
      <div className="h-1/2">
        <WeekTasks tasks={tasks} onUpdateTaskDate={onUpdateTaskDate} onUpdateTaskStatus={onUpdateTaskStatus} />
      </div>
    </div>
  )
}
