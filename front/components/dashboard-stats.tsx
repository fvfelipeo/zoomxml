"use client"

import { useEffect, useState } from "react"
import { Building2, FileText, Users, Activity } from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { backendApiCall } from "@/lib/auth"

interface DashboardStats {
  companies: {
    total: number
    active: number
    restricted: number
  }
  documents: {
    total: number
    processed: number
    pending: number
    errors: number
  }
  users: {
    total: number
    active: number
    admins: number
  }
  recentActivity: {
    documentsToday: number
    companiesThisWeek: number
  }
}

export function DashboardStats() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchStats = async () => {
      try {
        setLoading(true)

        // Buscar estatísticas do dashboard do backend
        const statsResponse = await backendApiCall<any>("/api/stats/dashboard")

        setStats({
          companies: {
            total: statsResponse.companies.total,
            active: statsResponse.companies.active,
            restricted: statsResponse.companies.restricted
          },
          documents: {
            total: statsResponse.documents.total,
            processed: statsResponse.documents.processed,
            pending: statsResponse.documents.pending,
            errors: statsResponse.documents.errors
          },
          users: {
            total: statsResponse.users.total,
            active: statsResponse.users.active,
            admins: statsResponse.users.admins
          },
          recentActivity: {
            documentsToday: statsResponse.recent_activity.documents_today,
            companiesThisWeek: statsResponse.recent_activity.companies_this_week
          }
        })
      } catch (err) {
        setError(err instanceof Error ? err.message : "Erro ao carregar estatísticas")
      } finally {
        setLoading(false)
      }
    }

    fetchStats()
  }, [])

  if (loading) {
    return (
      <div className="grid auto-rows-min gap-4 md:grid-cols-3">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <div className="h-4 bg-muted rounded w-24"></div>
              <div className="h-4 w-4 bg-muted rounded"></div>
            </CardHeader>
            <CardContent>
              <div className="h-8 bg-muted rounded w-16 mb-1"></div>
              <div className="h-3 bg-muted rounded w-32"></div>
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="grid auto-rows-min gap-4 md:grid-cols-3">
        <Card className="col-span-3">
          <CardContent className="pt-6">
            <p className="text-destructive">Erro ao carregar estatísticas: {error}</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!stats) return null

  return (
    <div className="grid auto-rows-min gap-4 md:grid-cols-3">
      {/* Estatísticas de Empresas */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Empresas</CardTitle>
          <Building2 className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.companies.total}</div>
          <p className="text-xs text-muted-foreground">
            {stats.companies.active} ativas • {stats.companies.restricted} restritas
          </p>
        </CardContent>
      </Card>

      {/* Estatísticas de Documentos */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Documentos</CardTitle>
          <FileText className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.documents.total}</div>
          <p className="text-xs text-muted-foreground">
            {stats.documents.processed} processados • {stats.documents.pending} pendentes
          </p>
        </CardContent>
      </Card>

      {/* Estatísticas de Usuários */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Usuários</CardTitle>
          <Users className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.users.total}</div>
          <p className="text-xs text-muted-foreground">
            {stats.users.active} ativos • {stats.users.admins} admins
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
