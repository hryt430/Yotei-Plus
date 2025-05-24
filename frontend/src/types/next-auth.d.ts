import NextAuth from "next-auth"

declare module "next-auth" {
  interface Session {
    accessToken?: string
    user: {
      id: string
      email: string
      name: string
      username: string
      role: string
    }
  }
  
  interface User {
    id: string
    username: string
    role: string
    accessToken?: string
  }
}