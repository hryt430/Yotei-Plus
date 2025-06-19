"use client"

import { createContext, useContext, useState, useCallback, type ReactNode } from "react"
import type { Friendship, User } from "@/types"

interface FriendManagementContextType {
  friendships: Friendship[]
  friends: Friendship[]
  friendRequests: Friendship[] 
  sentRequests: Friendship[]
  currentUserId: string
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
  currentUserId?: string
}

export function FriendManagementProvider({ children, currentUserId = "current-user" }: FriendManagementProviderProps) {
  // Mock user data
  const mockUsers: Record<string, User> = {
    "current-user": {
      id: "current-user",
      username: "Current User",
      email: "current@example.com",
      role: "user",
    },
    "user-1": {
      id: "user-1",
      username: "Sarah Johnson",
      email: "sarah.johnson@example.com",
      role: "user",
    },
    "user-2": {
      id: "user-2",
      username: "Mike Chen",
      email: "mike.chen@example.com",
      role: "user",
    },
    "user-3": {
      id: "user-3",
      username: "Emily Davis",
      email: "emily.davis@example.com",
      role: "user",
    },
    "user-4": {
      id: "user-4",
      username: "Alex Rodriguez",
      email: "alex.rodriguez@example.com",
      role: "user",
    },
    "user-5": {
      id: "user-5",
      username: "Lisa Wang",
      email: "lisa.wang@example.com",
      role: "user",
    },
  }

  // 全ての友達関係を一つの状態で管理（既存型システム準拠）
  const [friendships, setFriendships] = useState<Friendship[]>([
    // 受け入れ済みの友達関係
    {
      id: "friendship-1",
      requester_id: currentUserId,
      addressee_id: "user-1",
      status: "ACCEPTED",
      created_at: new Date(Date.now() - 7 * 86400000).toISOString(),
      updated_at: new Date(Date.now() - 6 * 86400000).toISOString(),
      accepted_at: new Date(Date.now() - 6 * 86400000).toISOString(),
      requester: mockUsers[currentUserId],
      addressee: mockUsers["user-1"],
    },
    {
      id: "friendship-2",
      requester_id: currentUserId,
      addressee_id: "user-2",
      status: "ACCEPTED",
      created_at: new Date(Date.now() - 14 * 86400000).toISOString(),
      updated_at: new Date(Date.now() - 13 * 86400000).toISOString(),
      accepted_at: new Date(Date.now() - 13 * 86400000).toISOString(),
      requester: mockUsers[currentUserId],
      addressee: mockUsers["user-2"],
    },
    {
      id: "friendship-3",
      requester_id: "user-3",
      addressee_id: currentUserId,
      status: "ACCEPTED",
      created_at: new Date(Date.now() - 30 * 86400000).toISOString(),
      updated_at: new Date(Date.now() - 29 * 86400000).toISOString(),
      accepted_at: new Date(Date.now() - 29 * 86400000).toISOString(),
      requester: mockUsers["user-3"],
      addressee: mockUsers[currentUserId],
    },
    // 受信した友達リクエスト（PENDING状態、自分がaddressee）
    {
      id: "request-1",
      requester_id: "user-4",
      addressee_id: currentUserId,
      status: "PENDING",
      created_at: new Date(Date.now() - 86400000).toISOString(), // 1 day ago
      updated_at: new Date(Date.now() - 86400000).toISOString(),
      requester: mockUsers["user-4"],
      addressee: mockUsers[currentUserId],
    },
    {
      id: "request-2",
      requester_id: "user-5",
      addressee_id: currentUserId,
      status: "PENDING",
      created_at: new Date(Date.now() - 2 * 86400000).toISOString(), // 2 days ago
      updated_at: new Date(Date.now() - 2 * 86400000).toISOString(),
      requester: mockUsers["user-5"],
      addressee: mockUsers[currentUserId],
    },
    // 送信した友達リクエスト（PENDING状態、自分がrequester）
    {
      id: "sent-request-1",
      requester_id: currentUserId,
      addressee_id: "user-6",
      status: "PENDING",
      created_at: new Date(Date.now() - 3 * 86400000).toISOString(), // 3 days ago
      updated_at: new Date(Date.now() - 3 * 86400000).toISOString(),
      requester: mockUsers[currentUserId],
      addressee: {
        id: "user-6",
        username: "John Smith",
        email: "john.smith@example.com",
        role: "user",
      },
    },
  ])

  // フィルタリング関数（既存型システム準拠）
  const friends = friendships.filter(f => f.status === "ACCEPTED")
  
  const friendRequests = friendships.filter(f => 
    f.status === "PENDING" && f.addressee_id === currentUserId
  )
  
  const sentRequests = friendships.filter(f => 
    f.status === "PENDING" && f.requester_id === currentUserId
  )

  const sendFriendRequest = useCallback((email: string, message?: string) => {
    // In a real app, this would make an API call
    const targetUserId = `user-${Date.now()}`
    const targetUser: User = {
      id: targetUserId,
      username: email.split("@")[0],
      email: email,
      role: "user",
    }

    const newFriendship: Friendship = {
      id: `friendship-${Date.now()}`,
      requester_id: currentUserId,
      addressee_id: targetUserId,
      status: "PENDING",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      requester: mockUsers[currentUserId],
      addressee: targetUser,
    }

    setFriendships(prev => [...prev, newFriendship])
    console.log("Friend request sent:", newFriendship)
  }, [currentUserId])

  const acceptFriendRequest = useCallback((requestId: string) => {
    setFriendships(prev => 
      prev.map(friendship => 
        friendship.id === requestId
          ? {
              ...friendship,
              status: "ACCEPTED" as const,
              updated_at: new Date().toISOString(),
              accepted_at: new Date().toISOString(),
            }
          : friendship
      )
    )
    console.log("Friend request accepted:", requestId)
  }, [])

  const rejectFriendRequest = useCallback((requestId: string) => {
    setFriendships(prev => prev.filter(f => f.id !== requestId))
    console.log("Friend request rejected:", requestId)
  }, [])

  const removeFriend = useCallback((friendId: string) => {
    setFriendships(prev => prev.filter(f => f.id !== friendId))
    console.log("Friend removed:", friendId)
  }, [])

  const blockUser = useCallback((userId: string) => {
    // Remove all friendships with this user and set status to BLOCKED
    setFriendships(prev => 
      prev.map(friendship => {
        if (friendship.requester_id === userId || friendship.addressee_id === userId) {
          return {
            ...friendship,
            status: "BLOCKED" as const,
            updated_at: new Date().toISOString(),
            blocked_at: new Date().toISOString(),
          }
        }
        return friendship
      })
    )
    console.log("User blocked:", userId)
  }, [])

  return (
    <FriendManagementContext.Provider
      value={{
        friendships,
        friends,
        friendRequests,
        sentRequests,
        currentUserId,
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