"use client"

import type React from "react"

import type { Task, TaskStatus } from "@/types"
import { Badge } from "@/components/ui/data-display/badge"
import { Checkbox } from "@/components/ui/forms/checkbox"
import { Calendar, Flag } from "lucide-react"

interface TaskItemProps {
  task: Task
  onUpdateTaskDate: (taskId: string, newDate: Date) => void
  onUpdateTaskStatus: (taskId: string, status: TaskStatus) => void
}

export function TaskItem({ task, onUpdateTaskDate, onUpdateTaskStatus }: TaskItemProps) {
  const handleDragStart = (e: React.DragEvent) => {
    e.dataTransfer.setData(
      "text/plain",
      JSON.stringify({
        taskId: task.id,
        type: "task",
      }),
    )
  }

  const isToday = (date: Date) => {
    const today = new Date()
    return date.toDateString() === today.toDateString()
  }

  const isOverdue = (date: Date) => {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    return date < today && task.status !== "DONE"
  }

  const formatDate = (date: Date) => {
    if (isToday(date)) return "Today"

    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    if (date.toDateString() === yesterday.toDateString()) return "Yesterday"

    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    if (date.toDateString() === tomorrow.toDateString()) return "Tomorrow"

    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: date.getFullYear() !== new Date().getFullYear() ? "numeric" : undefined,
    })
  }

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case "HIGH":
        return "bg-red-100 text-red-800 border-red-200"
      case "MEDIUM":
        return "bg-yellow-100 text-yellow-800 border-yellow-200"
      case "LOW":
        return "bg-green-100 text-green-800 border-green-200"
      default:
        return "bg-gray-100 text-gray-800 border-gray-200"
    }
  }

  const getCategoryLabel = (category: string) => {
    switch (category) {
      case "WORK":
        return "Work"
      case "PERSONAL":
        return "Personal"
      case "STUDY":
        return "Study"
      case "HEALTH":
        return "Health"
      case "SHOPPING":
        return "Shopping"
      case "OTHER":
        return "Other"
      default:
        return category
    }
  }

  // due_date が存在する場合のみ Date オブジェクトを作成
  const dueDate = task.due_date ? new Date(task.due_date) : null

  return (
    <div
      draggable
      onDragStart={handleDragStart}
      className={`group p-3 bg-gradient-to-r from-white to-gray-50/50 border border-gray-200/60 rounded-lg hover:shadow-md hover:shadow-gray-200/50 transition-all duration-300 cursor-move backdrop-blur-sm ${
        task.status === "DONE" ? "opacity-60" : ""
      } ${dueDate && isOverdue(dueDate) ? "border-red-200 bg-gradient-to-r from-red-50/30 to-white shadow-red-100/30" : ""} hover:border-gray-300/80 hover:-translate-y-0.5`}
      style={{
        boxShadow: "0 1px 3px -1px rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.06)",
      }}
    >
      <div className="flex items-start space-x-3">
        <Checkbox
          checked={task.status === "DONE"}
          onCheckedChange={(checked) => onUpdateTaskStatus(task.id, checked ? "DONE" : "TODO")}
          className="mt-0.5 shadow-sm"
        />

        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between mb-1">
            <h3 className={`font-medium text-gray-900 text-sm ${task.status === "DONE" ? "line-through" : ""}`}>
              {task.title}
            </h3>
            <Badge variant="outline" className={`text-xs shadow-sm ${getPriorityColor(task.priority)}`}>
              <Flag className="w-2 h-2 mr-1" />
              {task.priority}
            </Badge>
          </div>

          <p className="text-xs text-gray-600 mb-2 line-clamp-2">{task.description}</p>

          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3 text-xs text-gray-500">
              {dueDate && (
                <div className="flex items-center bg-gray-50/80 px-2 py-0.5 rounded">
                  <Calendar className="w-2 h-2 mr-1" />
                  <span className={isOverdue(dueDate) ? "text-red-600 font-medium" : ""}>
                    {formatDate(dueDate)}
                  </span>
                </div>
              )}
              <Badge variant="secondary" className="text-xs shadow-sm bg-gradient-to-r from-gray-100 to-gray-50">
                {getCategoryLabel(task.category)}
              </Badge>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}