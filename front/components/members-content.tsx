"use client"

import { useEffect, useState } from "react"
import { 
  UserCheck, 
  Plus, 
  Search, 
  MoreHorizontal,
  Edit,
  Trash2,
  Building2,
  User,
  Shield,
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

interface Member {
  id: number
  company_id: number
  user_id: number
  role: string
  permissions: string[]
  active: boolean
  created_at: string
  updated_at: string
  user?: {
    id: number
    name: string
    email: string
    role: string
    active: boolean
  }
  company?: {
    id: number
    name: string
    cnpj: string
  }
}

interface MembersResponse {
  members: Member[]
  pagination: {
    limit: number
    page: number
    total: number
  }
}

export function MembersContent() {
  const [members, setMembers] = useState<Member[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState("")
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  const fetchMembers = async (page = 1, search = "") => {
    try {
      setLoading(true)
      
      // Como não temos endpoint específico de membros, vamos simular
      // TODO: Implementar endpoint real de membros
      setMembers([])
      setTotalPages(1)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro ao carregar membros")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchMembers(currentPage, searchTerm)
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

  const handleEdit = (member: Member) => {
    // TODO: Implementar modal de edição
    console.log("Edit member:", member)
  }

  const handleDelete = async (member: Member) => {
    // TODO: Implementar confirmação e exclusão
    console.log("Delete member:", member)
  }

  const getRoleLabel = (role: string) => {
    switch (role) {
      case "admin":
        return "Administrador"
      case "manager":
        return "Gerente"
      case "viewer":
        return "Visualizador"
      case "editor":
        return "Editor"
      default:
        return role
    }
  }

  const getRoleIcon = (role: string) => {
    switch (role) {
      case "admin":
        return <Shield className="h-4 w-4 text-red-500" />
      case "manager":
        return <UserCheck className="h-4 w-4 text-blue-500" />
      default:
        return <User className="h-4 w-4 text-gray-500" />
    }
  }

  if (loading && members.length === 0) {
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
                <UserCheck className="h-5 w-5" />
                Gerenciar Membros
              </CardTitle>
              <CardDescription>
                {members.length} membros de empresas no sistema
              </CardDescription>
            </div>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Adicionar Membro
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
                  placeholder="Buscar por nome, email ou empresa..."
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

            {/* Tabela de Membros */}
            {error ? (
              <div className="text-center py-8">
                <p className="text-destructive">Erro ao carregar membros: {error}</p>
              </div>
            ) : members.length === 0 ? (
              <div className="text-center py-8">
                <UserCheck className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p className="text-muted-foreground">Nenhum membro encontrado</p>
                <p className="text-sm text-muted-foreground mt-2">
                  Os membros das empresas aparecerão aqui quando forem adicionados
                </p>
              </div>
            ) : (
              <div className="border rounded-lg">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Usuário</TableHead>
                      <TableHead>Empresa</TableHead>
                      <TableHead>Role na Empresa</TableHead>
                      <TableHead>Permissões</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Adicionado em</TableHead>
                      <TableHead className="w-[70px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {members.map((member) => (
                      <TableRow key={member.id}>
                        <TableCell>
                          {member.user ? (
                            <div className="flex items-center gap-3">
                              <div className="bg-primary/10 p-2 rounded-full">
                                {member.user.role === "admin" ? (
                                  <Shield className="h-4 w-4 text-primary" />
                                ) : (
                                  <User className="h-4 w-4 text-primary" />
                                )}
                              </div>
                              <div>
                                <div className="font-medium">{member.user.name}</div>
                                <div className="text-sm text-muted-foreground">
                                  {member.user.email}
                                </div>
                              </div>
                            </div>
                          ) : (
                            <span className="text-muted-foreground">Usuário não encontrado</span>
                          )}
                        </TableCell>
                        <TableCell>
                          {member.company ? (
                            <div>
                              <div className="font-medium">{member.company.name}</div>
                              <div className="text-sm text-muted-foreground">
                                {formatCNPJ(member.company.cnpj)}
                              </div>
                            </div>
                          ) : (
                            <span className="text-muted-foreground">-</span>
                          )}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {getRoleIcon(member.role)}
                            <Badge variant="outline">{getRoleLabel(member.role)}</Badge>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-1 flex-wrap">
                            {member.permissions.length > 0 ? (
                              member.permissions.slice(0, 2).map((permission) => (
                                <Badge key={permission} variant="secondary" className="text-xs">
                                  {permission}
                                </Badge>
                              ))
                            ) : (
                              <span className="text-muted-foreground text-sm">Nenhuma</span>
                            )}
                            {member.permissions.length > 2 && (
                              <Badge variant="secondary" className="text-xs">
                                +{member.permissions.length - 2}
                              </Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant={member.active ? "default" : "secondary"}>
                            {member.active ? "Ativo" : "Inativo"}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(member.created_at)}
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
                              <DropdownMenuItem onClick={() => handleEdit(member)}>
                                <Edit className="h-4 w-4 mr-2" />
                                Editar
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem 
                                onClick={() => handleDelete(member)}
                                className="text-destructive"
                              >
                                <Trash2 className="h-4 w-4 mr-2" />
                                Remover
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
