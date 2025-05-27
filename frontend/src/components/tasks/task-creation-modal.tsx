"use client"

import type React from "react"

import { useState } from "react"
import type { Task } from "@/types/task"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Label } from "@/components/ui/label"
import { X, Calendar, Flag, Tag } from "lucide-react"

interface TaskCreationModalProps {
  isOpen: boolean
  onClose: () => void
  onCreateTask: (task: Omit<Task, "id">) => void
}

export function TaskCreationModal({ isOpen, onClose, onCreateTask }: TaskCreationModalProps) {
  const [formData, setFormData] = useState({
    title: "",
    description: "",
    dueDate: new Date().toISOString().split("T")[0], // Today's date in YYYY-MM-DD format
    priority: "medium" as "low" | "medium" | "high",
    category: "",
    status: "pending" as "pending" | "in-progress" | "completed",
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.title.trim()) return

    onCreateTask({
      title: formData.title.trim(),
      description: formData.description.trim(),
      dueDate: new Date(formData.dueDate),
      priority: formData.priority,
      category: formData.category.trim() || "General",
      status: formData.status,
    })

    // Reset form
    setFormData({
      title: "",
      description: "",
      dueDate: new Date().toISOString().split("T")[0],
      priority: "medium",
      category: "",
      status: "pending",
    })
  }

  const handleClose = () => {
    onClose()
    // Reset form when closing
    setFormData({
      title: "",
      description: "",
      dueDate: new Date().toISOString().split("T")[0],
      priority: "medium",
      category: "",
      status: "pending",
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
            <Label htmlFor="dueDate" className="text-sm font-medium text-gray-700 flex items-center">
              <Calendar className="w-4 h-4 mr-2" />
              Due Date
            </Label>
            <Input
              id="dueDate"
              type="date"
              value={formData.dueDate}
              onChange={(e) => setFormData({ ...formData, dueDate: e.target.value })}
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
                onValueChange={(value: any) => setFormData({ ...formData, priority: value })}
              >
                <SelectTrigger className="border-gray-200">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Category */}
            <div className="space-y-2">
              <Label htmlFor="category" className="text-sm font-medium text-gray-700 flex items-center">
                <Tag className="w-4 h-4 mr-2" />
                Category
              </Label>
              <Input
                id="category"
                value={formData.category}
                onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                placeholder="e.g., Work, Personal"
                className="border-gray-200 focus:border-gray-300 focus:ring-gray-300"
              />
            </div>
          </div>

          {/* Status */}
          <div className="space-y-2">
            <Label className="text-sm font-medium text-gray-700">Status</Label>
            <Select value={formData.status} onValueChange={(value: any) => setFormData({ ...formData, status: value })}>
              <SelectTrigger className="border-gray-200">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="in-progress">In Progress</SelectItem>
                <SelectItem value="completed">Completed</SelectItem>
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
