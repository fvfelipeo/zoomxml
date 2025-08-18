import { NextRequest, NextResponse } from "next/server"

interface LoginRequest {
  email: string
  password: string
}



export async function POST(request: NextRequest) {
  try {
    const body: LoginRequest = await request.json()

    // Validate request body
    if (!body.email || !body.password) {
      return NextResponse.json(
        { error: "Email e senha são obrigatórios" },
        { status: 400 }
      )
    }

    // Get backend URL from environment or default to localhost
    const backendUrl = process.env.BACKEND_URL || "http://localhost:8000"

    // Make request to Go backend
    const response = await fetch(`${backendUrl}/api/auth/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    })

    const data = await response.json()

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status })
    }

    // Return user data from backend
    return NextResponse.json(data)

  } catch (error) {
    console.error("Login error:", error)
    return NextResponse.json(
      { error: "Erro interno do servidor" },
      { status: 500 }
    )
  }
}

// For development/testing purposes, here's what the login logic should look like:
/*
export async function POST(request: NextRequest) {
  try {
    const body: LoginRequest = await request.json()
    
    if (!body.email || !body.password) {
      return NextResponse.json(
        { error: "Email e senha são obrigatórios" },
        { status: 400 }
      )
    }

    const backendUrl = process.env.BACKEND_URL || "http://localhost:8080"
    
    // This would be the actual implementation once backend has login endpoint
    const response = await fetch(`${backendUrl}/api/auth/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    })

    if (!response.ok) {
      const errorData = await response.json()
      return NextResponse.json(errorData, { status: response.status })
    }

    const userData: User = await response.json()
    return NextResponse.json(userData)
    
  } catch (error) {
    console.error("Login error:", error)
    return NextResponse.json(
      { error: "Erro interno do servidor" },
      { status: 500 }
    )
  }
}
*/
