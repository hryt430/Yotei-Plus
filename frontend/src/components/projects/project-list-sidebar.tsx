"use client"

import { useState } from "react"
import { Button } from "@/components/ui/forms/button"
import { Badge } from "@/components/ui/data-display/badge"
import { ScrollArea } from "@/components/ui/layout/scroll-area"
// import { DragDropContext, Droppable, Draggable } from "@hello-pangea/dnd"
import { Plus, MoreHorizontal, ChevronRight, ChevronDown } from "lucide-react"
import type { ProjectView } from "@/types"

interface ProjectListSidebarProps {
  projects: ProjectView[]
  onCreateProject: () => void
  onSelectProject: (project: ProjectView) => void
  onReorderProjects: (projects: ProjectView[]) => void
}

export function ProjectListSidebar({
  projects,
  onCreateProject,
  onSelectProject,
  onReorderProjects,
}: ProjectListSidebarProps) {
  const [isExpanded, setIsExpanded] = useState(true)

  // const handleDragEnd = (result: any) => {
  //   if (!result.destination) return

  //   const items = Array.from(projects)
  //   const [reorderedItem] = items.splice(result.source.index, 1)
  //   items.splice(result.destination.index, 0, reorderedItem)

  //   onReorderProjects(items)
  // }

  const getStatusColor = (status: string) => {
    switch (status) {
      case "planning":
        return "bg-blue-100 text-blue-800 border-blue-200"
      case "active":
        return "bg-green-100 text-green-800 border-green-200"
      case "on-hold":
        return "bg-yellow-100 text-yellow-800 border-yellow-200"
      case "completed":
        return "bg-gray-100 text-gray-800 border-gray-200"
      case "cancelled":
        return "bg-red-100 text-red-800 border-red-200"
      default:
        return "bg-gray-100 text-gray-800 border-gray-200"
    }
  }

  return (
    <div className="border-b border-gray-200 bg-gray-50/30">
      {/* ヘッダー */}
      <div className="p-4 flex items-center justify-between">
        <button
          className="flex items-center text-sm font-medium text-gray-700 hover:text-gray-900"
          onClick={() => setIsExpanded(!isExpanded)}
        >
          {isExpanded ? <ChevronDown className="w-4 h-4 mr-2" /> : <ChevronRight className="w-4 h-4 mr-2" />}
          プロジェクト ({projects.length})
        </button>
        <Button
          size="sm"
          variant="ghost"
          onClick={onCreateProject}
          className="h-7 w-7 p-0 rounded-full hover:bg-gray-200"
        >
          <Plus className="w-4 h-4" />
          <span className="sr-only">新規プロジェクト</span>
        </Button>
      </div>

      {/* プロジェクトリスト */}
      {isExpanded && (
        <div className="px-2 pb-2">
          <ScrollArea className="h-[200px]">
            {/* Drag and Drop temporarily disabled */}
            <div className="space-y-1">
              {projects.map((project) => (
                <div
                  key={project.group.id}
                  className="flex items-center justify-between p-2 rounded-md text-sm hover:bg-gray-100"
                  onClick={() => onSelectProject(project as any)}
                >
                            <div className="flex items-center space-x-2 overflow-hidden">
                              <div
                                className="w-2 h-2 rounded-full flex-shrink-0"
                                style={{ backgroundColor: '#3B82F6' }}
                              />
                              <span className="truncate">{project.group.name}</span>
                            </div>
                            <div className="flex items-center space-x-1">
                              <Badge variant="outline" className={`text-xs`}>
                                {Math.round(project.stats.completionRate)}%
                              </Badge>
                              <Button variant="ghost" size="sm" className="h-6 w-6 p-0 hover:bg-gray-200 rounded-full">
                                <MoreHorizontal className="w-3 h-3" />
                              </Button>
                            </div>
                          </div>
                ))}

                {/* 新規プロジェクト作成ボタン */}
                <button
                  onClick={onCreateProject}
                  className="w-full flex items-center justify-center p-2 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-md mt-2"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  新規プロジェクト
                </button>
              </div>
          </ScrollArea>
        </div>
      )}
    </div>
  )
}

export default ProjectListSidebar
