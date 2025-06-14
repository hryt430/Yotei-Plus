"use client"

import { createContext, useContext, useState, useCallback, type ReactNode } from "react"
import type { Friend, FriendRequest } from "@/types/friend"

interface FriendManagementContextType {
  friends: Friend[]
  friendRequests: FriendRequest[]
  sendFriendRequest: (email: string, message?: string) => void
  acceptFriendRequest: (requestId: string) => void
  rejectFriendRequest: (requestId: string) => void
  removeFriend: (friendId: string) => void
  blockUser: (userId: string) => void
}

const FriendManagementContext = createContext<FriendManagementContextType | undefined>(undefined)

export function useFriendManagement() {
  const context = useContext(FriendManagementContext)
  if (!context) {
    throw new Error("useFriendManagement must be used within a FriendManagementProvider")
  }
  return context
}

interface FriendManagementProviderProps {
  children: ReactNode
}

export function FriendManagementProvider({ children }: FriendManagementProviderProps) {
  const [friends, setFriends] = useState<Friend[]>([
    {
      id: "friend-1",
      name: "Sarah Johnson",
      email: "sarah.johnson@example.com",
      status: "accepted",
      role: "Product Manager",
      company: "TechCorp",
      lastActive: new Date(Date.now() - 86400000), // 1 day ago
      mutualFriends: 5,
      addedAt: new Date(Date.now() - 7 * 86400000), // 1 week ago
      acceptedAt: new Date(Date.now() - 6 * 86400000),
    },
    {
      id: "friend-2",
      name: "Mike Chen",
      email: "mike.chen@example.com",
      status: "accepted",
      role: "Software Engineer",
      company: "StartupXYZ",
      lastActive: new Date(Date.now() - 3600000), // 1 hour ago
      mutualFriends: 2,
      addedAt: new Date(Date.now() - 14 * 86400000), // 2 weeks ago
      acceptedAt: new Date(Date.now() - 13 * 86400000),
    },
    {
      id: "friend-3",
      name: "Emily Davis",
      email: "emily.davis@example.com",
      status: "accepted",
      role: "UX Designer",
      company: "Creative Agency",
      lastActive: new Date(Date.now() - 7200000), // 2 hours ago
      mutualFriends: 8,
      addedAt: new Date(Date.now() - 30 * 86400000), // 1 month ago
      acceptedAt: new Date(Date.now() - 29 * 86400000),
    },
  ])

  const [friendRequests, setFriendRequests] = useState<FriendRequest[]>([
    {
      id: "request-1",
      fromUserId: "user-123",
      toUserId: "current-user",
      fromUser: {
        id: "user-123",
        name: "Alex Rodriguez",
        email: "alex.rodriguez@example.com",
        role: "Data Scientist",
        company: "AI Corp",
      },
      toUser: {
        id: "current-user",
        name: "Current User",
        email: "current@example.com",
      },
      message: "Hi! I'd love to connect and collaborate on some projects.",
      status: "pending",
      createdAt: new Date(Date.now() - 86400000), // 1 day ago
    },
    {
      id: "request-2",
      fromUserId: "user-456",
      toUserId: "current-user",
      fromUser: {
        id: "user-456",
        name: "Lisa Wang",
        email: "lisa.wang@example.com",
        role: "Marketing Manager",
        company: "Growth Co",
      },
      toUser: {
        id: "current-user",
        name: "Current User",
        email: "current@example.com",
      },
      status: "pending",
      createdAt: new Date(Date.now() - 2 * 86400000), // 2 days ago
    },
  ])

  const sendFriendRequest = useCallback((email: string, message?: string) => {
    // In a real app, this would make an API call
    const newRequest: FriendRequest = {
      id: `request-${Date.now()}`,
      fromUserId: "current-user",
      toUserId: `user-${Date.now()}`,
      fromUser: {
        id: "current-user",
        name: "Current User",
        email: "current@example.com",
      },
      toUser: {
        id: `user-${Date.now()}`,
        name: email.split("@")[0],
        email: email,
      },
      message: message,
      status: "pending",
      createdAt: new Date(),
    }

    console.log("Friend request sent:", newRequest)
    // In real app, you might add this to a "sent requests" list
  }, [])

  const acceptFriendRequest = useCallback(
    (requestId: string) => {
      const request = friendRequests.find((req) => req.id === requestId)
      if (!request) return

      // Add to friends list
      const newFriend: Friend = {
        id: request.fromUserId,
        name: request.fromUser.name,
        email: request.fromUser.email,
        status: "accepted",
        role: request.fromUser.role,
        company: request.fromUser.company,
        mutualFriends: 0,
        addedAt: request.createdAt,
        acceptedAt: new Date(),
      }

      setFriends((prev) => [...prev, newFriend])

      // Remove from requests
      setFriendRequests((prev) => prev.filter((req) => req.id !== requestId))

      console.log("Friend request accepted:", requestId)
    },
    [friendRequests],
  )

  const rejectFriendRequest = useCallback((requestId: string) => {
    setFriendRequests((prev) => prev.filter((req) => req.id !== requestId))
    console.log("Friend request rejected:", requestId)
  }, [])

  const removeFriend = useCallback((friendId: string) => {
    setFriends((prev) => prev.filter((friend) => friend.id !== friendId))
    console.log("Friend removed:", friendId)
  }, [])

  const blockUser = useCallback((userId: string) => {
    // Remove from friends if they are friends
    setFriends((prev) => prev.filter((friend) => friend.id !== userId))
    // Remove any pending requests
    setFriendRequests((prev) => prev.filter((req) => req.fromUserId !== userId))
    console.log("User blocked:", userId)
  }, [])

  return (
    <FriendManagementContext.Provider
      value={{
        friends,
        friendRequests,
        sendFriendRequest,
        acceptFriendRequest,
        rejectFriendRequest,
        removeFriend,
        blockUser,
      }}
    >
      {children}
    </FriendManagementContext.Provider>
  )
}
