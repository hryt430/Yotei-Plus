"use client"

import type React from "react"

import { useState } from "react"
import type { Task } from "@/types"
import { Button } from "@/components/ui/forms/button"
import { Badge } from "@/components/ui/data-display/badge"
import { ChevronLeft, ChevronRight } from "lucide-react"

interface CalendarProps {
  tasks: Task[]
  onUpdateTaskDate: (taskId: string, newDate: Date) => void
}

export function Calendar({ tasks, onUpdateTaskDate }: CalendarProps) {
  const [currentDate, setCurrentDate] = useState(new Date())

  const getDaysInMonth = (date: Date) => {
    return new Date(date.getFullYear(), date.getMonth() + 1, 0).getDate()
  }

  const getFirstDayOfMonth = (date: Date) => {
    return new Date(date.getFullYear(), date.getMonth(), 1).getDay()
  }

  // Go バックエンドから受信したタスクの日付フィルタリング
  const getTasksForDate = (date: Date) => {
    const targetYear = date.getFullYear();
    const targetMonth = date.getMonth();
    const targetDay = date.getDate();
    
    return tasks.filter((task) => {
      if (!task.due_date) return false;
      
      try {
        // ISO 8601文字列をDateオブジェクトに変換
        const taskDate = new Date(task.due_date);
        
        // 無効な日付をチェック
        if (isNaN(taskDate.getTime())) return false;
        
        // 日付部分のみを比較（時間は無視）
        return taskDate.getFullYear() === targetYear &&
              taskDate.getMonth() === targetMonth &&
              taskDate.getDate() === targetDay;
      } catch (error) {
        console.warn('Invalid due_date format:', task.due_date);
        return false;
      }
    });
  }

  const handleDrop = (e: React.DragEvent, date: Date) => {
    e.preventDefault()
    const data = JSON.parse(e.dataTransfer.getData("text/plain"))
    if (data.type === "task") {
      onUpdateTaskDate(data.taskId, date)
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
  }

  const navigateMonth = (direction: "prev" | "next") => {
    setCurrentDate((prev) => {
      const newDate = new Date(prev)
      if (direction === "prev") {
        newDate.setMonth(prev.getMonth() - 1)
      } else {
        newDate.setMonth(prev.getMonth() + 1)
      }
      return newDate
    })
  }

  const daysInMonth = getDaysInMonth(currentDate)
  const firstDay = getFirstDayOfMonth(currentDate)
  const monthYear = currentDate.toLocaleDateString("en-US", {
    month: "long",
    year: "numeric",
  })

  const days = []
  const today = new Date()

  // Empty cells for days before the first day of the month
  for (let i = 0; i < firstDay; i++) {
    days.push(<div key={`empty-${i}`} className="h-28"></div>)
  }

  // Days of the month
  for (let day = 1; day <= daysInMonth; day++) {
    const date = new Date(currentDate.getFullYear(), currentDate.getMonth(), day)
    const dayTasks = getTasksForDate(date)
    const isToday = date.toDateString() === today.toDateString()

    days.push(
      <div
        key={day}
        onDrop={(e) => handleDrop(e, date)}
        onDragOver={handleDragOver}
        className={`h-28 p-1.5 transition-all duration-200 cursor-pointer ${
          isToday
            ? "bg-blue-500 text-white rounded-lg shadow-lg"
            : "bg-slate-50 hover:bg-slate-100 hover:shadow-md hover:-translate-y-0.5 rounded-lg border border-gray-200/60"
        }`}
      >
        <div className="flex items-center justify-between mb-1">
          <span className={`text-xs ${isToday ? "font-bold text-white" : "font-medium text-gray-700"}`}>{day}</span>
          {isToday && (
            <Badge variant="default" className="text-xs bg-white text-blue-600 shadow-sm px-1 py-0">
              Today
            </Badge>
          )}
        </div>

        <div className="space-y-0.5 overflow-y-auto max-h-20">
          {dayTasks.slice(0, 3).map((task) => (
            <div
              key={task.id}
              className={`text-xs p-1 rounded border-l-2 shadow-sm transition-all duration-200 hover:shadow-md ${
                isToday ? "bg-white/90 text-gray-900" : "bg-white"
              } ${
                task.priority === "HIGH"
                  ? "border-l-red-400"
                  : task.priority === "MEDIUM"
                    ? "border-l-yellow-400"
                    : "border-l-green-400"
              } ${task.status === "DONE" ? "opacity-60 line-through" : ""}`}
            >
              <div className="font-medium truncate text-xs">{task.title}</div>
              <div className={`truncate text-xs ${isToday ? "text-gray-600" : "text-gray-500"}`}>{task.category}</div>
            </div>
          ))}
          {dayTasks.length > 3 && (
            <div className={`text-xs text-center py-0.5 ${isToday ? "text-white/80" : "text-gray-500"}`}>
              +{dayTasks.length - 3} more
            </div>
          )}
        </div>
      </div>,
    )
  }

  return (
    <div className="h-full flex flex-col shadow-sm bg-white">
      {/* Calendar Header */}
      <div className="p-4 border-b border-gray-200 bg-white shadow-sm flex-shrink-0">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">{monthYear}</h2>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => navigateMonth("prev")}
              className="hover:shadow-md transition-all duration-200"
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setCurrentDate(new Date())}
              className="hover:shadow-md transition-all duration-200"
            >
              Today
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => navigateMonth("next")}
              className="hover:shadow-md transition-all duration-200"
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Calendar Grid */}
      <div className="flex-1 p-3 bg-white overflow-hidden">
        {/* Day Headers */}
        <div className="grid grid-cols-7 gap-1 mb-2">
          {["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((day) => (
            <div key={day} className="p-1 text-center text-xs font-semibold text-gray-500 uppercase tracking-wide">
              {day}
            </div>
          ))}
        </div>

        {/* Calendar Days */}
        <div className="grid grid-cols-7 gap-1 h-full">{days}</div>
      </div>
    </div>
  )
}
