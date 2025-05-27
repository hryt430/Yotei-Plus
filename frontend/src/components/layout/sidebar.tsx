"use client"

import { useState } from "react"
import { Button } from "@/components/ui/forms/button"
import { Home, List, Plus, ChevronLeft, ChevronRight, Calendar, Search, Users } from "lucide-react"

interface SidebarProps {
  currentPage: "dashboard" | "tasks"
  onNavigate: (page: "dashboard" | "tasks") => void
  onCreateTask: () => void
  onAddFriend: () => void
}

export function Sidebar({ currentPage, onNavigate, onCreateTask, onAddFriend }: SidebarProps) {
  const [isCollapsed, setIsCollapsed] = useState(false)

  const menuItems = [
    {
      id: "dashboard",
      label: "Dashboard",
      icon: Home,
      page: "dashboard" as const,
    },
    {
      id: "tasks",
      label: "All Tasks",
      icon: List,
      page: "tasks" as const,
    },
  ]

  return (
    <div
      className={`bg-white border-r border-gray-200 transition-all duration-300 flex-shrink-0 ${isCollapsed ? "w-16" : "w-64"}`}
    >
      <div className="flex flex-col h-full">
        {/* Header */}
        <div className="p-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            {!isCollapsed && (
              <div>
                <h2 className="font-semibold text-gray-900">TaskFlow</h2>
                <p className="text-xs text-gray-500">Task Management</p>
              </div>
            )}
            <Button variant="ghost" size="sm" onClick={() => setIsCollapsed(!isCollapsed)} className="p-2">
              {isCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
            </Button>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="p-4 space-y-3">
          <Button
            onClick={onCreateTask}
            className={`w-full bg-gray-900 hover:bg-gray-800 ${isCollapsed ? "px-2" : ""}`}
            size={isCollapsed ? "sm" : "default"}
          >
            <Plus className="w-4 h-4" />
            {!isCollapsed && <span className="ml-2">New Task</span>}
          </Button>

          <Button
            onClick={onAddFriend}
            variant="outline"
            className={`w-full border-gray-300 hover:bg-gray-50 ${isCollapsed ? "px-2" : ""}`}
            size={isCollapsed ? "sm" : "default"}
          >
            <Users className="w-4 h-4" />
            {!isCollapsed && <span className="ml-2">Add Friend</span>}
          </Button>
        </div>

        {/* Navigation Menu */}
        <nav className="flex-1 px-4 space-y-2">
          {menuItems.map((item) => {
            const Icon = item.icon
            const isActive = currentPage === item.page

            return (
              <Button
                key={item.id}
                variant={isActive ? "secondary" : "ghost"}
                className={`w-full justify-start ${isCollapsed ? "px-2" : ""} ${isActive ? "bg-gray-100" : ""}`}
                onClick={() => onNavigate(item.page)}
              >
                <Icon className="w-4 h-4" />
                {!isCollapsed && <span className="ml-3">{item.label}</span>}
              </Button>
            )
          })}
        </nav>

        {/* Footer */}
        {!isCollapsed && (
          <div className="p-4 border-t border-gray-200">
            <div className="text-xs text-gray-500">
              <div className="flex items-center mb-1">
                <Calendar className="w-3 h-3 mr-1" />
                Dashboard: Quick overview
              </div>
              <div className="flex items-center">
                <Search className="w-3 h-3 mr-1" />
                All Tasks: Advanced search
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
