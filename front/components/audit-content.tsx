"use client"

import { useEffect, useState } from "react"
import { 
  History, 
  Search, 
  Filter,
  User,
  Building2,
  FileText,
  Users,
  Key,
  Eye,
  Plus,
  Edit,
  Trash2
} from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { backendApiCall } from "@/lib/auth"

interface AuditLog {
  id: number
  actor_id: number
  action: string
  entity: string
  entity_id?: number
  details?: any
  ip_address?: string
  user_agent?: string
  created_at: string
  actor?: {
    id: number
    name: string
    email: string
    role: string
  }
}

interface AuditLogsResponse {
  audit_logs: AuditLog[]
  pagination: {
    limit: number
    page: number
    total: number
  }
}

export function AuditContent() {
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState("")
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  const fetchAuditLogs = async (page = 1, search = "") => {
    try {
      setLoading(true)
      
      // Como não temos endpoint específico de auditoria, vamos simular
      // TODO: Implementar endpoint real de auditoria
      setAuditLogs([])
      setTotalPages(1)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro ao carregar logs de auditoria")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchAuditLogs(currentPage, searchTerm)
  }, [currentPage, searchTerm])

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('pt-BR', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  }

  const handleSearch = (value: string) => {
    setSearchTerm(value)
    setCurrentPage(1)
  }

  const getActionLabel = (action: string) => {
    switch (action) {
      case "CREATE":
        return "Criação"
      case "UPDATE":
        return "Atualização"
      case "DELETE":
        return "Exclusão"
      case "LOGIN":
        return "Login"
      case "LOGOUT":
        return "Logout"
      case "VIEW":
        return "Visualização"
      default:
        return action
    }
  }

  const getActionIcon = (action: string) => {
    switch (action) {
      case "CREATE":
        return <Plus className="h-4 w-4 text-green-500" />
      case "UPDATE":
        return <Edit className="h-4 w-4 text-blue-500" />
      case "DELETE":
        return <Trash2 className="h-4 w-4 text-red-500" />
      case "LOGIN":
      case "LOGOUT":
        return <User className="h-4 w-4 text-purple-500" />
      case "VIEW":
        return <Eye className="h-4 w-4 text-gray-500" />
      default:
        return <History className="h-4 w-4 text-gray-500" />
    }
  }

  const getEntityIcon = (entity: string) => {
    switch (entity) {
      case "User":
        return <Users className="h-4 w-4" />
      case "Company":
        return <Building2 className="h-4 w-4" />
      case "Document":
        return <FileText className="h-4 w-4" />
      case "Credential":
        return <Key className="h-4 w-4" />
      default:
        return <History className="h-4 w-4" />
    }
  }

  const getEntityLabel = (entity: string) => {
    switch (entity) {
      case "User":
        return "Usuário"
      case "Company":
        return "Empresa"
      case "Document":
        return "Documento"
      case "Credential":
        return "Credencial"
      case "CompanyMember":
        return "Membro"
      default:
        return entity
    }
  }

  if (loading && auditLogs.length === 0) {
    return (
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-64" />
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex gap-4">
                <Skeleton className="h-10 flex-1" />
                <Skeleton className="h-10 w-32" />
              </div>
              <div className="space-y-2">
                {[1, 2, 3, 4, 5].map((i) => (
                  <Skeleton key={i} className="h-16 w-full" />
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <History className="h-5 w-5" />
                Logs de Auditoria
              </CardTitle>
              <CardDescription>
                {auditLogs.length} registros de atividades do sistema
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {/* Filtros e Busca */}
            <div className="flex gap-4">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Buscar por usuário, ação ou entidade..."
                  value={searchTerm}
                  onChange={(e) => handleSearch(e.target.value)}
                  className="pl-10"
                />
              </div>
              <Button variant="outline">
                <Filter className="h-4 w-4 mr-2" />
                Filtros
              </Button>
            </div>

            {/* Tabela de Logs */}
            {error ? (
              <div className="text-center py-8">
                <p className="text-destructive">Erro ao carregar logs: {error}</p>
              </div>
            ) : auditLogs.length === 0 ? (
              <div className="text-center py-8">
                <History className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p className="text-muted-foreground">Nenhum log de auditoria encontrado</p>
                <p className="text-sm text-muted-foreground mt-2">
                  As atividades do sistema aparecerão aqui conforme forem executadas
                </p>
              </div>
            ) : (
              <div className="border rounded-lg">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Usuário</TableHead>
                      <TableHead>Ação</TableHead>
                      <TableHead>Entidade</TableHead>
                      <TableHead>Detalhes</TableHead>
                      <TableHead>IP</TableHead>
                      <TableHead>Data/Hora</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {auditLogs.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell>
                          {log.actor ? (
                            <div>
                              <div className="font-medium">{log.actor.name}</div>
                              <div className="text-sm text-muted-foreground">
                                {log.actor.email}
                              </div>
                            </div>
                          ) : (
                            <span className="text-muted-foreground">Sistema</span>
                          )}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {getActionIcon(log.action)}
                            <Badge variant="outline">{getActionLabel(log.action)}</Badge>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {getEntityIcon(log.entity)}
                            <span>{getEntityLabel(log.entity)}</span>
                            {log.entity_id && (
                              <span className="text-sm text-muted-foreground">
                                #{log.entity_id}
                              </span>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          {log.details ? (
                            <div className="max-w-xs">
                              <pre className="text-xs text-muted-foreground whitespace-pre-wrap">
                                {typeof log.details === 'string' 
                                  ? log.details 
                                  : JSON.stringify(log.details, null, 2).substring(0, 100) + '...'
                                }
                              </pre>
                            </div>
                          ) : (
                            <span className="text-muted-foreground">-</span>
                          )}
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          {log.ip_address || "-"}
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(log.created_at)}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}

            {/* Paginação */}
            {totalPages > 1 && (
              <div className="flex items-center justify-between">
                <p className="text-sm text-muted-foreground">
                  Página {currentPage} de {totalPages}
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                    disabled={currentPage === 1}
                  >
                    Anterior
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                    disabled={currentPage === totalPages}
                  >
                    Próxima
                  </Button>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
