"use client"

import { useState } from "react"
import { Button } from "@/components/ui/forms/button"
import { Badge } from "@/components/ui/data-display/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/layout/tabs"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/data-display/card"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/data-display/avatar"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/navigation/dropdown-menu"
import { GanttChart } from "./gantt-chart"
import { ProjectTaskCreationModal } from "./project-task-creation-modal"
import { ProjectMemberModal } from "./project-member-modal"
import { ArrowLeft, Plus, BarChart3, ListIcon, Users, Calendar, MoreHorizontal, Edit, Trash2 } from "lucide-react"
import type { ProjectView, ProjectTask, ProjectMember } from "@/types"
import { useProjectManagement } from "./project-management-provider"

interface ProjectDetailPageProps {
  project: ProjectView
  onCreateTask: () => void
  onBack: () => void
}

export function ProjectDetailPage({ project, onCreateTask: _onCreateTask, onBack }: ProjectDetailPageProps) {
  const [activeTab, setActiveTab] = useState<"overview" | "gantt" | "team">("overview")
  const [isTaskModalOpen, setIsTaskModalOpen] = useState(false)
  const [isMemberModalOpen, setIsMemberModalOpen] = useState(false)
  // const [isTaskEditModalOpen, setIsTaskEditModalOpen] = useState(false)
  // const [selectedTask, setSelectedTask] = useState<ProjectTask | null>(null)

  const { addTaskToProject, deleteTaskFromProject, addMemberToProject, removeMemberFromProject } =
    useProjectManagement()

  // const handleProjectTaskUpdate = (taskId: string, updates: Partial<ProjectTask>) => {
  //   updateTask(project.id, taskId, updates)
  // }

  const handleTaskSelect = (task: ProjectTask) => {
    console.log("Selected task:", task.title)
  }

  const handleCreateTask = (taskData: Omit<ProjectTask, "id">) => {
    addTaskToProject(project.group.id, taskData as any)
  }

  const handleAddMember = (memberData: Omit<ProjectMember, "id">) => {
    addMemberToProject(project.group.id, memberData as any)
  }

  // const handleEditTask = (task: ProjectTask) => {
  //   setSelectedTask(task)
  //   setIsTaskEditModalOpen(true)
  // }

  const handleDeleteTask = (taskId: string) => {
    deleteTaskFromProject(project.group.id, taskId)
  }

  const handleRemoveMember = (memberId: string) => {
    if (confirm("Are you sure you want to remove this member from the project?")) {
      removeMemberFromProject(project.group.id, memberId)
    }
  }

  const formatDate = (date: Date) => {
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    })
  }

  return (
    <div className="h-full flex flex-col bg-white">
      {/* Header */}
      <div className="p-6 border-b border-gray-200 bg-white">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-4">
            <Button variant="ghost" size="sm" onClick={onBack} className="p-2">
              <ArrowLeft className="w-4 h-4" />
            </Button>
            <div>
              <div className="flex items-center space-x-3 mb-2">
                <div className="w-4 h-4 rounded-full bg-blue-500" />
                <h1 className="text-2xl font-bold text-gray-900">{project.group.name}</h1>
                <Badge
                  variant="outline"
                  className={`${
                    project.stats.completionRate === 100
                      ? "bg-green-100 text-green-800 border-green-200"
                      : project.stats.completionRate > 50
                        ? "bg-blue-100 text-blue-800 border-blue-200"
                        : "bg-yellow-100 text-yellow-800 border-yellow-200"
                  }`}
                >
                  {project.stats.completionRate}% Complete
                </Badge>
              </div>
              <p className="text-gray-600 max-w-2xl">{project.group.description}</p>
            </div>
          </div>
          <Button onClick={() => setIsTaskModalOpen(true)} className="bg-blue-600 hover:bg-blue-700 text-white">
            <Plus className="w-4 h-4 mr-2" />
            Add Task
          </Button>
        </div>

        {/* Simple Project Info */}
        <div className="flex items-center space-x-6 text-sm text-gray-600">
          <div className="flex items-center space-x-2">
            <Calendar className="w-4 h-4" />
            <span>
              {new Date(project.group.created_at).toLocaleDateString()} - {new Date(project.group.updated_at).toLocaleDateString()}
            </span>
          </div>
          <div className="flex items-center space-x-2">
            <Users className="w-4 h-4" />
            <span>{project.members.length} members</span>
          </div>
          <Badge
            variant="outline"
            className="bg-blue-100 text-blue-800 border-blue-200"
          >
            {project.group.type}
          </Badge>
        </div>
      </div>

      {/* Content Tabs */}
      <div className="flex-1 min-h-0">
        <Tabs value={activeTab} onValueChange={(value: any) => setActiveTab(value)} className="h-full flex flex-col">
          <div className="px-6 pt-4 border-b border-gray-200">
            <TabsList className="grid w-fit grid-cols-3">
              <TabsTrigger value="overview" className="flex items-center">
                <ListIcon className="w-4 h-4 mr-2" />
                Overview
              </TabsTrigger>
              <TabsTrigger value="gantt" className="flex items-center">
                <BarChart3 className="w-4 h-4 mr-2" />
                Gantt Chart
              </TabsTrigger>
              <TabsTrigger value="team" className="flex items-center">
                <Users className="w-4 h-4 mr-2" />
                Team
              </TabsTrigger>
            </TabsList>
          </div>

          {/* Overview Tab */}
          <TabsContent value="overview" className="flex-1 min-h-0 mt-0">
            <div className="p-6 h-full overflow-auto">
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Tasks Section */}
                <div className="lg:col-span-2">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold text-gray-900">Tasks</h3>
                    <Button onClick={() => setIsTaskModalOpen(true)} variant="outline" size="sm">
                      <Plus className="w-4 h-4 mr-2" />
                      Add Task
                    </Button>
                  </div>

                  {project.tasks.length === 0 ? (
                    <Card>
                      <CardContent className="p-8 text-center">
                        <ListIcon className="w-12 h-12 text-gray-300 mx-auto mb-4" />
                        <p className="text-gray-500 mb-2">No tasks in this project yet</p>
                        <p className="text-sm text-gray-400 mb-4">Get started by adding your first task</p>
                        <Button
                          onClick={() => setIsTaskModalOpen(true)}
                          className="bg-blue-600 hover:bg-blue-700 text-white"
                        >
                          <Plus className="w-4 h-4 mr-2" />
                          Add Task
                        </Button>
                      </CardContent>
                    </Card>
                  ) : (
                    <div className="space-y-3">
                      {project.tasks.map((task) => (
                        <Card key={task.id} className="hover:shadow-md transition-all">
                          <CardContent className="p-4">
                            <div className="flex items-center justify-between mb-2">
                              <h4 className="font-medium text-gray-900">{task.title}</h4>
                              <div className="flex items-center space-x-2">
                                <Badge
                                  variant="outline"
                                  className={
                                    task.priority === "HIGH"
                                      ? "bg-red-100 text-red-800 border-red-200"
                                      : task.priority === "MEDIUM"
                                        ? "bg-yellow-100 text-yellow-800 border-yellow-200"
                                        : "bg-green-100 text-green-800 border-green-200"
                                  }
                                >
                                  {task.priority}
                                </Badge>
                                <Badge
                                  variant="outline"
                                  className={
                                    task.status === "DONE"
                                      ? "bg-green-100 text-green-800 border-green-200"
                                      : task.status === "IN_PROGRESS"
                                        ? "bg-blue-100 text-blue-800 border-blue-200"
                                        : task.status === "TODO"
                                          ? "bg-red-100 text-red-800 border-red-200"
                                          : "bg-gray-100 text-gray-800 border-gray-200"
                                  }
                                >
                                  {task.status}
                                </Badge>
                                <DropdownMenu>
                                  <DropdownMenuTrigger asChild>
                                    <Button variant="ghost" size="sm" className="p-1">
                                      <MoreHorizontal className="w-4 h-4" />
                                    </Button>
                                  </DropdownMenuTrigger>
                                  <DropdownMenuContent align="end">
                                    <DropdownMenuItem onClick={() => {/* TODO: Implement edit */}}>
                                      <Edit className="w-4 h-4 mr-2" />
                                      Edit Task
                                    </DropdownMenuItem>
                                    <DropdownMenuItem
                                      onClick={() => handleDeleteTask(task.id)}
                                      className="text-red-600"
                                    >
                                      <Trash2 className="w-4 h-4 mr-2" />
                                      Delete Task
                                    </DropdownMenuItem>
                                  </DropdownMenuContent>
                                </DropdownMenu>
                              </div>
                            </div>
                            <p className="text-sm text-gray-600 mb-3">{task.description}</p>
                            <div className="flex items-center justify-between">
                              <div className="flex items-center space-x-4 text-sm text-gray-500">
                                <div className="flex items-center space-x-1">
                                  <Calendar className="w-4 h-4" />
                                  <span>{task.end_date ? formatDate(new Date(task.end_date)) : 'No end date'}</span>
                                </div>
                                {task.assignee_id && (
                                  <div className="flex items-center space-x-1">
                                    <Users className="w-4 h-4" />
                                    <span>Assigned to: {task.assignee_id}</span>
                                  </div>
                                )}
                              </div>
                              <div className="text-sm text-gray-600">{task.progress}% complete</div>
                            </div>
                          </CardContent>
                        </Card>
                      ))}
                    </div>
                  )}
                </div>

                {/* Project Info Sidebar */}
                <div className="space-y-6">
                  {/* Project Details */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">Project Details</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div>
                        <p className="text-sm font-medium text-gray-600">Timeline</p>
                        <p className="text-sm text-gray-900 mt-1">
                          {new Date(project.group.created_at).toLocaleDateString()} - {new Date(project.group.updated_at).toLocaleDateString()}
                        </p>
                      </div>
                      {['React', 'TypeScript', 'Frontend'].length > 0 && (
                        <div>
                          <p className="text-sm font-medium text-gray-600 mb-2">Tags</p>
                          <div className="flex flex-wrap gap-1">
                            {['React', 'TypeScript', 'Frontend'].map((tag: string) => (
                              <Badge key={tag} variant="outline" className="text-xs bg-gray-50 text-gray-600">
                                {tag}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      )}
                    </CardContent>
                  </Card>

                  {/* Team Members */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">Team Members</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-3">
                        {project.members.slice(0, 5).map((member) => (
                          <div key={member.id} className="flex items-center space-x-3">
                            <Avatar className="w-8 h-8">
                              <AvatarImage src={member.user?.username || "/placeholder.svg"} />
                              <AvatarFallback className="bg-blue-100 text-blue-600 text-xs">
                                {member.user?.username
                                  .split(" ")
                                  .map((n) => n[0])
                                  .join("")
                                  .toUpperCase()}
                              </AvatarFallback>
                            </Avatar>
                            <div className="flex-1">
                              <p className="text-sm font-medium text-gray-900">{member.user?.username || 'Unknown'}</p>
                              <p className="text-xs text-gray-500">{member.role}</p>
                            </div>
                          </div>
                        ))}
                        {project.members.length > 5 && (
                          <p className="text-xs text-gray-500 text-center pt-2">
                            +{project.members.length - 5} more members
                          </p>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </div>
            </div>
          </TabsContent>

          {/* Gantt Chart Tab */}
          <TabsContent value="gantt" className="flex-1 min-h-0 mt-0">
            <GanttChart tasks={project.tasks} onTaskSelect={handleTaskSelect} />
          </TabsContent>

          {/* Team Tab */}
          <TabsContent value="team" className="flex-1 min-h-0 mt-0">
            <div className="p-6 h-full overflow-auto">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-semibold text-gray-900">Team Members</h3>
                <Button onClick={() => setIsMemberModalOpen(true)} className="bg-blue-600 hover:bg-blue-700 text-white">
                  <Plus className="w-4 h-4 mr-2" />
                  Add Member
                </Button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {project.members.map((member) => (
                  <Card key={member.id}>
                    <CardContent className="p-4">
                      <div className="flex items-center space-x-3 mb-3">
                        <Avatar className="w-12 h-12">
                          <AvatarImage src="/placeholder.svg" />
                          <AvatarFallback className="bg-blue-100 text-blue-600">
                            {(member.user?.username || 'U')
                              .split(" ")
                              .map((n: string) => n[0])
                              .join("")
                              .toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <div className="flex-1">
                          <h4 className="font-medium text-gray-900">{member.user?.username || 'Unknown'}</h4>
                          <p className="text-sm text-gray-500">{member.role}</p>
                        </div>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm" className="p-1">
                              <MoreHorizontal className="w-4 h-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleRemoveMember(member.id)} className="text-red-600">
                              <Trash2 className="w-4 h-4 mr-2" />
                              Remove Member
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                      <div className="text-sm text-gray-600">
                        <p>Role: {member.role}</p>
                        <p className="mt-1">
                          Tasks: {project.tasks.filter((task) => task.assignee_id === member.user_id).length}
                        </p>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      {/* Modals */}
      <ProjectTaskCreationModal
        isOpen={isTaskModalOpen}
        onClose={() => setIsTaskModalOpen(false)}
        onCreateTask={handleCreateTask}
        project={project}
      />

      <ProjectMemberModal
        isOpen={isMemberModalOpen}
        onClose={() => setIsMemberModalOpen(false)}
        onAddMember={handleAddMember}
      />

      {/* TODO: Add TaskEditModal when implemented */}
    </div>
  )
}

export default ProjectDetailPage
