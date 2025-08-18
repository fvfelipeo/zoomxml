export interface User {
  id: number
  name: string
  email: string
  role: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface LoginResponse {
  id: number
  name: string
  email: string
  role: string
  active: boolean
  token: string
  created_at: string
  updated_at: string
}

export interface ApiError {
  error: string
  details?: any
}

// Token management
export const AUTH_TOKEN_KEY = "auth_token"
export const USER_DATA_KEY = "user_data"

export function getAuthToken(): string | null {
  if (typeof window === "undefined") return null
  return localStorage.getItem(AUTH_TOKEN_KEY)
}

export function setAuthToken(token: string): void {
  if (typeof window === "undefined") return
  localStorage.setItem(AUTH_TOKEN_KEY, token)
}

export function removeAuthToken(): void {
  if (typeof window === "undefined") return
  localStorage.removeItem(AUTH_TOKEN_KEY)
  localStorage.removeItem(USER_DATA_KEY)
}

export function getCurrentUser(): User | null {
  if (typeof window === "undefined") return null
  const userData = localStorage.getItem(USER_DATA_KEY)
  if (!userData) return null
  
  try {
    return JSON.parse(userData) as User
  } catch {
    return null
  }
}

export function setCurrentUser(user: User): void {
  if (typeof window === "undefined") return
  localStorage.setItem(USER_DATA_KEY, JSON.stringify(user))
}

export function isAuthenticated(): boolean {
  return getAuthToken() !== null && getCurrentUser() !== null
}

export function isAdmin(): boolean {
  const user = getCurrentUser()
  return user?.role === "admin"
}

// API call helper with authentication
export async function apiCall<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getAuthToken()
  const baseUrl = process.env.NEXT_PUBLIC_API_URL || "/api"
  
  const config: RequestInit = {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    },
  }

  const response = await fetch(`${baseUrl}${endpoint}`, config)
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: "Erro desconhecido" }))
    throw new Error(errorData.error || `HTTP ${response.status}`)
  }

  return response.json()
}

// Authentication API calls
export async function login(credentials: LoginCredentials): Promise<LoginResponse> {
  const response = await fetch("/api/auth/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(credentials),
  })

  if (!response.ok) {
    const errorData = await response.json()
    throw new Error(errorData.error || "Erro ao fazer login")
  }

  return response.json()
}

export async function logout(): Promise<void> {
  // Clear local storage
  removeAuthToken()
  
  // Optionally call backend logout endpoint if it exists
  try {
    await fetch("/api/auth/logout", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${getAuthToken()}`,
      },
    })
  } catch {
    // Ignore logout errors - token is already cleared locally
  }
}

// Backend API calls (direct to Go backend)
export async function backendApiCall<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getAuthToken()
  const backendUrl = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080"
  
  const config: RequestInit = {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    },
  }

  const response = await fetch(`${backendUrl}${endpoint}`, config)
  
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: "Erro desconhecido" }))
    throw new Error(errorData.error || `HTTP ${response.status}`)
  }

  return response.json()
}

// Redirect helpers
export function redirectToLogin(): void {
  if (typeof window !== "undefined") {
    window.location.href = "/login"
  }
}

export function redirectToDashboard(): void {
  if (typeof window !== "undefined") {
    window.location.href = "/dashboard"
  }
}
