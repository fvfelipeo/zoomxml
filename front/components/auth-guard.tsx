"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { isAuthenticated, getCurrentUser, type User } from "@/lib/auth"
import { Skeleton } from "@/components/ui/skeleton"

interface AuthGuardProps {
  children: React.ReactNode
  requireAuth?: boolean
  requireAdmin?: boolean
  redirectTo?: string
}

export function AuthGuard({ 
  children, 
  requireAuth = true, 
  requireAdmin = false,
  redirectTo = "/login" 
}: AuthGuardProps) {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(true)
  const [user, setUser] = useState<User | null>(null)

  useEffect(() => {
    const checkAuth = () => {
      const authenticated = isAuthenticated()
      const currentUser = getCurrentUser()
      
      setUser(currentUser)
      
      if (requireAuth && !authenticated) {
        router.push(redirectTo)
        return
      }
      
      if (requireAdmin && (!currentUser || currentUser.role !== "admin")) {
        router.push("/dashboard") // Redirect non-admin users to dashboard
        return
      }
      
      setIsLoading(false)
    }

    checkAuth()
  }, [router, requireAuth, requireAdmin, redirectTo])

  if (isLoading) {
    return (
      <div className="min-h-screen flex flex-col">
        <header className="flex h-16 shrink-0 items-center gap-2 border-b bg-background px-4">
          <Skeleton className="h-6 w-24" />
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4">
          <div className="grid auto-rows-min gap-4 md:grid-cols-3">
            <Skeleton className="aspect-video rounded-xl" />
            <Skeleton className="aspect-video rounded-xl" />
            <Skeleton className="aspect-video rounded-xl" />
          </div>
          <Skeleton className="min-h-[100vh] flex-1 rounded-xl md:min-h-min" />
        </div>
      </div>
    )
  }

  return <>{children}</>
}

// Hook for getting current user with loading state
export function useCurrentUser() {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const currentUser = getCurrentUser()
    setUser(currentUser)
    setIsLoading(false)
  }, [])

  return { user, isLoading }
}
