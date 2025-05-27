"use client"

import { useState } from "react"
import type { Task } from "@/types"
import { TaskListWithFilters } from "@/components/tasks/task-list-with-filters"
import { TaskStatsChart } from "@/components/tasks/task-stats-chart"

// Sample tasks data
const sampleTasks: Task[] = [
  {
    id: "1",
    title: "Review project proposal",
    description: "Go through the Q4 project proposal and provide feedback",
    dueDate: new Date(),
    priority: "HIGH",
    status: "TODO",
    category: "Work",
  },
  {
    id: "2",
    title: "Team standup meeting",
    description: "Daily standup with the development team",
    dueDate: new Date(),
    priority: "MEDIUM",
    status: "DONE",
    category: "Meetings",
  },
  {
    id: "3",
    title: "Update documentation",
    description: "Update the API documentation for the new endpoints",
    dueDate: new Date(Date.now() - 86400000),
    priority: "LOW",
    status: "DONE",
    category: "Documentation",
  },
  {
    id: "4",
    title: "Client presentation",
    description: "Present the new features to the client",
    dueDate: new Date(Date.now() + 86400000),
    priority: "HIGH",
    status: "IN_PROGRESS",
    category: "Meetings",
  },
  {
    id: "5",
    title: "Code review",
    description: "Review pull requests from team members",
    dueDate: new Date(Date.now() - 172800000),
    priority: "MEDIUM",
    status: "DONE",
    category: "Development",
  },
  {
    id: "6",
    title: "Database optimization",
    description: "Optimize database queries for better performance",
    dueDate: new Date(Date.now() - 432000000),
    priority: "high",
    status: "pending",
    category: "Development",
  },
  {
    id: "7",
    title: "UI/UX improvements",
    description: "Implement new design system components",
    dueDate: new Date(Date.now() + 172800000),
    priority: "MEDIUM",
    status: "IN_PROGRESS",
    category: "Design",
  },
  {
    id: "8",
    title: "Security audit",
    description: "Conduct security review of the application",
    dueDate: new Date(Date.now() - 259200000),
    priority: "HIGH",
    status: "DONE",
    category: "Security",
  },
]

export default function AllTasksPage() {
  const [tasks] = useState<Task[]>(sampleTasks)

  const updateTaskDate = (taskId: string, newDate: Date) => {
    // This would update the task date
    console.log("Update task date:", taskId, newDate)
  }

  const updateTaskStatus = (taskId: string, status: "TODO" | "DONE" | "IN_PROGRESS") => {
    // This would update the task status
    console.log("Update task status:", taskId, status)
  }

  return (
    <div className="h-[calc(100vh-89px)] flex">
      {/* Left Panel - Task List with Filters (3/5) */}
      <div className="w-3/5 border-r border-gray-200">
        <TaskListWithFilters tasks={tasks} onUpdateTaskDate={updateTaskDate} onUpdateTaskStatus={updateTaskStatus} />
      </div>

      {/* Right Panel - Statistics (2/5) */}
      <div className="w-2/5 bg-gray-50/30">
        <TaskStatsChart tasks={tasks} />
      </div>
    </div>
  )
}
