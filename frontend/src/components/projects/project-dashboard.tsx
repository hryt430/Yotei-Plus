"use client"

import { useState } from "react"
import { Button } from "@/components/ui/forms/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/data-display/card"
import { Badge } from "@/components/ui/data-display/badge"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/data-display/avatar"
import { Progress } from "@/components/ui/data-display/progress"
import {
  Plus,
  Folder,
  Users,
  Calendar,
  TrendingUp,
  AlertTriangle,
  CheckCircle,
  Clock,
  DollarSign,
  MoreHorizontal,
} from "lucide-react"
import type { ProjectView, ProjectStatus } from "@/types"
import { useProject } from "@/lib/hooks/useProject"

interface ProjectDashboardProps {
  onCreateProject: () => void
  onViewProject: (project: ProjectView) => void
}

export function ProjectDashboard({ onCreateProject, onViewProject }: ProjectDashboardProps) {
  const { projects } = useProject()
  const [filter, setFilter] = useState<"all" | "active" | "completed" | "planning">("all")

  // プロジェクトのステータスを計算する関数
  const getProjectStatus = (project: ProjectView): ProjectStatus => {
    const { stats } = project
    if (stats.completionRate === 100) return 'completed'
    if (stats.totalTasks === 0) return 'planning'
    if (stats.tasksInProgress > 0 || stats.completedTasks > 0) return 'active'
    return 'planning'
  }

  // プロジェクトの優先度を計算する関数（過去期限タスクの数に基づく）
  const getProjectPriority = (project: ProjectView): 'HIGH' | 'MEDIUM' | 'LOW' => {
    const { stats } = project
    if (stats.overdueTasks > 5) return 'HIGH'
    if (stats.overdueTasks > 0) return 'MEDIUM'
    return 'LOW'
  }

  const filteredProjects = projects.filter((project: ProjectView) => {
    if (filter === "all") return true
    const status = getProjectStatus(project)
    return status === filter
  })

  const getStatusBadge = (status: ProjectStatus) => {
    switch (status) {
      case "planning":
        return "bg-blue-100 text-blue-800 border-blue-200"
      case "active":
        return "bg-green-100 text-green-800 border-green-200"
      case "on-hold":
        return "bg-yellow-100 text-yellow-800 border-yellow-200"
      case "completed":
        return "bg-gray-100 text-gray-800 border-gray-200"
      default:
        return "bg-gray-100 text-gray-800 border-gray-200"
    }
  }

  const getPriorityBadge = (priority: string) => {
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

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    })
  }

  // プロジェクトの残り日数を計算
  const getDaysRemaining = (project: ProjectView): number => {
    if (project.stats.daysRemaining !== undefined) {
      return project.stats.daysRemaining
    }
    
    // 最も遅いタスクの期限から計算
    const latestDueDate = project.tasks
      .filter(task => task.due_date)
      .map(task => new Date(task.due_date!))
      .sort((a, b) => b.getTime() - a.getTime())[0]
    
    if (!latestDueDate) return 0
    
    const today = new Date()
    const diffTime = latestDueDate.getTime() - today.getTime()
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
    return diffDays
  }

  // Calculate overview stats
  const totalProjects = projects.length
  const activeProjects = projects.filter((p: ProjectView) => getProjectStatus(p) === 'active').length
  const completedProjects = projects.filter((p: ProjectView) => getProjectStatus(p) === 'completed').length
  const overdueProjects = projects.filter((p: ProjectView) => p.stats.overdueTasks > 0).length

  // プロジェクトカラーを生成（ハッシュベース）
  const getProjectColor = (projectId: string): string => {
    const colors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899']
    const hash = projectId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
    return colors[hash % colors.length]
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="p-6 border-b border-gray-200 bg-white flex-shrink-0">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Project Dashboard</h1>
            <p className="text-gray-600">Manage and track your projects</p>
          </div>
          <Button onClick={onCreateProject} className="bg-blue-600 hover:bg-blue-700 text-white">
            <Plus className="w-4 h-4 mr-2" />
            New Project
          </Button>
        </div>

        {/* Overview Stats */}
        <div className="grid grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Total Projects</p>
                  <p className="text-2xl font-bold text-gray-900">{totalProjects}</p>
                </div>
                <Folder className="w-8 h-8 text-blue-600" />
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Active</p>
                  <p className="text-2xl font-bold text-green-600">{activeProjects}</p>
                </div>
                <TrendingUp className="w-8 h-8 text-green-600" />
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Completed</p>
                  <p className="text-2xl font-bold text-gray-600">{completedProjects}</p>
                </div>
                <CheckCircle className="w-8 h-8 text-gray-600" />
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Overdue</p>
                  <p className="text-2xl font-bold text-red-600">{overdueProjects}</p>
                </div>
                <AlertTriangle className="w-8 h-8 text-red-600" />
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Filter Tabs */}
      <div className="p-6 border-b border-gray-200 bg-gray-50/30 flex-shrink-0">
        <div className="flex space-x-1 bg-white rounded-lg p-1 border border-gray-200">
          {[
            { key: "all", label: "All Projects", count: totalProjects },
            { key: "active", label: "Active", count: activeProjects },
            { key: "completed", label: "Completed", count: completedProjects },
            { key: "planning", label: "Planning", count: projects.filter(p => getProjectStatus(p) === 'planning').length },
          ].map((tab) => (
            <button
              key={tab.key}
              onClick={() => setFilter(tab.key as any)}
              className={`flex-1 text-sm py-2 px-4 rounded transition-colors ${
                filter === tab.key ? "bg-blue-600 text-white shadow-sm" : "text-gray-600 hover:text-gray-900"
              }`}
            >
              {tab.label} ({tab.count})
            </button>
          ))}
        </div>
      </div>

      {/* Projects Grid */}
      <div className="flex-1 p-6 overflow-y-auto">
        {filteredProjects.length === 0 ? (
          <div className="text-center py-12">
            <Folder className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No projects found</h3>
            <p className="text-gray-600 mb-4">Get started by creating your first project</p>
            <Button onClick={onCreateProject} className="bg-blue-600 hover:bg-blue-700 text-white">
              <Plus className="w-4 h-4 mr-2" />
              Create Project
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredProjects.map((project: ProjectView) => {
              const status = getProjectStatus(project)
              const priority = getProjectPriority(project)
              const daysRemaining = getDaysRemaining(project)
              const projectColor = getProjectColor(project.group.id)

              return (
                <Card
                  key={project.group.id}
                  className="hover:shadow-lg transition-all duration-300 hover:-translate-y-1 cursor-pointer"
                  onClick={() => onViewProject(project)}
                >
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <div className="w-3 h-3 rounded-full" style={{ backgroundColor: projectColor }} />
                          <CardTitle className="text-lg font-semibold text-gray-900 truncate">
                            {project.group.name}
                          </CardTitle>
                        </div>
                        <CardDescription className="text-sm text-gray-600 line-clamp-2">
                          {project.group.description || 'No description'}
                        </CardDescription>
                      </div>
                      <Button variant="ghost" size="sm" className="p-1">
                        <MoreHorizontal className="w-4 h-4" />
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {/* Status and Priority */}
                    <div className="flex items-center justify-between">
                      <Badge variant="outline" className={getStatusBadge(status)}>
                        {status.charAt(0).toUpperCase() + status.slice(1)}
                      </Badge>
                      <Badge variant="outline" className={getPriorityBadge(priority)}>
                        {priority} Priority
                      </Badge>
                    </div>

                    {/* Progress */}
                    <div className="space-y-2">
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-gray-600">Progress</span>
                        <span className="font-medium">{Math.round(project.stats.completionRate)}%</span>
                      </div>
                      <Progress value={project.stats.completionRate} className="h-2" />
                    </div>

                    {/* Team Members */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Users className="w-4 h-4 text-gray-500" />
                        <span className="text-sm text-gray-600">{project.members.length} members</span>
                      </div>
                      <div className="flex -space-x-2">
                        {project.members.slice(0, 3).map((member) => (
                          <Avatar key={member.id} className="w-6 h-6 border-2 border-white">
                            <AvatarImage src="/placeholder.svg" />
                            <AvatarFallback className="bg-blue-100 text-blue-600 text-xs">
                              {(member.user?.username || 'U')
                                .split(" ")
                                .map((n: string) => n[0])
                                .join("")
                                .toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                        ))}
                        {project.members.length > 3 && (
                          <div className="w-6 h-6 rounded-full bg-gray-100 border-2 border-white flex items-center justify-center">
                            <span className="text-xs text-gray-600">+{project.members.length - 3}</span>
                          </div>
                        )}
                      </div>
                    </div>

                    {/* Tasks Summary */}
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center space-x-4">
                        <div className="flex items-center space-x-1">
                          <CheckCircle className="w-4 h-4 text-green-600" />
                          <span className="text-gray-600">
                            {project.stats.completedTasks}/{project.stats.totalTasks}
                          </span>
                        </div>
                        {project.stats.overdueTasks > 0 && (
                          <div className="flex items-center space-x-1">
                            <AlertTriangle className="w-4 h-4 text-red-600" />
                            <span className="text-red-600">{project.stats.overdueTasks}</span>
                          </div>
                        )}
                      </div>
                    </div>

                    {/* Timeline */}
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center space-x-1">
                        <Calendar className="w-4 h-4 text-gray-500" />
                        <span className="text-gray-600">{formatDate(project.group.created_at)}</span>
                      </div>
                      <div
                        className={`flex items-center space-x-1 ${
                          daysRemaining < 0 
                            ? "text-red-600" 
                            : daysRemaining < 7 
                            ? "text-yellow-600" 
                            : "text-gray-600"
                        }`}
                      >
                        <Clock className="w-4 h-4" />
                        <span>
                          {daysRemaining < 0 
                            ? `${Math.abs(daysRemaining)} days overdue` 
                            : `${daysRemaining} days left`}
                        </span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

export default ProjectDashboard