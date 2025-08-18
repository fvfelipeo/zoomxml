"use client"

import { useEffect, useState } from "react"
import { 
  FileText, 
  Plus, 
  Search, 
  MoreHorizontal,
  Download,
  Eye,
  Trash2,
  CheckCircle,
  Clock,
  XCircle,
  Filter,
  Building2
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

interface Document {
  id: number
  company_id: number
  type: string
  key?: string
  number?: string
  series?: string
  issue_date?: string
  due_date?: string
  amount?: number
  status: string
  verification_code?: string
  provider_cnpj?: string
  taker_cnpj?: string
  service_value?: number
  provider_name?: string
  taker_name?: string
  is_cancelled: boolean
  is_substituted: boolean
  created_at: string
  updated_at: string
  company?: {
    id: number
    name: string
    cnpj: string
  }
}

interface DocumentsResponse {
  documents: Document[]
  pagination: {
    limit: number
    page: number
    total: number
  }
}

export function DocumentsContent() {
  const [documents, setDocuments] = useState<Document[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState("")
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  const fetchDocuments = async (page = 1, search = "") => {
    try {
      setLoading(true)
      
      const params = new URLSearchParams({
        page: page.toString(),
        limit: "10"
      })
      
      if (search) {
        params.append("search", search)
      }
      
      // Como não temos endpoint específico de documentos, vamos simular
      // TODO: Implementar endpoint real de documentos
      setDocuments([])
      setTotalPages(1)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro ao carregar documentos")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchDocuments(currentPage, searchTerm)
  }, [currentPage, searchTerm])

  const formatDate = (dateString?: string) => {
    if (!dateString) return "-"
    return new Date(dateString).toLocaleDateString('pt-BR')
  }

  const formatCurrency = (value?: number) => {
    if (!value) return "-"
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL'
    }).format(value)
  }

  const formatCNPJ = (cnpj?: string) => {
    if (!cnpj) return "-"
    return cnpj.replace(/(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})/, '$1.$2.$3/$4-$5')
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "processed":
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case "pending":
        return <Clock className="h-4 w-4 text-yellow-500" />
      case "error":
        return <XCircle className="h-4 w-4 text-red-500" />
      default:
        return <Clock className="h-4 w-4 text-gray-500" />
    }
  }

  const getStatusLabel = (status: string) => {
    switch (status) {
      case "processed":
        return "Processado"
      case "pending":
        return "Pendente"
      case "error":
        return "Erro"
      default:
        return "Desconhecido"
    }
  }

  const handleSearch = (value: string) => {
    setSearchTerm(value)
    setCurrentPage(1)
  }

  const handleView = (document: Document) => {
    // TODO: Implementar visualização do documento
    console.log("View document:", document)
  }

  const handleDownload = (document: Document) => {
    // TODO: Implementar download do documento
    console.log("Download document:", document)
  }

  const handleDelete = async (document: Document) => {
    // TODO: Implementar confirmação e exclusão
    console.log("Delete document:", document)
  }

  if (loading && documents.length === 0) {
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
                <FileText className="h-5 w-5" />
                Gerenciar Documentos
              </CardTitle>
              <CardDescription>
                {documents.length} documentos no sistema
              </CardDescription>
            </div>
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Importar Documentos
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
                  placeholder="Buscar por número, chave ou empresa..."
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

            {/* Tabela de Documentos */}
            {error ? (
              <div className="text-center py-8">
                <p className="text-destructive">Erro ao carregar documentos: {error}</p>
              </div>
            ) : documents.length === 0 ? (
              <div className="text-center py-8">
                <FileText className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p className="text-muted-foreground">Nenhum documento encontrado</p>
                <p className="text-sm text-muted-foreground mt-2">
                  Os documentos aparecerão aqui quando forem processados pelo sistema
                </p>
              </div>
            ) : (
              <div className="border rounded-lg">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Documento</TableHead>
                      <TableHead>Empresa</TableHead>
                      <TableHead>Tipo</TableHead>
                      <TableHead>Valor</TableHead>
                      <TableHead>Data Emissão</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Criado em</TableHead>
                      <TableHead className="w-[70px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {documents.map((document) => (
                      <TableRow key={document.id}>
                        <TableCell>
                          <div>
                            <div className="font-medium">
                              {document.number ? `Nº ${document.number}` : `Doc ${document.id}`}
                              {document.series && ` - Série ${document.series}`}
                            </div>
                            {document.verification_code && (
                              <div className="text-sm text-muted-foreground font-mono">
                                {document.verification_code}
                              </div>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          {document.company ? (
                            <div>
                              <div className="font-medium">{document.company.name}</div>
                              <div className="text-sm text-muted-foreground">
                                {formatCNPJ(document.company.cnpj)}
                              </div>
                            </div>
                          ) : (
                            <span className="text-muted-foreground">-</span>
                          )}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{document.type}</Badge>
                        </TableCell>
                        <TableCell>
                          {formatCurrency(document.service_value || document.amount)}
                        </TableCell>
                        <TableCell>
                          {formatDate(document.issue_date)}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {getStatusIcon(document.status)}
                            <span className="text-sm">{getStatusLabel(document.status)}</span>
                            {document.is_cancelled && (
                              <Badge variant="destructive" className="text-xs">Cancelado</Badge>
                            )}
                            {document.is_substituted && (
                              <Badge variant="secondary" className="text-xs">Substituído</Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(document.created_at)}
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
                              <DropdownMenuItem onClick={() => handleView(document)}>
                                <Eye className="h-4 w-4 mr-2" />
                                Visualizar
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => handleDownload(document)}>
                                <Download className="h-4 w-4 mr-2" />
                                Download
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem 
                                onClick={() => handleDelete(document)}
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
