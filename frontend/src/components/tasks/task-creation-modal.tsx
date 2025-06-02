"use client"

import type React from "react"

import { useState } from "react"
import type { Task, TaskStatus, TaskPriority, TaskCategory, TaskRequest } from "@/types"
import { Button } from "@/components/ui/forms/button"
import { Input } from "@/components/ui/forms/input"
import { Textarea } from "@/components/ui/forms/textarea"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/forms/select"
import { Label } from "@/components/ui/forms/label"
import { X, Calendar, Flag, Tag } from "lucide-react"

interface TaskCreationModalProps {
  isOpen: boolean
  onClose: () => void
  onCreateTask: (task: TaskRequest) => void
}

export function TaskCreationModal({ isOpen, onClose, onCreateTask }: TaskCreationModalProps) {
  const [formData, setFormData] = useState({
    title: "",
    description: "",
    due_date: new Date().toISOString().split("T")[0], // Today's date in YYYY-MM-DD format
    priority: "MEDIUM" as TaskPriority,
    category: "OTHER" as TaskCategory,
    status: "TODO" as TaskStatus,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.title.trim()) return

    const taskRequest: TaskRequest = {
      title: formData.title.trim(),
      description: formData.description.trim(),
      due_date: formData.due_date,
      priority: formData.priority,
      category: formData.category,
      status: formData.status,
    }

    onCreateTask(taskRequest)

    // Reset form
    setFormData({
      title: "",
      description: "",
      due_date: new Date().toISOString().split("T")[0],
      priority: "MEDIUM",
      category: "OTHER",
      status: "TODO",
    })
  }

  const handleClose = () => {
    onClose()
    // Reset form when closing
    setFormData({
      title: "",
      description: "",
      due_date: new Date().toISOString().split("T")[0],
      priority: "MEDIUM",
      category: "OTHER",
      status: "TODO",
    })
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={handleClose} />

      {/* Modal */}
      <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Create New Task</h2>
          <Button variant="ghost" size="sm" onClick={handleClose}>
            <X className="w-4 h-4" />
          </Button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {/* Title */}
          <div className="space-y-2">
            <Label htmlFor="title" className="text-sm font-medium text-gray-700">
              Task Title *
            </Label>
            <Input
              id="title"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder="Enter task title..."
              className="border-gray-200 focus:border-gray-300 focus:ring-gray-300"
              required
            />
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description" className="text-sm font-medium text-gray-700">
              Description
            </Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Enter task description..."
              className="border-gray-200 focus:border-gray-300 focus:ring-gray-300 min-h-[80px]"
              rows={3}
            />
          </div>

          {/* Due Date */}
          <div className="space-y-2">
            <Label htmlFor="due_date" className="text-sm font-medium text-gray-700 flex items-center">
              <Calendar className="w-4 h-4 mr-2" />
              Due Date
            </Label>
            <Input
              id="due_date"
              type="date"
              value={formData.due_date}
              onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
              className="border-gray-200 focus:border-gray-300 focus:ring-gray-300"
            />
          </div>

          {/* Priority and Category Row */}
          <div className="grid grid-cols-2 gap-4">
            {/* Priority */}
            <div className="space-y-2">
              <Label className="text-sm font-medium text-gray-700 flex items-center">
                <Flag className="w-4 h-4 mr-2" />
                Priority
              </Label>
              <Select
                value={formData.priority}
                onValueChange={(value: TaskPriority) => setFormData({ ...formData, priority: value })}
              >
                <SelectTrigger className="border-gray-200">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="LOW">Low</SelectItem>
                  <SelectItem value="MEDIUM">Medium</SelectItem>
                  <SelectItem value="HIGH">High</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Category */}
            <div className="space-y-2">
              <Label className="text-sm font-medium text-gray-700 flex items-center">
                <Tag className="w-4 h-4 mr-2" />
                Category
              </Label>
              <Select
                value={formData.category}
                onValueChange={(value: TaskCategory) => setFormData({ ...formData, category: value })}
              >
                <SelectTrigger className="border-gray-200">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="WORK">Work</SelectItem>
                  <SelectItem value="PERSONAL">Personal</SelectItem>
                  <SelectItem value="STUDY">Study</SelectItem>
                  <SelectItem value="HEALTH">Health</SelectItem>
                  <SelectItem value="SHOPPING">Shopping</SelectItem>
                  <SelectItem value="OTHER">Other</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Status */}
          <div className="space-y-2">
            <Label className="text-sm font-medium text-gray-700">Status</Label>
            <Select 
              value={formData.status} 
              onValueChange={(value: TaskStatus) => setFormData({ ...formData, status: value })}
            >
              <SelectTrigger className="border-gray-200">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="TODO">Todo</SelectItem>
                <SelectItem value="IN_PROGRESS">In Progress</SelectItem>
                <SelectItem value="DONE">Done</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Actions */}
          <div className="flex space-x-3 pt-4">
            <Button type="button" variant="outline" onClick={handleClose} className="flex-1">
              Cancel
            </Button>
            <Button type="submit" className="flex-1 bg-gray-900 hover:bg-gray-800">
              Create Task
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}