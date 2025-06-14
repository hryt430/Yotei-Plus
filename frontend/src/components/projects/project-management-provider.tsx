"use client"

import type React from "react"
import { createContext, useState, useContext, type ReactNode } from "react"

// Define data types
export interface ProjectTask {
  id: string
  name: string
  description: string
  status: "open" | "in progress" | "completed"
}

export interface ProjectMember {
  id: string
  name: string
  role: string
}

export interface Project {
  id: string
  name: string
  description: string
  tasks: ProjectTask[]
  members: ProjectMember[]
}

// Define context type
interface ProjectManagementContextType {
  projects: Project[]
  addProject: (projectData: Omit<Project, "id" | "tasks" | "members">) => void
  updateProject: (projectId: string, projectData: Omit<Project, "id" | "tasks" | "members">) => void
  deleteProject: (projectId: string) => void
  getProject: (projectId: string) => Project | undefined
  addTaskToProject: (projectId: string, taskData: Omit<ProjectTask, "id">) => void
  deleteTaskFromProject: (projectId: string, taskId: string) => void
  addMemberToProject: (projectId: string, memberData: Omit<ProjectMember, "id">) => void
  removeMemberFromProject: (projectId: string, memberId: string) => void
}

// Create context
const ProjectManagementContext = createContext<ProjectManagementContextType | undefined>(undefined)

// Create provider component
interface ProjectManagementProviderProps {
  children: ReactNode
}

export const ProjectManagementProvider: React.FC<ProjectManagementProviderProps> = ({ children }) => {
  const [projects, setProjects] = useState<Project[]>([])

  const addProject = (projectData: Omit<Project, "id" | "tasks" | "members">) => {
    const newProject: Project = {
      ...projectData,
      id: `project_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      tasks: [],
      members: [],
    }
    setProjects([...projects, newProject])
  }

  const updateProject = (projectId: string, projectData: Omit<Project, "id" | "tasks" | "members">) => {
    setProjects((prev) => prev.map((project) => (project.id === projectId ? { ...project, ...projectData } : project)))
  }

  const deleteProject = (projectId: string) => {
    setProjects((prev) => prev.filter((project) => project.id !== projectId))
  }

  const getProject = (projectId: string) => {
    return projects.find((project) => project.id === projectId)
  }

  const addTaskToProject = (projectId: string, taskData: Omit<ProjectTask, "id">) => {
    setProjects((prev) =>
      prev.map((project) => {
        if (project.id === projectId) {
          const newTask: ProjectTask = {
            ...taskData,
            id: `task_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
          }
          return {
            ...project,
            tasks: [...project.tasks, newTask],
          }
        }
        return project
      }),
    )
  }

  const deleteTaskFromProject = (projectId: string, taskId: string) => {
    setProjects((prev) =>
      prev.map((project) => {
        if (project.id === projectId) {
          return {
            ...project,
            tasks: project.tasks.filter((task) => task.id !== taskId),
          }
        }
        return project
      }),
    )
  }

  const addMemberToProject = (projectId: string, memberData: Omit<ProjectMember, "id">) => {
    setProjects((prev) =>
      prev.map((project) => {
        if (project.id === projectId) {
          const newMember: ProjectMember = {
            ...memberData,
            id: `member_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
          }
          return {
            ...project,
            members: [...project.members, newMember],
          }
        }
        return project
      }),
    )
  }

  const removeMemberFromProject = (projectId: string, memberId: string) => {
    setProjects((prev) =>
      prev.map((project) => {
        if (project.id === projectId) {
          return {
            ...project,
            members: project.members.filter((member) => member.id !== memberId),
          }
        }
        return project
      }),
    )
  }

  const value: ProjectManagementContextType = {
    projects,
    addProject,
    updateProject,
    deleteProject,
    getProject,
    addTaskToProject,
    deleteTaskFromProject,
    addMemberToProject,
    removeMemberFromProject,
  }

  return <ProjectManagementContext.Provider value={value}>{children}</ProjectManagementContext.Provider>
}

// Create custom hook
export const useProjectManagement = () => {
  const context = useContext(ProjectManagementContext)
  if (!context) {
    throw new Error("useProjectManagement must be used within a ProjectManagementProvider")
  }
  return context
}
