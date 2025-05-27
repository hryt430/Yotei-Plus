"use client"

import { useEffect, useState } from "react"
import type { Task } from "@/types/task"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Calendar, TrendingUp } from "lucide-react"

interface TaskAnalyticsPanelProps {
  tasks: Task[]
}

export function TaskAnalyticsPanel({ tasks }: TaskAnalyticsPanelProps) {
  const [animatedProgress, setAnimatedProgress] = useState(0)
  const [animatedWeeklyProgress, setAnimatedWeeklyProgress] = useState(0)

  // Today's tasks calculation
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const endOfToday = new Date(today)
  endOfToday.setHours(23, 59, 59, 999)

  const todayTasks = tasks.filter((task) => {
    const taskDate = new Date(task.dueDate)
    taskDate.setHours(0, 0, 0, 0)
    return taskDate.getTime() === today.getTime()
  })

  const todayCompleted = todayTasks.filter((task) => task.status === "completed").length
  const todayTotal = todayTasks.length
  const todayProgress = todayTotal > 0 ? (todayCompleted / todayTotal) * 100 : 0

  // Upcoming week tasks (next 7 days including today)
  const upcomingWeekTasks = tasks.filter((task) => {
    const taskDate = new Date(task.dueDate)
    const nextWeek = new Date()
    nextWeek.setDate(nextWeek.getDate() + 7)
    return taskDate >= today && taskDate <= nextWeek
  })

  const weekCompleted = upcomingWeekTasks.filter((task) => task.status === "completed").length
  const weekOverdue = upcomingWeekTasks.filter((task) => {
    const taskDate = new Date(task.dueDate)
    return taskDate < today && task.status !== "completed"
  }).length
  const weekTotal = upcomingWeekTasks.length
  const weekProgress = weekTotal > 0 ? ((weekCompleted + weekOverdue) / weekTotal) * 100 : 0

  // Weekly progress by day (next 7 days)
  const weekDays = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
  const weeklyData = weekDays.map((day, index) => {
    const date = new Date()
    date.setDate(date.getDate() + index)
    date.setHours(0, 0, 0, 0)

    const dayTasks = tasks.filter((task) => {
      const taskDate = new Date(task.dueDate)
      taskDate.setHours(0, 0, 0, 0)
      return taskDate.getTime() === date.getTime()
    })

    const completed = dayTasks.filter((task) => task.status === "completed").length
    const total = dayTasks.length
    const progressPercent = total > 0 ? (completed / total) * 100 : 0

    return {
      day,
      progress: progressPercent,
      completed,
      total,
      hasData: total > 0,
    }
  })

  // Animation effects
  useEffect(() => {
    const timer = setTimeout(() => {
      setAnimatedProgress(todayProgress)
    }, 300)
    return () => clearTimeout(timer)
  }, [todayProgress])

  useEffect(() => {
    const timer = setTimeout(() => {
      setAnimatedWeeklyProgress(weekProgress)
    }, 800)
    return () => clearTimeout(timer)
  }, [weekProgress])

  // Get color based on progress percentage
  const getProgressColor = (progress: number) => {
    if (progress === 100) return "#065f46" // dark green
    if (progress >= 80) return "#10b981" // green
    if (progress >= 60) return "#f59e0b" // yellow
    if (progress >= 40) return "#f97316" // orange
    if (progress >= 20) return "#f87171" // light red
    if (progress >= 1) return "#ef4444" // red
    return "#9ca3af" // gray
  }

  // Circular progress component
  const CircularProgress = ({ progress, size = 160 }: { progress: number; size?: number }) => {
    const radius = (size - 20) / 2
    const circumference = 2 * Math.PI * radius
    const strokeDasharray = circumference
    const strokeDashoffset = circumference - (progress / 100) * circumference

    return (
      <div className="relative flex items-center justify-center">
        <svg width={size} height={size} className="transform -rotate-90">
          {/* Background circle */}
          <circle cx={size / 2} cy={size / 2} r={radius} stroke="#e5e7eb" strokeWidth="8" fill="transparent" />
          {/* Progress circle */}
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            stroke="#10b981"
            strokeWidth="8"
            fill="transparent"
            strokeDasharray={strokeDasharray}
            strokeDashoffset={strokeDashoffset}
            strokeLinecap="round"
            className="transition-all duration-3000 ease-out"
            style={{
              filter: "drop-shadow(0 0 6px rgba(16, 185, 129, 0.3))",
            }}
          />
        </svg>
        {/* Center text */}
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <div className="text-2xl font-bold text-gray-900">{Math.round(animatedWeeklyProgress)}%</div>
          <div className="text-xs text-gray-500 mt-1">Complete</div>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col bg-gray-50/30">
      {/* Today's Progress - Compact */}
      <div className="flex-shrink-0 p-3 border-b border-gray-200">
        <Card className="h-full">
          <CardHeader className="pb-2">
            <CardTitle className="text-base font-semibold flex items-center">
              <Calendar className="w-4 h-4 mr-2 text-blue-600" />
              Today's Progress
            </CardTitle>
          </CardHeader>
          <CardContent className="pb-3">
            {/* Progress bar with stats to the right */}
            <div className="flex items-center space-x-3">
              {/* Progress bar */}
              <div className="flex-1 bg-gray-200 rounded-full h-3 overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-green-500 to-green-600 rounded-full transition-all duration-1000 ease-out"
                  style={{
                    width: `${animatedProgress}%`,
                    boxShadow: "0 0 8px rgba(16, 185, 129, 0.4)",
                  }}
                />
              </div>
              {/* Percentage and count to the right */}
              <div className="text-base font-bold text-gray-900 whitespace-nowrap">
                {Math.round(todayProgress)}% ({todayCompleted}/{todayTotal})
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Upcoming Week - Pie chart positioned lower, Weekly Overview larger and lower */}
      <div className="flex-1 p-3">
        <Card className="h-full">
          <CardHeader className="pb-2">
            <CardTitle className="text-base font-semibold flex items-center">
              <TrendingUp className="w-4 h-4 mr-2 text-purple-600" />
              Upcoming Week
            </CardTitle>
          </CardHeader>
          <CardContent className="h-full flex flex-col pb-3">
            {/* Circular progress chart positioned lower */}
            <div className="flex-1 flex items-center justify-center pt-4">
              <CircularProgress progress={animatedWeeklyProgress} size={160} />
            </div>

            {/* Weekly progress boxes - larger and positioned lower */}
            <div className="flex-shrink-0 mt-4">
              <div className="text-base font-medium text-gray-700 mb-4 text-center">Weekly Overview</div>
              <div className="space-y-3">
                {/* Day labels */}
                <div className="flex justify-between text-sm text-gray-500 font-medium">
                  {weeklyData.map((day) => (
                    <span key={day.day} className="flex-1 text-center">
                      {day.day}
                    </span>
                  ))}
                </div>
                {/* Progress boxes - larger */}
                <div className="flex justify-between space-x-2">
                  {weeklyData.map((day, index) => (
                    <div key={day.day} className="flex-1 flex justify-center">
                      <div
                        className="w-8 h-8 rounded transition-all duration-700 ease-out"
                        style={{
                          backgroundColor: day.hasData ? getProgressColor(day.progress) : "#e5e7eb",
                          transitionDelay: `${index * 100}ms`,
                        }}
                      />
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
