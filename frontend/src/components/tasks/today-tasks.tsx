"use client"

import type { Task } from "@/types/task"
import { TaskItem } from "@/components/task-item"
import { Button } from "@/components/ui/button"
import { Plus } from "lucide-react"

interface TodayTasksProps {
  tasks: Task[]
  onUpdateTaskDate: (taskId: string, newDate: Date) => void
  onUpdateTaskStatus: (taskId: string, status: "pending" | "completed" | "in-progress") => void
  onCreateTask: () => void
}

export function TodayTasks({ tasks, onUpdateTaskDate, onUpdateTaskStatus, onCreateTask }: TodayTasksProps) {
  // Filter tasks with due dates up to today (including overdue)
  const today = new Date()
  today.setHours(23, 59, 59, 999) // End of today

  const todayTasks = tasks
    .filter((task) => task.dueDate <= today)
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
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-lg font-bold text-gray-900 bg-gradient-to-r from-gray-900 to-gray-700 bg-clip-text">
            Today & Overdue
          </h2>
          <Button
            size="sm"
            className="bg-gradient-to-r from-gray-900 to-gray-800 hover:from-gray-800 hover:to-gray-700 shadow-lg hover:shadow-xl transition-all duration-300 hover:-translate-y-0.5"
            onClick={onCreateTask}
          >
            <Plus className="w-4 h-4 mr-2" />
            Add Task
          </Button>
        </div>
        <p className="text-sm text-gray-600/80">Tasks due today and earlier</p>
      </div>

      {/* Task List */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3 bg-gradient-to-b from-transparent to-gray-50/20">
        {todayTasks.length === 0 ? (
          <div className="text-center py-8">
            <div className="text-gray-400 mb-2">No tasks due today</div>
            <p className="text-sm text-gray-500">You're all caught up!</p>
          </div>
        ) : (
          todayTasks.map((task) => (
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
