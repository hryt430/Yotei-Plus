"use client"

import type React from "react"

import { useState, useRef } from "react"
import { Button } from "@/components/ui/forms/button"
import { Badge } from "@/components/ui/data-display/badge"
import { ScrollArea } from "@/components/ui/layout/scroll-area"
import { Calendar, ZoomIn, ZoomOut, MoreHorizontal } from "lucide-react"
import type { ProjectTask } from "@/types"

interface GanttChartProps {
  tasks: ProjectTask[]
  onTaskUpdate?: (taskId: string, updates: Partial<ProjectTask>) => void
  onTaskSelect?: (task: ProjectTask) => void
}

// interface GanttTask extends ProjectTask {
//   level: number
//   hasChildren: boolean
//   isExpanded: boolean
// }

export function GanttChart({ tasks, onTaskUpdate, onTaskSelect }: GanttChartProps) {
  const [viewMode, setViewMode] = useState<"days" | "weeks" | "months">("weeks")
  // const [currentDate, setCurrentDate] = useState(new Date())
  const [selectedTask, setSelectedTask] = useState<string | null>(null)
  const [isDragging, setIsDragging] = useState(false)
  const [dragTask, setDragTask] = useState<string | null>(null)
  const chartRef = useRef<HTMLDivElement>(null)

  // Calculate date range based on tasks
  const getDateRange = () => {
    if (tasks.length === 0) {
      const today = new Date()
      const start = new Date(today)
      start.setDate(start.getDate() - 30)
      const end = new Date(today)
      end.setDate(end.getDate() + 60)
      return { start, end }
    }

    const dates = tasks.flatMap((task) => [new Date(task.start_date || ''), new Date(task.end_date || '')])
    const minDate = new Date(Math.min(...dates.map((d) => d.getTime())))
    const maxDate = new Date(Math.max(...dates.map((d) => d.getTime())))

    // Add padding
    minDate.setDate(minDate.getDate() - 7)
    maxDate.setDate(maxDate.getDate() + 7)

    return { start: minDate, end: maxDate }
  }

  const { start: startDate, end: endDate } = getDateRange()

  // Generate time periods based on view mode
  const generateTimePeriods = () => {
    const periods = []
    const current = new Date(startDate)

    while (current <= endDate) {
      periods.push(new Date(current))

      if (viewMode === "days") {
        current.setDate(current.getDate() + 1)
      } else if (viewMode === "weeks") {
        current.setDate(current.getDate() + 7)
      } else {
        current.setMonth(current.getMonth() + 1)
      }
    }

    return periods
  }

  const timePeriods = generateTimePeriods()

  // Calculate task position and width
  const getTaskPosition = (task: ProjectTask) => {
    const totalDuration = endDate.getTime() - startDate.getTime()
    const taskStartDate = new Date(task.start_date || '')
    const taskEndDate = new Date(task.end_date || '')
    const taskStart = taskStartDate.getTime() - startDate.getTime()
    const taskDuration = taskEndDate.getTime() - taskStartDate.getTime()

    const left = (taskStart / totalDuration) * 100
    const width = (taskDuration / totalDuration) * 100

    return { left: Math.max(0, left), width: Math.max(1, width) }
  }

  // Get task color based on status and priority
  const getTaskColor = (task: ProjectTask) => {
    if (task.status === "DONE") return "bg-green-500"
    if (task.status === "IN_PROGRESS") return "bg-blue-500"

    switch (task.priority) {
      case "HIGH":
        return "bg-red-400"
      case "MEDIUM":
        return "bg-yellow-400"
      case "LOW":
        return "bg-gray-400"
      default:
        return "bg-gray-400"
    }
  }

  // Format date for display
  const formatPeriodLabel = (date: Date) => {
    if (viewMode === "days") {
      return date.toLocaleDateString("en-US", { month: "short", day: "numeric" })
    } else if (viewMode === "weeks") {
      const weekEnd = new Date(date)
      weekEnd.setDate(weekEnd.getDate() + 6)
      return `${date.getDate()}/${date.getMonth() + 1} - ${weekEnd.getDate()}/${weekEnd.getMonth() + 1}`
    } else {
      return date.toLocaleDateString("en-US", { month: "short", year: "numeric" })
    }
  }

  // Check if task has dependencies
  const hasDependencies = (task: ProjectTask) => {
    return task.dependencies && task.dependencies.length > 0
  }

  // Get dependency lines
  const getDependencyLines = (task: ProjectTask, taskIndex: number) => {
    if (!task.dependencies || task.dependencies.length === 0) return []

    return task.dependencies
      .map((depId) => {
        const depTask = tasks.find((t) => t.id === depId)
        const depIndex = tasks.findIndex((t) => t.id === depId)

        if (!depTask || depIndex === -1) return null

        const depPosition = getTaskPosition(depTask)
        const taskPosition = getTaskPosition(task)

        return {
          from: { x: depPosition.left + depPosition.width, y: depIndex },
          to: { x: taskPosition.left, y: taskIndex },
        }
      })
      .filter(Boolean)
  }

  // Handle task drag
  const handleTaskDrag = (taskId: string, event: React.MouseEvent) => {
    if (!onTaskUpdate) return

    setIsDragging(true)
    setDragTask(taskId)

    const startX = event.clientX
    const task = tasks.find((t) => t.id === taskId)
    if (!task) return

    const handleMouseMove = (e: MouseEvent) => {
      const deltaX = e.clientX - startX
      const chartWidth = chartRef.current?.offsetWidth || 1000
      const totalDuration = endDate.getTime() - startDate.getTime()
      const deltaTime = (deltaX / chartWidth) * totalDuration

      const taskStartDate = new Date(task.start_date || '')
      const taskEndDate = new Date(task.end_date || '')
      const newStartDate = new Date(taskStartDate.getTime() + deltaTime)
      const newEndDate = new Date(taskEndDate.getTime() + deltaTime)

      onTaskUpdate(taskId, {
        start_date: newStartDate.toISOString().split('T')[0],
        end_date: newEndDate.toISOString().split('T')[0],
      })
    }

    const handleMouseUp = () => {
      setIsDragging(false)
      setDragTask(null)
      document.removeEventListener("mousemove", handleMouseMove)
      document.removeEventListener("mouseup", handleMouseUp)
    }

    document.addEventListener("mousemove", handleMouseMove)
    document.addEventListener("mouseup", handleMouseUp)
  }

  // Today indicator position
  const getTodayPosition = () => {
    const today = new Date()
    const totalDuration = endDate.getTime() - startDate.getTime()
    const todayOffset = today.getTime() - startDate.getTime()
    return (todayOffset / totalDuration) * 100
  }

  const todayPosition = getTodayPosition()

  return (
    <div className="h-full flex flex-col bg-white">
      {/* Header Controls */}
      <div className="p-4 border-b border-gray-200 bg-gray-50/30">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <h3 className="text-lg font-semibold text-gray-900 flex items-center">
              <Calendar className="w-5 h-5 mr-2" />
              Gantt Chart
            </h3>
            <div className="flex items-center space-x-2">
              <Button
                variant={viewMode === "days" ? "default" : "outline"}
                size="sm"
                onClick={() => setViewMode("days")}
              >
                Days
              </Button>
              <Button
                variant={viewMode === "weeks" ? "default" : "outline"}
                size="sm"
                onClick={() => setViewMode("weeks")}
              >
                Weeks
              </Button>
              <Button
                variant={viewMode === "months" ? "default" : "outline"}
                size="sm"
                onClick={() => setViewMode("months")}
              >
                Months
              </Button>
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <Button variant="outline" size="sm">
              <ZoomOut className="w-4 h-4" />
            </Button>
            <Button variant="outline" size="sm">
              <ZoomIn className="w-4 h-4" />
            </Button>
            <Button variant="outline" size="sm">
              Today
            </Button>
          </div>
        </div>
      </div>

      {/* Chart Container */}
      <div className="flex-1 flex overflow-hidden">
        {/* Task List Panel */}
        <div className="w-80 border-r border-gray-200 bg-gray-50/30">
          <div className="p-3 border-b border-gray-200 bg-white">
            <h4 className="font-medium text-gray-900">Tasks</h4>
          </div>
          <ScrollArea className="h-full">
            <div className="p-2 space-y-1">
              {tasks.map((task) => (
                <div
                  key={task.id}
                  className={`p-3 rounded-lg border cursor-pointer transition-all ${
                    selectedTask === task.id
                      ? "bg-blue-50 border-blue-200"
                      : "bg-white border-gray-200 hover:bg-gray-50"
                  }`}
                  onClick={() => {
                    setSelectedTask(task.id)
                    onTaskSelect?.(task)
                  }}
                >
                  <div className="flex items-center justify-between mb-2">
                    <h5 className="font-medium text-sm text-gray-900 truncate">{task.title}</h5>
                    <Button variant="ghost" size="sm" className="h-6 w-6 p-0">
                      <MoreHorizontal className="w-3 h-3" />
                    </Button>
                  </div>

                  <div className="flex items-center space-x-2 mb-2">
                    <Badge
                      variant="outline"
                      className={`text-xs ${
                        task.priority === "HIGH"
                          ? "bg-red-100 text-red-800 border-red-200"
                          : task.priority === "MEDIUM"
                            ? "bg-yellow-100 text-yellow-800 border-yellow-200"
                            : "bg-green-100 text-green-800 border-green-200"
                      }`}
                    >
                      {task.priority}
                    </Badge>
                    <Badge
                      variant="outline"
                      className={`text-xs ${
                        task.status === "DONE"
                          ? "bg-green-100 text-green-800 border-green-200"
                          : task.status === "IN_PROGRESS"
                            ? "bg-blue-100 text-blue-800 border-blue-200"
                            : "bg-gray-100 text-gray-800 border-gray-200"
                      }`}
                    >
                      {task.status}
                    </Badge>
                  </div>

                  <div className="text-xs text-gray-500 space-y-1">
                    <div>Start: {task.start_date ? new Date(task.start_date).toLocaleDateString() : 'N/A'}</div>
                    <div>End: {task.end_date ? new Date(task.end_date).toLocaleDateString() : 'N/A'}</div>
                    {task.assignee_id && <div>ID: {task.assignee_id}</div>}
                  </div>

                  {/* Progress Bar */}
                  <div className="mt-2">
                    <div className="flex items-center justify-between text-xs text-gray-600 mb-1">
                      <span>Progress</span>
                      <span>{task.progress}%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-1.5">
                      <div
                        className="bg-blue-500 h-1.5 rounded-full transition-all"
                        style={{ width: `${task.progress}%` }}
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>
        </div>

        {/* Chart Panel */}
        <div className="flex-1 overflow-auto">
          <div ref={chartRef} className="relative min-w-full">
            {/* Timeline Header */}
            <div className="sticky top-0 z-10 bg-white border-b border-gray-200">
              <div className="h-12 flex">
                {timePeriods.map((period, index) => (
                  <div
                    key={index}
                    className="flex-1 min-w-24 px-2 py-3 border-r border-gray-200 text-xs font-medium text-gray-600 text-center"
                  >
                    {formatPeriodLabel(period)}
                  </div>
                ))}
              </div>
            </div>

            {/* Chart Grid and Tasks */}
            <div className="relative">
              {/* Grid Lines */}
              <div className="absolute inset-0 flex">
                {timePeriods.map((_, index) => (
                  <div key={index} className="flex-1 min-w-24 border-r border-gray-100" />
                ))}
              </div>

              {/* Today Indicator */}
              {todayPosition >= 0 && todayPosition <= 100 && (
                <div className="absolute top-0 bottom-0 w-0.5 bg-red-500 z-20" style={{ left: `${todayPosition}%` }}>
                  <div className="absolute -top-2 -left-2 w-4 h-4 bg-red-500 rounded-full" />
                </div>
              )}

              {/* Task Bars */}
              <div className="relative">
                {tasks.map((task, index) => {
                  const position = getTaskPosition(task)
                  const taskColor = getTaskColor(task)

                  return (
                    <div key={task.id} className="relative h-16 border-b border-gray-100 flex items-center">
                      {/* Task Bar */}
                      <div
                        className={`absolute h-8 rounded-md cursor-pointer transition-all hover:shadow-md ${taskColor} ${
                          selectedTask === task.id ? "ring-2 ring-blue-400" : ""
                        } ${isDragging && dragTask === task.id ? "opacity-70" : ""}`}
                        style={{
                          left: `${position.left}%`,
                          width: `${position.width}%`,
                        }}
                        onMouseDown={(e) => handleTaskDrag(task.id, e)}
                        onClick={() => {
                          setSelectedTask(task.id)
                          onTaskSelect?.(task)
                        }}
                      >
                        {/* Task Progress */}
                        <div
                          className="h-full bg-black bg-opacity-20 rounded-md transition-all"
                          style={{ width: `${task.progress}%` }}
                        />

                        {/* Task Label */}
                        <div className="absolute inset-0 flex items-center px-2">
                          <span className="text-white text-xs font-medium truncate">{task.title}</span>
                        </div>

                        {/* Resize Handles */}
                        <div className="absolute left-0 top-0 bottom-0 w-1 cursor-w-resize hover:bg-white hover:bg-opacity-30" />
                        <div className="absolute right-0 top-0 bottom-0 w-1 cursor-e-resize hover:bg-white hover:bg-opacity-30" />
                      </div>

                      {/* Dependencies */}
                      {hasDependencies(task) && (
                        <div className="absolute inset-0 pointer-events-none">
                          {getDependencyLines(task, index).map(
                            (line, lineIndex) =>
                              line && (
                                <svg key={lineIndex} className="absolute inset-0 w-full h-full" style={{ zIndex: 5 }}>
                                  <path
                                    d={`M ${line.from.x}% ${(line.from.y + 0.5) * 64} L ${line.to.x}% ${(line.to.y + 0.5) * 64}`}
                                    stroke="#6b7280"
                                    strokeWidth="2"
                                    fill="none"
                                    markerEnd="url(#arrowhead)"
                                  />
                                  <defs>
                                    <marker
                                      id="arrowhead"
                                      markerWidth="10"
                                      markerHeight="7"
                                      refX="9"
                                      refY="3.5"
                                      orient="auto"
                                    >
                                      <polygon points="0 0, 10 3.5, 0 7" fill="#6b7280" />
                                    </marker>
                                  </defs>
                                </svg>
                              ),
                          )}
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="p-3 border-t border-gray-200 bg-gray-50/30">
        <div className="flex items-center space-x-6 text-xs">
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-green-500 rounded" />
            <span>Completed</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-blue-500 rounded" />
            <span>In Progress</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-red-400 rounded" />
            <span>High Priority</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-yellow-400 rounded" />
            <span>Medium Priority</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-gray-400 rounded" />
            <span>Low Priority</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-0.5 h-3 bg-red-500" />
            <span>Today</span>
          </div>
        </div>
      </div>
    </div>
  )
}

export default GanttChart
