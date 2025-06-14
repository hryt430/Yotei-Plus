"use client"

import type React from "react"
import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Textarea } from "@/components/ui/textarea"
import {
  X,
  Search,
  UserPlus,
  Mail,
  Users,
  Check,
  MoreHorizontal,
  MessageCircle,
  Calendar,
  Building,
  Clock,
  UserCheck,
  Send,
} from "lucide-react"
import type { Friend, FriendRequest, FriendSearchResult } from "@/types/friend"

interface FriendManagementModalProps {
  isOpen: boolean
  onClose: () => void
  friends: Friend[]
  friendRequests: FriendRequest[]
  onSendFriendRequest: (email: string, message?: string) => void
  onAcceptFriendRequest: (requestId: string) => void
  onRejectFriendRequest: (requestId: string) => void
  onRemoveFriend: (friendId: string) => void
  onBlockUser: (userId: string) => void
}

export function FriendManagementModal({
  isOpen,
  onClose,
  friends,
  friendRequests,
  onSendFriendRequest,
  onAcceptFriendRequest,
  onRejectFriendRequest,
  onRemoveFriend,
  onBlockUser,
}: FriendManagementModalProps) {
  const [activeTab, setActiveTab] = useState("friends")
  const [searchTerm, setSearchTerm] = useState("")
  const [inviteEmail, setInviteEmail] = useState("")
  const [inviteMessage, setInviteMessage] = useState("")
  const [searchResults, setSearchResults] = useState<FriendSearchResult[]>([])
  const [isSearching, setIsSearching] = useState(false)

  // Mock search function - in real app, this would call an API
  const handleSearch = async () => {
    if (!searchTerm.trim()) return

    setIsSearching(true)
    // Simulate API call
    setTimeout(() => {
      const mockResults: FriendSearchResult[] = [
        {
          id: "search-1",
          name: "Alice Johnson",
          email: "alice.johnson@example.com",
          avatar: "/placeholder.svg",
          role: "Product Manager",
          company: "TechCorp",
          mutualFriends: 3,
          relationshipStatus: "none",
        },
        {
          id: "search-2",
          name: "Bob Smith",
          email: "bob.smith@example.com",
          role: "Developer",
          company: "StartupXYZ",
          mutualFriends: 1,
          relationshipStatus: "pending-sent",
        },
        {
          id: "search-3",
          name: "Carol Davis",
          email: "carol.davis@example.com",
          role: "Designer",
          company: "Creative Agency",
          mutualFriends: 0,
          relationshipStatus: "friends",
        },
      ].filter(
        (user) =>
          user.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
          user.email.toLowerCase().includes(searchTerm.toLowerCase()),
      )
      setSearchResults(mockResults)
      setIsSearching(false)
    }, 1000)
  }

  const handleSendInvite = (e: React.FormEvent) => {
    e.preventDefault()
    if (!inviteEmail.trim()) return

    onSendFriendRequest(inviteEmail.trim(), inviteMessage.trim() || undefined)
    setInviteEmail("")
    setInviteMessage("")
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "pending":
        return (
          <Badge variant="outline" className="text-yellow-600 border-yellow-200 bg-yellow-50">
            <Clock className="w-3 h-3 mr-1" />
            Pending
          </Badge>
        )
      case "accepted":
        return (
          <Badge variant="outline" className="text-green-600 border-green-200 bg-green-50">
            <UserCheck className="w-3 h-3 mr-1" />
            Friends
          </Badge>
        )
      default:
        return null
    }
  }

  const getRelationshipButton = (result: FriendSearchResult) => {
    switch (result.relationshipStatus) {
      case "none":
        return (
          <Button
            size="sm"
            onClick={() => onSendFriendRequest(result.email)}
            className="bg-blue-600 hover:bg-blue-700 text-white"
          >
            <UserPlus className="w-3 h-3 mr-1" />
            Add Friend
          </Button>
        )
      case "pending-sent":
        return (
          <Badge variant="outline" className="text-yellow-600 border-yellow-200 bg-yellow-50">
            <Clock className="w-3 h-3 mr-1" />
            Sent
          </Badge>
        )
      case "pending-received":
        return (
          <div className="flex space-x-1">
            <Button
              size="sm"
              onClick={() => onAcceptFriendRequest(result.id)}
              className="bg-green-600 hover:bg-green-700 text-white"
            >
              <Check className="w-3 h-3" />
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => onRejectFriendRequest(result.id)}
              className="border-gray-300 hover:bg-gray-50"
            >
              <X className="w-3 h-3" />
            </Button>
          </div>
        )
      case "friends":
        return (
          <Badge variant="outline" className="text-green-600 border-green-200 bg-green-50">
            <UserCheck className="w-3 h-3 mr-1" />
            Friends
          </Badge>
        )
      default:
        return null
    }
  }

  const pendingRequests = friendRequests.filter((req) => req.status === "pending")
  const sentRequests = friendRequests.filter((req) => req.status === "pending") // In real app, filter by current user
  const acceptedFriends = friends.filter((friend) => friend.status === "accepted")

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} />

      {/* Modal */}
      <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-4xl mx-4 max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 bg-gradient-to-r from-white to-gray-50/30">
          <div className="flex items-center">
            <Users className="w-6 h-6 mr-3 text-blue-600" />
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Friend Management</h2>
              <p className="text-sm text-gray-600">Manage your connections and network</p>
            </div>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="w-4 h-4" />
          </Button>
        </div>

        {/* Content */}
        <div className="p-6">
          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="friends" className="flex items-center">
                <Users className="w-4 h-4 mr-2" />
                Friends ({acceptedFriends.length})
              </TabsTrigger>
              <TabsTrigger value="requests" className="flex items-center">
                <UserPlus className="w-4 h-4 mr-2" />
                Requests ({pendingRequests.length})
              </TabsTrigger>
              <TabsTrigger value="search" className="flex items-center">
                <Search className="w-4 h-4 mr-2" />
                Find People
              </TabsTrigger>
              <TabsTrigger value="invite" className="flex items-center">
                <Mail className="w-4 h-4 mr-2" />
                Invite
              </TabsTrigger>
            </TabsList>

            {/* Friends List */}
            <TabsContent value="friends" className="mt-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-medium text-gray-900">Your Friends</h3>
                  <div className="text-sm text-gray-500">{acceptedFriends.length} friends</div>
                </div>
                <ScrollArea className="h-96">
                  {acceptedFriends.length === 0 ? (
                    <div className="text-center py-12">
                      <Users className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                      <p className="text-gray-500">No friends yet</p>
                      <p className="text-sm text-gray-400">Start by inviting people or searching for users</p>
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {acceptedFriends.map((friend) => (
                        <div
                          key={friend.id}
                          className="flex items-center justify-between p-4 bg-gray-50/50 rounded-lg border border-gray-200/60 hover:bg-gray-100/50 transition-colors"
                        >
                          <div className="flex items-center space-x-3">
                            <Avatar className="w-12 h-12">
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
                              {friend.role && friend.company && (
                                <div className="text-xs text-gray-400 flex items-center mt-1">
                                  <Building className="w-3 h-3 mr-1" />
                                  {friend.role} at {friend.company}
                                </div>
                              )}
                              {friend.lastActive && (
                                <div className="text-xs text-gray-400 flex items-center mt-1">
                                  <Clock className="w-3 h-3 mr-1" />
                                  Last active {friend.lastActive.toLocaleDateString()}
                                </div>
                              )}
                            </div>
                          </div>
                          <div className="flex items-center space-x-2">
                            <Button size="sm" variant="outline" className="border-gray-300 hover:bg-gray-50">
                              <MessageCircle className="w-3 h-3 mr-1" />
                              Message
                            </Button>
                            <Button size="sm" variant="outline" className="border-gray-300 hover:bg-gray-50">
                              <Calendar className="w-3 h-3 mr-1" />
                              Schedule
                            </Button>
                            <Button size="sm" variant="ghost" className="text-gray-400 hover:text-gray-600">
                              <MoreHorizontal className="w-4 h-4" />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </ScrollArea>
              </div>
            </TabsContent>

            {/* Friend Requests */}
            <TabsContent value="requests" className="mt-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-medium text-gray-900">Friend Requests</h3>
                  <div className="text-sm text-gray-500">{pendingRequests.length} pending</div>
                </div>
                <ScrollArea className="h-96">
                  {pendingRequests.length === 0 ? (
                    <div className="text-center py-12">
                      <UserPlus className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                      <p className="text-gray-500">No pending requests</p>
                      <p className="text-sm text-gray-400">Friend requests will appear here</p>
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {pendingRequests.map((request) => (
                        <div
                          key={request.id}
                          className="flex items-center justify-between p-4 bg-blue-50/50 rounded-lg border border-blue-200/60"
                        >
                          <div className="flex items-center space-x-3">
                            <Avatar className="w-12 h-12">
                              <AvatarImage src={request.fromUser.avatar || "/placeholder.svg"} />
                              <AvatarFallback className="bg-blue-100 text-blue-600 font-medium">
                                {request.fromUser.name
                                  .split(" ")
                                  .map((n) => n[0])
                                  .join("")
                                  .toUpperCase()}
                              </AvatarFallback>
                            </Avatar>
                            <div>
                              <div className="font-medium text-gray-900">{request.fromUser.name}</div>
                              <div className="text-sm text-gray-500">{request.fromUser.email}</div>
                              {request.fromUser.role && request.fromUser.company && (
                                <div className="text-xs text-gray-400 flex items-center mt-1">
                                  <Building className="w-3 h-3 mr-1" />
                                  {request.fromUser.role} at {request.fromUser.company}
                                </div>
                              )}
                              {request.message && (
                                <div className="text-sm text-gray-600 mt-2 p-2 bg-white rounded border">
                                  "{request.message}"
                                </div>
                              )}
                              <div className="text-xs text-gray-400 mt-1">
                                Sent {request.createdAt.toLocaleDateString()}
                              </div>
                            </div>
                          </div>
                          <div className="flex space-x-2">
                            <Button
                              size="sm"
                              onClick={() => onAcceptFriendRequest(request.id)}
                              className="bg-green-600 hover:bg-green-700 text-white"
                            >
                              <Check className="w-3 h-3 mr-1" />
                              Accept
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => onRejectFriendRequest(request.id)}
                              className="border-gray-300 hover:bg-gray-50"
                            >
                              <X className="w-3 h-3 mr-1" />
                              Decline
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </ScrollArea>
              </div>
            </TabsContent>

            {/* Search People */}
            <TabsContent value="search" className="mt-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-medium text-gray-900">Find People</h3>
                </div>
                <div className="flex space-x-2">
                  <div className="flex-1 relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                    <Input
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      placeholder="Search by name or email..."
                      className="pl-10 border-gray-200 focus:border-blue-300 focus:ring-blue-300"
                      onKeyPress={(e) => e.key === "Enter" && handleSearch()}
                    />
                  </div>
                  <Button onClick={handleSearch} disabled={isSearching} className="bg-blue-600 hover:bg-blue-700">
                    {isSearching ? "Searching..." : "Search"}
                  </Button>
                </div>
                <ScrollArea className="h-80">
                  {searchResults.length === 0 && !isSearching ? (
                    <div className="text-center py-12">
                      <Search className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                      <p className="text-gray-500">Search for people to connect with</p>
                      <p className="text-sm text-gray-400">Enter a name or email address to get started</p>
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {searchResults.map((result) => (
                        <div
                          key={result.id}
                          className="flex items-center justify-between p-4 bg-gray-50/50 rounded-lg border border-gray-200/60 hover:bg-gray-100/50 transition-colors"
                        >
                          <div className="flex items-center space-x-3">
                            <Avatar className="w-12 h-12">
                              <AvatarImage src={result.avatar || "/placeholder.svg"} />
                              <AvatarFallback className="bg-blue-100 text-blue-600 font-medium">
                                {result.name
                                  .split(" ")
                                  .map((n) => n[0])
                                  .join("")
                                  .toUpperCase()}
                              </AvatarFallback>
                            </Avatar>
                            <div>
                              <div className="font-medium text-gray-900">{result.name}</div>
                              <div className="text-sm text-gray-500">{result.email}</div>
                              {result.role && result.company && (
                                <div className="text-xs text-gray-400 flex items-center mt-1">
                                  <Building className="w-3 h-3 mr-1" />
                                  {result.role} at {result.company}
                                </div>
                              )}
                              {result.mutualFriends > 0 && (
                                <div className="text-xs text-blue-600 mt-1">
                                  {result.mutualFriends} mutual friend{result.mutualFriends > 1 ? "s" : ""}
                                </div>
                              )}
                            </div>
                          </div>
                          <div className="flex items-center space-x-2">{getRelationshipButton(result)}</div>
                        </div>
                      ))}
                    </div>
                  )}
                </ScrollArea>
              </div>
            </TabsContent>

            {/* Invite by Email */}
            <TabsContent value="invite" className="mt-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-medium text-gray-900">Invite Friends</h3>
                </div>
                <form onSubmit={handleSendInvite} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="inviteEmail" className="text-sm font-medium text-gray-700">
                      Email Address
                    </Label>
                    <div className="relative">
                      <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                      <Input
                        id="inviteEmail"
                        type="email"
                        value={inviteEmail}
                        onChange={(e) => setInviteEmail(e.target.value)}
                        placeholder="Enter email address..."
                        className="pl-10 border-gray-200 focus:border-blue-300 focus:ring-blue-300"
                        required
                      />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="inviteMessage" className="text-sm font-medium text-gray-700">
                      Personal Message (Optional)
                    </Label>
                    <Textarea
                      id="inviteMessage"
                      value={inviteMessage}
                      onChange={(e) => setInviteMessage(e.target.value)}
                      placeholder="Add a personal message to your invitation..."
                      className="border-gray-200 focus:border-blue-300 focus:ring-blue-300 min-h-[80px]"
                      rows={3}
                    />
                  </div>
                  <Button type="submit" className="w-full bg-blue-600 hover:bg-blue-700 text-white">
                    <Send className="w-4 h-4 mr-2" />
                    Send Invitation
                  </Button>
                </form>

                <div className="mt-8 p-4 bg-blue-50 rounded-lg border border-blue-200">
                  <h4 className="font-medium text-blue-900 mb-2">Invite Multiple People</h4>
                  <p className="text-sm text-blue-700 mb-3">
                    You can also invite multiple people at once by separating email addresses with commas.
                  </p>
                  <div className="text-xs text-blue-600">
                    Example: alice@example.com, bob@example.com, carol@example.com
                  </div>
                </div>
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </div>
    </div>
  )
}

export default FriendManagementModal
