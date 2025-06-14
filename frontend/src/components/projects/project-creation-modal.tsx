"use client"

import type React from "react"
import { useState } from "react"
import { Button } from "@/components/ui/forms/button"
import { Input } from "@/components/ui/forms/input"
import { Textarea } from "@/components/ui/forms/textarea"
import { Label } from "@/components/ui/forms/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/forms/select"
import { X, ChevronDown, ChevronUp, Calendar, Target, Folder } from "lucide-react"
import type { CreateProjectRequest } from "@/types"

interface ProjectCreationModalProps {
  isOpen: boolean
  onClose: () => void
  onCreateProject: (project: CreateProjectRequest) => void
}

export function ProjectCreationModal({ isOpen, onClose, onCreateProject }: ProjectCreationModalProps) {
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    startDate: new Date().toISOString().split("T")[0],
    endDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split("T")[0], // 30 days from now
    priority: "MEDIUM" as "LOW" | "MEDIUM" | "HIGH",
  })
  const [showAdvanced, setShowAdvanced] = useState(false)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.name.trim()) return

    const project: CreateProjectRequest = {
      name: formData.name.trim(),
      description: formData.description.trim(),
      start_date: formData.startDate,
      end_date: formData.endDate,
    }

    onCreateProject(project)
    handleClose()
  }

  const handleClose = () => {
    onClose()
    // フォームをリセット
    setFormData({
      name: "",
      description: "",
      startDate: new Date().toISOString().split("T")[0],
      endDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split("T")[0],
      priority: "MEDIUM",
    })
    setShowAdvanced(false)
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* バックドロップ */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={handleClose} />

      {/* モーダル */}
      <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-md mx-4 overflow-hidden">
        {/* ヘッダー */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 bg-gradient-to-r from-white to-gray-50/30">
          <div className="flex items-center">
            <Folder className="w-6 h-6 mr-3 text-blue-600" />
            <div>
              <h2 className="text-xl font-semibold text-gray-900">新規プロジェクト</h2>
              <p className="text-sm text-gray-600">基本情報を入力してください</p>
            </div>
          </div>
          <Button variant="ghost" size="sm" onClick={handleClose}>
            <X className="w-4 h-4" />
          </Button>
        </div>

        {/* コンテンツ */}
        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {/* プロジェクト名 */}
          <div className="space-y-2">
            <Label htmlFor="name" className="text-sm font-medium text-gray-700">
              プロジェクト名 *
            </Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="プロジェクト名を入力..."
              className="border-gray-200 focus:border-blue-300 focus:ring-blue-300"
              required
            />
          </div>

          {/* 説明 */}
          <div className="space-y-2">
            <Label htmlFor="description" className="text-sm font-medium text-gray-700">
              説明
            </Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="プロジェクトの目的や概要を入力..."
              className="border-gray-200 focus:border-blue-300 focus:ring-blue-300 min-h-[100px]"
              rows={4}
            />
          </div>

          {/* 詳細設定（折りたたみ可能） */}
          <div className="space-y-4">
            <button
              type="button"
              className="flex items-center justify-between w-full text-sm font-medium text-gray-700 hover:text-gray-900"
              onClick={() => setShowAdvanced(!showAdvanced)}
            >
              <span>詳細設定</span>
              {showAdvanced ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
            </button>

            {showAdvanced && (
              <div className="space-y-4 pt-2 pl-2 border-l-2 border-gray-100">
                {/* 優先度 */}
                <div className="space-y-2">
                  <Label htmlFor="priority" className="text-sm font-medium text-gray-700 flex items-center">
                    <Target className="w-4 h-4 mr-2 text-gray-500" />
                    優先度
                  </Label>
                  <Select
                    value={formData.priority}
                    onValueChange={(value: any) => setFormData({ ...formData, priority: value })}
                  >
                    <SelectTrigger className="border-gray-200">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="LOW">低</SelectItem>
                      <SelectItem value="MEDIUM">中</SelectItem>
                      <SelectItem value="HIGH">高</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {/* 期間 */}
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label htmlFor="startDate" className="text-sm font-medium text-gray-700 flex items-center">
                      <Calendar className="w-4 h-4 mr-2 text-gray-500" />
                      開始日
                    </Label>
                    <Input
                      id="startDate"
                      type="date"
                      value={formData.startDate}
                      onChange={(e) => setFormData({ ...formData, startDate: e.target.value })}
                      className="border-gray-200 focus:border-blue-300 focus:ring-blue-300"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="endDate" className="text-sm font-medium text-gray-700 flex items-center">
                      <Calendar className="w-4 h-4 mr-2 text-gray-500" />
                      終了日
                    </Label>
                    <Input
                      id="endDate"
                      type="date"
                      value={formData.endDate}
                      onChange={(e) => setFormData({ ...formData, endDate: e.target.value })}
                      className="border-gray-200 focus:border-blue-300 focus:ring-blue-300"
                    />
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* フッター */}
          <div className="flex items-center justify-end space-x-3 pt-4">
            <Button variant="outline" type="button" onClick={handleClose}>
              キャンセル
            </Button>
            <Button type="submit" disabled={!formData.name.trim()} className="bg-blue-600 hover:bg-blue-700 text-white">
              <Folder className="w-4 h-4 mr-2" />
              作成
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default ProjectCreationModal
