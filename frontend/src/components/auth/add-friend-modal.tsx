"use client"

import type React from "react"
import { useState } from "react"
import { Button } from "@/components/ui/forms/button"
import { Input } from "@/components/ui/forms/input"
import { Label } from "@/components/ui/forms/label"
import { Badge } from "@/components/ui/data-display/badge"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/data-display/avatar"
import { X, Search, UserPlus, Mail, Users, Check } from "lucide-react"

interface Friend {
  id: string
  name: string
  email: string
  avatar?: string
  status: "pending" | "accepted" | "suggested"
}

interface AddFriendModalProps {
  isOpen: boolean
  onClose: () => void
}

export function AddFriendModal({ isOpen, onClose }: AddFriendModalProps) {
  const [searchTerm, setSearchTerm] = useState("")
  const [inviteEmail, setInviteEmail] = useState("")
  const [friends, setFriends] = useState<Friend[]>([
    {
      id: "1",
      name: "Sarah Johnson",
      email: "sarah.johnson@example.com",
      status: "suggested",
    },
    {
      id: "2",
      name: "Mike Chen",
      email: "mike.chen@example.com",
      status: "pending",
    },
    {
      id: "3",
      name: "Emily Davis",
      email: "emily.davis@example.com",
      status: "accepted",
    },
    {
      id: "4",
      name: "Alex Rodriguez",
      email: "alex.rodriguez@example.com",
      status: "suggested",
    },
  ])

  const handleSendInvite = (e: React.FormEvent) => {
    e.preventDefault()
    if (!inviteEmail.trim()) return

    // Add new friend invitation
    const newFriend: Friend = {
      id: Date.now().toString(),
      name: inviteEmail.split("@")[0],
      email: inviteEmail.trim(),
      status: "pending",
    }

    setFriends((prev) => [...prev, newFriend])
    setInviteEmail("")
  }

  const handleAddFriend = (friendId: string) => {
    setFriends((prev) => prev.map((friend) => (friend.id === friendId ? { ...friend, status: "pending" } : friend)))
  }

  const filteredFriends = friends.filter(
    (friend) =>
      friend.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      friend.email.toLowerCase().includes(searchTerm.toLowerCase()),
  )

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "pending":
        return (
          <Badge variant="outline" className="text-yellow-600 border-yellow-200 bg-yellow-50">
            Pending
          </Badge>
        )
      case "accepted":
        return (
          <Badge variant="outline" className="text-green-600 border-green-200 bg-green-50">
            <Check className="w-3 h-3 mr-1" />
            Friends
          </Badge>
        )
      case "suggested":
        return (
          <Badge variant="outline" className="text-blue-600 border-blue-200 bg-blue-50">
            Suggested
          </Badge>
        )
      default:
        return null
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} />

      {/* Modal */}
      <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4 max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 bg-gradient-to-r from-white to-gray-50/30">
          <div className="flex items-center">
            <Users className="w-5 h-5 mr-3 text-blue-600" />
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Add Friends</h2>
              <p className="text-sm text-gray-600">Connect with colleagues and friends</p>
            </div>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="w-4 h-4" />
          </Button>
        </div>

        <div className="p-6 space-y-6 max-h-[calc(90vh-120px)] overflow-y-auto">
          {/* Invite by Email */}
          <div className="space-y-3">
            <Label className="text-sm font-medium text-gray-700 flex items-center">
              <Mail className="w-4 h-4 mr-2" />
              Invite by Email
            </Label>
            <form onSubmit={handleSendInvite} className="flex space-x-2">
              <Input
                type="email"
                value={inviteEmail}
                onChange={(e) => setInviteEmail(e.target.value)}
                placeholder="Enter email address..."
                className="flex-1 border-gray-200 focus:border-blue-300 focus:ring-blue-300"
              />
              <Button type="submit" className="bg-blue-600 hover:bg-blue-700">
                <UserPlus className="w-4 h-4" />
              </Button>
            </form>
          </div>

          {/* Search Friends */}
          <div className="space-y-3">
            <Label className="text-sm font-medium text-gray-700">Find People</Label>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
              <Input
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search by name or email..."
                className="pl-10 border-gray-200 focus:border-blue-300 focus:ring-blue-300"
              />
            </div>
          </div>

          {/* Friends List */}
          <div className="space-y-3">
            <Label className="text-sm font-medium text-gray-700">People ({filteredFriends.length})</Label>
            <div className="space-y-3 max-h-64 overflow-y-auto">
              {filteredFriends.length === 0 ? (
                <div className="text-center py-8">
                  <Users className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                  <div className="text-gray-400 mb-2">No people found</div>
                  <p className="text-sm text-gray-500">Try searching with different terms</p>
                </div>
              ) : (
                filteredFriends.map((friend) => (
                  <div
                    key={friend.id}
                    className="flex items-center justify-between p-3 bg-gray-50/50 rounded-lg border border-gray-200/60 hover:bg-gray-100/50 transition-colors"
                  >
                    <div className="flex items-center space-x-3">
                      <Avatar className="w-10 h-10">
                        <AvatarImage src={friend.avatar || "/placeholder.svg"} />
                        <AvatarFallback className="bg-blue-100 text-blue-600 font-medium">
                          {friend.name
                            .split(" ")
                            .map((n) => n[0])
                            .join("")
                            .toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium text-gray-900">{friend.name}</div>
                        <div className="text-sm text-gray-500">{friend.email}</div>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      {getStatusBadge(friend.status)}
                      {friend.status === "suggested" && (
                        <Button
                          size="sm"
                          onClick={() => handleAddFriend(friend.id)}
                          className="bg-blue-600 hover:bg-blue-700"
                        >
                          <UserPlus className="w-3 h-3 mr-1" />
                          Add
                        </Button>
                      )}
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-6 border-t border-gray-200 bg-gray-50/30">
          <div className="flex justify-between items-center">
            <div className="text-sm text-gray-500">
              {friends.filter((f) => f.status === "accepted").length} friends connected
            </div>
            <Button onClick={onClose} variant="outline">
              Done
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
