'use client';

import { useState, useEffect } from 'react';
import { Document, DocumentFilters, FetchNFSeRequest } from '@/types/api';
import { getNFSeDocuments, fetchNFSeDocuments } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { 
  Download, 
  Search, 
  FileText, 
  Calendar, 
  AlertCircle, 
  CheckCircle, 
  Clock,
  RefreshCw
} from 'lucide-react';
import { toast } from 'sonner';
import { format } from 'date-fns';
import { ptBR } from 'date-fns/locale';

interface NFSeDocumentsListProps {
  companyId: number;
  companyName: string;
}

export function NFSeDocumentsList({ companyId, companyName }: NFSeDocumentsListProps) {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [fetchLoading, setFetchLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [fetchDialogOpen, setFetchDialogOpen] = useState(false);
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [filters, setFilters] = useState<DocumentFilters>({
    page: 1,
    limit: 20,
  });

  const loadDocuments = async () => {
    try {
      setLoading(true);
      const response = await getNFSeDocuments(companyId, filters);
      setDocuments(response.data);
    } catch (error) {
      toast.error('Erro ao carregar documentos NFSe');
      console.error('Erro ao carregar documentos NFSe:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDocuments();
  }, [companyId, filters]);

  const handleFetchDocuments = async () => {
    if (!startDate || !endDate) {
      toast.error('Por favor, selecione as datas de início e fim');
      return;
    }

    try {
      setFetchLoading(true);
      const fetchRequest: FetchNFSeRequest = {
        start_date: startDate,
        end_date: endDate,
      };

      const response = await fetchNFSeDocuments(companyId, fetchRequest);
      
      if (response.success) {
        toast.success(`${response.documents_count} documentos encontrados e processados`);
        setFetchDialogOpen(false);
        loadDocuments(); // Recarregar a lista
      } else {
        toast.error(response.error || 'Erro ao buscar documentos');
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || 'Erro ao buscar documentos NFSe';
      toast.error(errorMessage);
      console.error('Erro ao buscar documentos NFSe:', error);
    } finally {
      setFetchLoading(false);
    }
  };

  const filteredDocuments = documents?.filter(doc =>
    (doc.number && doc.number.includes(searchTerm)) ||
    (doc.verification_code && doc.verification_code.includes(searchTerm)) ||
    (doc.provider_name && doc.provider_name.toLowerCase().includes(searchTerm.toLowerCase())) ||
    (doc.taker_name && doc.taker_name.toLowerCase().includes(searchTerm.toLowerCase()))
  ) || [];

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'processed':
        return (
          <Badge variant="default">
            <CheckCircle className="w-3 h-3 mr-1" />
            Processado
          </Badge>
        );
      case 'pending':
        return (
          <Badge variant="secondary">
            <Clock className="w-3 h-3 mr-1" />
            Pendente
          </Badge>
        );
      case 'error':
        return (
          <Badge variant="destructive">
            <AlertCircle className="w-3 h-3 mr-1" />
            Erro
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const formatCurrency = (value?: number) => {
    if (!value) return '-';
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value);
  };

  const formatCNPJ = (cnpj?: string) => {
    if (!cnpj) return '-';
    return cnpj.replace(/^(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})$/, '$1.$2.$3/$4-$5');
  };

  // Definir datas padrão (último mês)
  useEffect(() => {
    const today = new Date();
    const lastMonth = new Date(today.getFullYear(), today.getMonth() - 1, today.getDate());
    
    setEndDate(format(today, 'yyyy-MM-dd'));
    setStartDate(format(lastMonth, 'yyyy-MM-dd'));
  }, []);

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Documentos NFSe - {companyName}</CardTitle>
            <CardDescription>
              Visualize e gerencie os documentos NFSe da empresa
            </CardDescription>
          </div>
          <div className="flex space-x-2">
            <Button variant="outline" onClick={loadDocuments}>
              <RefreshCw className="w-4 h-4 mr-2" />
              Atualizar
            </Button>
            <Dialog open={fetchDialogOpen} onOpenChange={setFetchDialogOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Download className="w-4 h-4 mr-2" />
                  Buscar Documentos
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Buscar Documentos NFSe</DialogTitle>
                  <DialogDescription>
                    Selecione o período para buscar novos documentos NFSe
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="start-date">Data Inicial</Label>
                      <Input
                        id="start-date"
                        type="date"
                        value={startDate}
                        onChange={(e) => setStartDate(e.target.value)}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="end-date">Data Final</Label>
                      <Input
                        id="end-date"
                        type="date"
                        value={endDate}
                        onChange={(e) => setEndDate(e.target.value)}
                      />
                    </div>
                  </div>
                </div>
                <DialogFooter>
                  <Button
                    variant="outline"
                    onClick={() => setFetchDialogOpen(false)}
                    disabled={fetchLoading}
                  >
                    Cancelar
                  </Button>
                  <Button onClick={handleFetchDocuments} disabled={fetchLoading}>
                    {fetchLoading ? 'Buscando...' : 'Buscar Documentos'}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center space-x-2 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Buscar por número, código de verificação, prestador ou tomador..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-8"
            />
          </div>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="text-muted-foreground">Carregando documentos...</div>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Número</TableHead>
                <TableHead>Código Verificação</TableHead>
                <TableHead>Prestador</TableHead>
                <TableHead>Tomador</TableHead>
                <TableHead>Valor</TableHead>
                <TableHead>Data Emissão</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Ações</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredDocuments.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                    Nenhum documento encontrado
                  </TableCell>
                </TableRow>
              ) : (
                filteredDocuments.map((document) => (
                  <TableRow key={document.id}>
                    <TableCell className="font-mono text-sm">
                      {document.number || '-'}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {document.verification_code || '-'}
                    </TableCell>
                    <TableCell>
                      <div>
                        <div className="font-medium">
                          {document.provider_name || document.provider_trade_name || '-'}
                        </div>
                        {document.provider_cnpj && (
                          <div className="text-sm text-muted-foreground">
                            {formatCNPJ(document.provider_cnpj)}
                          </div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div>
                        <div className="font-medium">{document.taker_name || '-'}</div>
                        {document.taker_cnpj && (
                          <div className="text-sm text-muted-foreground">
                            {formatCNPJ(document.taker_cnpj)}
                          </div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>{formatCurrency(document.service_value)}</TableCell>
                    <TableCell>
                      {document.issue_date 
                        ? format(new Date(document.issue_date), 'dd/MM/yyyy', { locale: ptBR })
                        : '-'
                      }
                    </TableCell>
                    <TableCell>{getStatusBadge(document.status)}</TableCell>
                    <TableCell className="text-right">
                      <Button variant="outline" size="sm">
                        <FileText className="w-4 h-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
