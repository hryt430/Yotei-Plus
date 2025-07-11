"use client"

import { useState } from "react"
import type { Task, TaskStatus, TaskPriority } from "@/types"
import { TaskItem } from "@/components/tasks/task-item"
import { Input } from "@/components/ui/forms/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/forms/select"
import { Button } from "@/components/ui/forms/button"
import { Search, Filter, SortAsc, RotateCcw } from "lucide-react"

interface TaskListWithFiltersProps {
  tasks: Task[]
  onUpdateTaskDate: (taskId: string, newDate: Date) => void
  onUpdateTaskStatus: (taskId: string, status: TaskStatus) => void
}

export function TaskListWithFilters({ tasks, onUpdateTaskDate, onUpdateTaskStatus }: TaskListWithFiltersProps) {
  const [searchTerm, setSearchTerm] = useState("")
  const [sortBy, setSortBy] = useState<"date" | "priority" | "status" | "category">("date")
  const [filterStatus, setFilterStatus] = useState<"all" | TaskStatus>("all")
  const [filterPriority, setFilterPriority] = useState<"all" | TaskPriority>("all")

  const filteredTasks = tasks
    .filter((task) => {
      const matchesSearch =
        task.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
        task.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        task.category.toLowerCase().includes(searchTerm.toLowerCase())

      const matchesStatus = filterStatus === "all" || task.status === filterStatus
      const matchesPriority = filterPriority === "all" || task.priority === filterPriority

      return matchesSearch && matchesStatus && matchesPriority
    })
    .sort((a, b) => {
      switch (sortBy) {
        case "priority":
          const priorityOrder = { HIGH: 3, MEDIUM: 2, LOW: 1 }
          return priorityOrder[b.priority] - priorityOrder[a.priority]
        case "status":
          return a.status.localeCompare(b.status)
        case "category":
          return a.category.localeCompare(b.category)
        case "date":
        default:
          // Handle due_date comparison with null checks
          const aDate = a.due_date ? new Date(a.due_date).getTime() : 0
          const bDate = b.due_date ? new Date(b.due_date).getTime() : 0
          return bDate - aDate
      }
    })

  const clearFilters = () => {
    setSearchTerm("")
    setFilterStatus("all")
    setFilterPriority("all")
    setSortBy("date")
  }

  return (
    <div className="h-full flex flex-col">
      {/* Compact Header with Search and Filters */}
      <div className="p-4 border-b border-gray-200 bg-white">
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">All Tasks ({filteredTasks.length})</h2>
            <Button variant="outline" size="sm" onClick={clearFilters} className="text-xs px-2 py-1 h-7">
              <RotateCcw className="w-3 h-3 mr-1" />
              Clear
            </Button>
          </div>

          {/* Compact Search */}
          <div className="relative">
            <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 text-gray-400 w-3 h-3" />
            <Input
              placeholder="Search tasks..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-7 h-8 text-sm border-gray-200 focus:border-gray-300 focus:ring-gray-300"
            />
          </div>

          {/* Compact Filters */}
          <div className="grid grid-cols-3 gap-2">
            <div>
              <Select value={filterStatus} onValueChange={(value: "all" | TaskStatus) => setFilterStatus(value)}>
                <SelectTrigger className="h-7 border-gray-200 text-xs">
                  <div className="flex items-center">
                    <Filter className="w-3 h-3 mr-1" />
                    <SelectValue placeholder="Status" />
                  </div>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="TODO">Todo</SelectItem>
                  <SelectItem value="IN_PROGRESS">In Progress</SelectItem>
                  <SelectItem value="DONE">Done</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Select value={filterPriority} onValueChange={(value: "all" | TaskPriority) => setFilterPriority(value)}>
                <SelectTrigger className="h-7 border-gray-200 text-xs">
                  <SelectValue placeholder="Priority" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Priority</SelectItem>
                  <SelectItem value="HIGH">High</SelectItem>
                  <SelectItem value="MEDIUM">Medium</SelectItem>
                  <SelectItem value="LOW">Low</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Select value={sortBy} onValueChange={(value: "date" | "priority" | "status" | "category") => setSortBy(value)}>
                <SelectTrigger className="h-7 border-gray-200 text-xs">
                  <div className="flex items-center">
                    <SortAsc className="w-3 h-3 mr-1" />
                    <SelectValue placeholder="Sort" />
                  </div>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="date">Date</SelectItem>
                  <SelectItem value="priority">Priority</SelectItem>
                  <SelectItem value="status">Status</SelectItem>
                  <SelectItem value="category">Category</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>
      </div>

      {/* Task List */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {filteredTasks.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-gray-400 mb-2">No tasks found</div>
            <p className="text-sm text-gray-500">Try adjusting your search or filters</p>
          </div>
        ) : (
          filteredTasks.map((task) => (
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