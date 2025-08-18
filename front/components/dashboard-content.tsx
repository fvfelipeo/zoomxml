"use client"

import { useEffect, useState } from "react"
import { 
  Building2, 
  FileText, 
  Users, 
  TrendingUp, 
  Activity, 
  Clock, 
  CheckCircle,
  AlertCircle,
  XCircle,
  Zap,
  Shield
} from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { Progress } from "@/components/ui/progress"
import { backendApiCall } from "@/lib/auth"

interface DashboardData {
  companies: {
    total: number
    active: number
    restricted: number
    auto_fetch: number
    this_week: number
  }
  documents: {
    total: number
    processed: number
    pending: number
    errors: number
    today: number
  }
  users: {
    total: number
    active: number
    admins: number
  }
  recent_activity: {
    documents_today: number
    companies_this_week: number
    last_sync_time?: string
  }
}

export function DashboardContent() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true)
        
        // Buscar estatísticas do dashboard
        const statsResponse = await backendApiCall<DashboardData>("/api/stats/dashboard")
        setData(statsResponse)
      } catch (err) {
        setError(err instanceof Error ? err.message : "Erro ao carregar dados")
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [])

  const getCompanyHealthScore = () => {
    if (!data) return 0
    const { companies } = data
    if (companies.total === 0) return 0
    
    const activePercentage = (companies.active / companies.total) * 100
    const autoFetchPercentage = (companies.auto_fetch / companies.total) * 100
    const restrictedPenalty = (companies.restricted / companies.total) * 20
    
    return Math.max(0, Math.min(100, (activePercentage + autoFetchPercentage) / 2 - restrictedPenalty))
  }

  const getDocumentProcessingRate = () => {
    if (!data || data.documents.total === 0) return 0
    return (data.documents.processed / data.documents.total) * 100
  }

  const formatLastSync = () => {
    if (!data?.recent_activity.last_sync_time) return "Nunca"
    return new Date(data.recent_activity.last_sync_time).toLocaleDateString('pt-BR', {
      day: '2-digit',
      month: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  if (loading) {
    return (
      <div className="min-h-[100vh] flex-1 rounded-xl md:min-h-min">
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <Skeleton className="h-6 w-32" />
                <Skeleton className="h-4 w-48" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-16 mb-2" />
                <Skeleton className="h-4 w-full" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-[100vh] flex-1 rounded-xl md:min-h-min">
        <Card>
          <CardContent className="pt-6 text-center">
            <AlertCircle className="h-12 w-12 text-destructive mx-auto mb-4" />
            <p className="text-destructive font-medium">Erro ao carregar dados</p>
            <p className="text-muted-foreground text-sm mt-2">{error}</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!data) return null

  const healthScore = getCompanyHealthScore()
  const processingRate = getDocumentProcessingRate()

  return (
    <div className="min-h-[100vh] flex-1 rounded-xl md:min-h-min">
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        
        {/* Saúde do Sistema */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Saúde do Sistema</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{healthScore.toFixed(0)}%</div>
            <Progress value={healthScore} className="mt-2" />
            <p className="text-xs text-muted-foreground mt-2">
              Baseado em empresas ativas e configurações
            </p>
          </CardContent>
        </Card>

        {/* Empresas Ativas */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Empresas Ativas</CardTitle>
            <Building2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data.companies.active}</div>
            <p className="text-xs text-muted-foreground">
              de {data.companies.total} empresas cadastradas
            </p>
            <div className="flex gap-2 mt-2">
              <Badge variant="secondary" className="text-xs">
                {data.companies.auto_fetch} auto-sync
              </Badge>
              {data.companies.restricted > 0 && (
                <Badge variant="outline" className="text-xs">
                  {data.companies.restricted} restritas
                </Badge>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Processamento de Documentos */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Documentos</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data.documents.total}</div>
            <p className="text-xs text-muted-foreground">
              {data.documents.today} processados hoje
            </p>
            {data.documents.total > 0 && (
              <div className="flex gap-1 mt-2">
                <div className="flex items-center gap-1">
                  <CheckCircle className="h-3 w-3 text-green-500" />
                  <span className="text-xs">{data.documents.processed}</span>
                </div>
                <div className="flex items-center gap-1">
                  <Clock className="h-3 w-3 text-yellow-500" />
                  <span className="text-xs">{data.documents.pending}</span>
                </div>
                <div className="flex items-center gap-1">
                  <XCircle className="h-3 w-3 text-red-500" />
                  <span className="text-xs">{data.documents.errors}</span>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Atividade Recente */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Atividade Recente</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Esta semana:</span>
                <span className="text-sm font-medium">{data.recent_activity.companies_this_week} empresas</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Hoje:</span>
                <span className="text-sm font-medium">{data.recent_activity.documents_today} documentos</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Última sync:</span>
                <span className="text-sm font-medium">{formatLastSync()}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Usuários do Sistema */}
        {data.users.total > 0 && (
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Usuários</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data.users.active}</div>
              <p className="text-xs text-muted-foreground">
                de {data.users.total} usuários cadastrados
              </p>
              <div className="flex gap-2 mt-2">
                <Badge variant="default" className="text-xs">
                  {data.users.admins} admins
                </Badge>
                <Badge variant="secondary" className="text-xs">
                  {data.users.total - data.users.admins} usuários
                </Badge>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Status de Segurança */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Segurança</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Autenticação:</span>
                <Badge variant="default" className="text-xs">
                  <CheckCircle className="h-3 w-3 mr-1" />
                  Ativa
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Empresas restritas:</span>
                <span className="text-sm font-medium">{data.companies.restricted}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Auto-sync:</span>
                <Badge variant={data.companies.auto_fetch > 0 ? "default" : "secondary"} className="text-xs">
                  <Zap className="h-3 w-3 mr-1" />
                  {data.companies.auto_fetch} ativas
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>

      </div>
    </div>
  )
}
