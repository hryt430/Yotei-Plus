"use client"

import type { Task } from "@/types/task"
import { TaskItem } from "@/components/task-item"

interface WeekTasksProps {
  tasks: Task[]
  onUpdateTaskDate: (taskId: string, newDate: Date) => void
  onUpdateTaskStatus: (taskId: string, status: "pending" | "completed" | "in-progress") => void
}

export function WeekTasks({ tasks, onUpdateTaskDate, onUpdateTaskStatus }: WeekTasksProps) {
  // Filter tasks for the last week
  const oneWeekAgo = new Date()
  oneWeekAgo.setDate(oneWeekAgo.getDate() - 7)
  oneWeekAgo.setHours(0, 0, 0, 0)

  const weekTasks = tasks
    .filter((task) => {
      const taskDate = new Date(task.dueDate)
      return taskDate >= oneWeekAgo
    })
    .sort((a, b) => {
      // Sort by status (pending first), then by due date (most recent first)
      if (a.status !== b.status) {
        if (a.status === "pending") return -1
        if (b.status === "pending") return 1
      }
      return b.dueDate.getTime() - a.dueDate.getTime()
    })

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div
        className="p-4 border-b border-gray-200/60 bg-gradient-to-r from-white to-gray-50/30 backdrop-blur-sm"
        style={{
          boxShadow: "0 1px 3px -1px rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.06)",
        }}
      >
        <h2 className="text-lg font-bold text-gray-900 bg-gradient-to-r from-gray-900 to-gray-700 bg-clip-text mb-2">
          This Week
        </h2>
        <p className="text-sm text-gray-600/80">Tasks from the last 7 days</p>
      </div>

      {/* Task List */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3 bg-gradient-to-b from-transparent to-gray-50/20">
        {weekTasks.length === 0 ? (
          <div className="text-center py-8">
            <div className="text-gray-400 mb-2">No tasks this week</div>
            <p className="text-sm text-gray-500">Tasks will appear here</p>
          </div>
        ) : (
          weekTasks.map((task) => (
            <TaskItem
              key={task.id}
              task={task}
              onUpdateTaskDate={onUpdateTaskDate}
              onUpdateTaskStatus={onUpdateTaskStatus}
            />
          ))
        )}
      </div>
    </div>
  )
}
