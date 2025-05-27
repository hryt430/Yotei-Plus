"use client"

import type { Task } from "@/types/task"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { type ChartConfig, ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { PieChart, Pie, Cell, ResponsiveContainer, BarChart, Bar, XAxis, YAxis } from "recharts"
import { CheckCircle, Clock, TrendingUp } from "lucide-react"

interface TaskStatsChartProps {
  tasks: Task[]
}

export function TaskStatsChart({ tasks }: TaskStatsChartProps) {
  // Calculate statistics
  const totalTasks = tasks.length
  const completedTasks = tasks.filter((task) => task.status === "completed").length
  const pendingTasks = tasks.filter((task) => task.status === "pending").length
  const inProgressTasks = tasks.filter((task) => task.status === "in-progress").length

  const completionRate = totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0

  // Pie chart data
  const pieData = [
    {
      name: "Completed",
      value: completedTasks,
      color: "#10b981",
    },
    {
      name: "In Progress",
      value: inProgressTasks,
      color: "#f59e0b",
    },
    {
      name: "Pending",
      value: pendingTasks,
      color: "#ef4444",
    },
  ]

  // Priority distribution
  const priorityData = [
    {
      priority: "High",
      count: tasks.filter((task) => task.priority === "high").length,
    },
    {
      priority: "Medium",
      count: tasks.filter((task) => task.priority === "medium").length,
    },
    {
      priority: "Low",
      count: tasks.filter((task) => task.priority === "low").length,
    },
  ]

  // Category distribution
  const categoryData = tasks.reduce(
    (acc, task) => {
      acc[task.category] = (acc[task.category] || 0) + 1
      return acc
    },
    {} as Record<string, number>,
  )

  const categoryChartData = Object.entries(categoryData).map(([category, count]) => ({
    category,
    count,
  }))

  const chartConfig = {
    completed: {
      label: "Completed",
      color: "#10b981",
    },
    inProgress: {
      label: "In Progress",
      color: "#f59e0b",
    },
    pending: {
      label: "Pending",
      color: "#ef4444",
    },
  } satisfies ChartConfig

  return (
    <div className="h-full overflow-y-auto p-6 space-y-6">
      {/* Overview Cards */}
      <div className="grid grid-cols-2 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center">
              <CheckCircle className="w-4 h-4 mr-2 text-green-600" />
              Completed
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{completedTasks}</div>
            <p className="text-xs text-gray-600">{completionRate}% completion rate</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center">
              <Clock className="w-4 h-4 mr-2 text-blue-600" />
              Total Tasks
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalTasks}</div>
            <p className="text-xs text-gray-600">{pendingTasks + inProgressTasks} remaining</p>
          </CardContent>
        </Card>
      </div>

      {/* Completion Status Pie Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Task Status Distribution</CardTitle>
          <CardDescription>Overview of task completion status</CardDescription>
        </CardHeader>
        <CardContent>
          <ChartContainer config={chartConfig} className="h-[200px]">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={40}
                  outerRadius={80}
                  paddingAngle={2}
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <ChartTooltip content={<ChartTooltipContent />} />
              </PieChart>
            </ResponsiveContainer>
          </ChartContainer>
          <div className="flex justify-center space-x-4 mt-4">
            {pieData.map((entry) => (
              <div key={entry.name} className="flex items-center text-sm">
                <div className="w-3 h-3 rounded-full mr-2" style={{ backgroundColor: entry.color }} />
                <span className="text-gray-600">
                  {entry.name}: {entry.value}
                </span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Priority Distribution */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Priority Distribution</CardTitle>
          <CardDescription>Tasks by priority level</CardDescription>
        </CardHeader>
        <CardContent>
          <ChartContainer config={chartConfig} className="h-[150px]">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={priorityData}>
                <XAxis dataKey="priority" />
                <YAxis />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="count" fill="#6b7280" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </ChartContainer>
        </CardContent>
      </Card>

      {/* Category Breakdown */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Categories</CardTitle>
          <CardDescription>Tasks by category</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {categoryChartData.map((item) => (
              <div key={item.category} className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-700">{item.category}</span>
                <div className="flex items-center space-x-2">
                  <div className="w-20 bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full"
                      style={{ width: `${(item.count / totalTasks) * 100}%` }}
                    />
                  </div>
                  <span className="text-sm text-gray-600 w-8">{item.count}</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Quick Stats */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center">
            <TrendingUp className="w-4 h-4 mr-2" />
            Quick Stats
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600">Overdue tasks:</span>
              <span className="font-medium">
                {tasks.filter((task) => task.dueDate < new Date() && task.status !== "completed").length}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600">Due today:</span>
              <span className="font-medium">
                {tasks.filter((task) => task.dueDate.toDateString() === new Date().toDateString()).length}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600">High priority:</span>
              <span className="font-medium">{tasks.filter((task) => task.priority === "high").length}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
