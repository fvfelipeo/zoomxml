"use client"

import { useEffect, useState } from "react"
import { 
  Key, 
  Plus, 
  Search, 
  MoreHorizontal,
  Edit,
  Trash2,
  Eye,
  EyeOff,
  Building2,
  Filter
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { backendApiCall } from "@/lib/auth"

interface Credential {
  id: number
  company_id: number
  name: string
  type: string
  username?: string
  password?: string
  certificate_path?: string
  private_key_path?: string
  environment: string
  active: boolean
  created_at: string
  updated_at: string
  company?: {
    id: number
    name: string
    cnpj: string
  }
}

interface CredentialsResponse {
  credentials: Credential[]
  pagination: {
    limit: number
    page: number
    total: number
  }
}

export function CredentialsContent() {
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState("")
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [visiblePasswords, setVisiblePasswords] = useState<Set<number>>(new Set())

  const fetchCredentials = async (page = 1, search = "") => {
    try {
      setLoading(true)
      
      // Como não temos endpoint específico de credenciais, vamos simular
      // TODO: Implementar endpoint real de credenciais
      setCredentials([])
      setTotalPages(1)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro ao carregar credenciais")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchCredentials(currentPage, searchTerm)
  }, [currentPage, searchTerm])

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('pt-BR', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const formatCNPJ = (cnpj: string) => {
    return cnpj.replace(/(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})/, '$1.$2.$3/$4-$5')
  }

  const handleSearch = (value: string) => {
    setSearchTerm(value)
    setCurrentPage(1)
  }

  const handleEdit = (credential: Credential) => {
    // TODO: Implementar modal de edição
    console.log("Edit credential:", credential)
  }

  const handleDelete = async (credential: Credential) => {
    // TODO: Implementar confirmação e exclusão
    console.log("Delete credential:", credential)
  }

  const togglePasswordVisibility = (credentialId: number) => {
    setVisiblePasswords(prev => {
      const newSet = new Set(prev)
      if (newSet.has(credentialId)) {
        newSet.delete(credentialId)
      } else {
        newSet.add(credentialId)
      }
      return newSet
    })
  }

  const maskPassword = (password?: string) => {
    if (!password) return "-"
    return "•".repeat(password.length)
  }

  const getTypeLabel = (type: string) => {
    switch (type) {
      case "certificate":
        return "Certificado"
      case "user_password":
        return "Usuário/Senha"
      case "api_key":
        return "Chave API"
      default:
        return type
    }
  }

  const getEnvironmentLabel = (environment: string) => {
    switch (environment) {
      case "production":
        return "Produção"
      case "homologation":
        return "Homologação"
      case "sandbox":
        return "Sandbox"
      default:
        return environment
    }
  }

  if (loading && credentials.length === 0) {
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
                <Key className="h-5 w-5" />
                Gerenciar Credenciais
              </CardTitle>
              <CardDescription>
                {credentials.length} credenciais cadastradas no sistema
              </CardDescription>
            </div>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Nova Credencial
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {/* Filtros e Busca */}
            <div className="flex gap-4">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Buscar por nome, empresa ou tipo..."
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

            {/* Tabela de Credenciais */}
            {error ? (
              <div className="text-center py-8">
                <p className="text-destructive">Erro ao carregar credenciais: {error}</p>
              </div>
            ) : credentials.length === 0 ? (
              <div className="text-center py-8">
                <Key className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p className="text-muted-foreground">Nenhuma credencial encontrada</p>
                <p className="text-sm text-muted-foreground mt-2">
                  As credenciais de acesso às APIs aparecerão aqui
                </p>
              </div>
            ) : (
              <div className="border rounded-lg">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Nome</TableHead>
                      <TableHead>Empresa</TableHead>
                      <TableHead>Tipo</TableHead>
                      <TableHead>Usuário</TableHead>
                      <TableHead>Senha</TableHead>
                      <TableHead>Ambiente</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Criado em</TableHead>
                      <TableHead className="w-[70px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {credentials.map((credential) => (
                      <TableRow key={credential.id}>
                        <TableCell>
                          <div className="font-medium">{credential.name}</div>
                        </TableCell>
                        <TableCell>
                          {credential.company ? (
                            <div>
                              <div className="font-medium">{credential.company.name}</div>
                              <div className="text-sm text-muted-foreground">
                                {formatCNPJ(credential.company.cnpj)}
                              </div>
                            </div>
                          ) : (
                            <span className="text-muted-foreground">-</span>
                          )}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{getTypeLabel(credential.type)}</Badge>
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          {credential.username || "-"}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <span className="font-mono text-sm">
                              {visiblePasswords.has(credential.id) 
                                ? credential.password 
                                : maskPassword(credential.password)
                              }
                            </span>
                            {credential.password && (
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 w-6 p-0"
                                onClick={() => togglePasswordVisibility(credential.id)}
                              >
                                {visiblePasswords.has(credential.id) ? (
                                  <EyeOff className="h-3 w-3" />
                                ) : (
                                  <Eye className="h-3 w-3" />
                                )}
                              </Button>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant={credential.environment === "production" ? "default" : "secondary"}>
                            {getEnvironmentLabel(credential.environment)}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={credential.active ? "default" : "secondary"}>
                            {credential.active ? "Ativa" : "Inativa"}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(credential.created_at)}
                        </TableCell>
                        <TableCell>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" className="h-8 w-8 p-0">
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuLabel>Ações</DropdownMenuLabel>
                              <DropdownMenuItem onClick={() => handleEdit(credential)}>
                                <Edit className="h-4 w-4 mr-2" />
                                Editar
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem 
                                onClick={() => handleDelete(credential)}
                                className="text-destructive"
                              >
                                <Trash2 className="h-4 w-4 mr-2" />
                                Excluir
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
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
