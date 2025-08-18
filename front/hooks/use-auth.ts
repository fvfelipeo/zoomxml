"use client"

import { useState, useEffect, useCallback } from "react"
import { useRouter } from "next/navigation"
import { 
  getCurrentUser, 
  getAuthToken, 
  isAuthenticated, 
  isAdmin, 
  logout as authLogout,
  type User 
} from "@/lib/auth"

interface UseAuthReturn {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isAdmin: boolean
  isLoading: boolean
  logout: () => Promise<void>
  refreshAuth: () => void
}

export function useAuth(): UseAuthReturn {
  const router = useRouter()
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const refreshAuth = useCallback(() => {
    const currentUser = getCurrentUser()
    const currentToken = getAuthToken()
    
    setUser(currentUser)
    setToken(currentToken)
    setIsLoading(false)
  }, [])

  const logout = useCallback(async () => {
    try {
      await authLogout()
      setUser(null)
      setToken(null)
      router.push("/login")
    } catch (error) {
      console.error("Logout error:", error)
      // Force logout even if API call fails
      setUser(null)
      setToken(null)
      router.push("/login")
    }
  }, [router])

  useEffect(() => {
    refreshAuth()
  }, [refreshAuth])

  // Listen for storage changes (for multi-tab sync)
  useEffect(() => {
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === "auth_token" || e.key === "user_data") {
        refreshAuth()
      }
    }

    window.addEventListener("storage", handleStorageChange)
    return () => window.removeEventListener("storage", handleStorageChange)
  }, [refreshAuth])

  return {
    user,
    token,
    isAuthenticated: isAuthenticated(),
    isAdmin: isAdmin(),
    isLoading,
    logout,
    refreshAuth
  }
}

// Hook for protecting routes that require authentication
export function useRequireAuth(redirectTo: string = "/login") {
  const { isAuthenticated, isLoading } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push(redirectTo)
    }
  }, [isAuthenticated, isLoading, router, redirectTo])

  return { isAuthenticated, isLoading }
}

// Hook for protecting admin-only routes
export function useRequireAdmin(redirectTo: string = "/dashboard") {
  const { isAdmin, isAuthenticated, isLoading } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        router.push("/login")
      } else if (!isAdmin) {
        router.push(redirectTo)
      }
    }
  }, [isAdmin, isAuthenticated, isLoading, router, redirectTo])

  return { isAdmin, isAuthenticated, isLoading }
}
