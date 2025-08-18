'use client';

import { useState, useEffect } from 'react';
import { Company, CompanyFilters } from '@/types/api';
import { getCompanies, deleteCompany } from '@/lib/api';
import { useAuth } from '@/contexts/AuthContext';
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
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Plus, Search, Edit, Trash2, Building2, Lock, Unlock, Zap, ZapOff } from 'lucide-react';
import { toast } from 'sonner';
import { format } from 'date-fns';
import { ptBR } from 'date-fns/locale';

interface CompaniesListProps {
  onCreateCompany: () => void;
  onEditCompany: (company: Company) => void;
  onViewCompany: (company: Company) => void;
}

export function CompaniesList({ onCreateCompany, onEditCompany, onViewCompany }: CompaniesListProps) {
  const [companies, setCompanies] = useState<Company[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [filters, setFilters] = useState<CompanyFilters>({
    page: 1,
    limit: 20,
  });
  const { isAdmin } = useAuth();

  const loadCompanies = async () => {
    try {
      setLoading(true);
      const response = await getCompanies(filters);
      setCompanies(response.data);
    } catch (error) {
      toast.error('Erro ao carregar empresas');
      console.error('Erro ao carregar empresas:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCompanies();
  }, [filters]);

  const handleDeleteCompany = async (companyId: number) => {
    try {
      await deleteCompany(companyId);
      toast.success('Empresa excluída com sucesso');
      loadCompanies();
    } catch (error) {
      toast.error('Erro ao excluir empresa');
      console.error('Erro ao excluir empresa:', error);
    }
  };

  const filteredCompanies = companies?.filter(company =>
    company.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    company.cnpj.includes(searchTerm) ||
    (company.trade_name && company.trade_name.toLowerCase().includes(searchTerm.toLowerCase()))
  ) || [];

  const getRestrictedBadge = (restricted: boolean) => {
    return restricted ? (
      <Badge variant="destructive">
        <Lock className="w-3 h-3 mr-1" />
        Restrita
      </Badge>
    ) : (
      <Badge variant="secondary">
        <Unlock className="w-3 h-3 mr-1" />
        Pública
      </Badge>
    );
  };

  const getAutoFetchBadge = (autoFetch: boolean) => {
    return autoFetch ? (
      <Badge variant="default">
        <Zap className="w-3 h-3 mr-1" />
        Auto
      </Badge>
    ) : (
      <Badge variant="outline">
        <ZapOff className="w-3 h-3 mr-1" />
        Manual
      </Badge>
    );
  };

  const getStatusBadge = (active: boolean) => {
    return active ? (
      <Badge variant="default">Ativa</Badge>
    ) : (
      <Badge variant="outline">Inativa</Badge>
    );
  };

  const formatCNPJ = (cnpj: string) => {
    return cnpj.replace(/^(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})$/, '$1.$2.$3/$4-$5');
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Empresas</CardTitle>
            <CardDescription>
              Gerencie as empresas do sistema
            </CardDescription>
          </div>
          <Button onClick={onCreateCompany}>
            <Plus className="w-4 h-4 mr-2" />
            Nova Empresa
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center space-x-2 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Buscar empresas..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-8"
            />
          </div>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="text-muted-foreground">Carregando empresas...</div>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="min-w-[200px]">Empresa</TableHead>
                  <TableHead className="min-w-[140px]">CNPJ</TableHead>
                  <TableHead className="min-w-[120px] hidden sm:table-cell">Cidade</TableHead>
                  <TableHead className="min-w-[100px] hidden md:table-cell">Tipo</TableHead>
                  <TableHead className="min-w-[100px] hidden md:table-cell">Busca</TableHead>
                  <TableHead className="min-w-[80px]">Status</TableHead>
                  <TableHead className="min-w-[100px] hidden lg:table-cell">Criada em</TableHead>
                  <TableHead className="text-right min-w-[120px]">Ações</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredCompanies.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                      Nenhuma empresa encontrada
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredCompanies.map((company) => (
                    <TableRow key={company.id}>
                      <TableCell className="min-w-[200px]">
                        <div>
                          <div className="font-medium">{company.name}</div>
                          {company.trade_name && (
                            <div className="text-sm text-muted-foreground">{company.trade_name}</div>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="font-mono text-sm min-w-[140px]">
                        {formatCNPJ(company.cnpj)}
                      </TableCell>
                      <TableCell className="min-w-[120px] hidden sm:table-cell">
                        {company.city && company.state ? `${company.city}/${company.state}` : '-'}
                      </TableCell>
                      <TableCell className="min-w-[100px] hidden md:table-cell">
                        {getRestrictedBadge(company.restricted)}
                      </TableCell>
                      <TableCell className="min-w-[100px] hidden md:table-cell">
                        {getAutoFetchBadge(company.auto_fetch)}
                      </TableCell>
                      <TableCell className="min-w-[80px]">
                        {getStatusBadge(company.active)}
                      </TableCell>
                      <TableCell className="min-w-[100px] hidden lg:table-cell">
                        {format(new Date(company.created_at), 'dd/MM/yyyy', { locale: ptBR })}
                      </TableCell>
                      <TableCell className="text-right min-w-[120px]">
                        <div className="flex items-center justify-end space-x-1">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onViewCompany(company)}
                            className="h-8 w-8 p-0"
                          >
                            <Building2 className="w-4 h-4" />
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onEditCompany(company)}
                            className="h-8 w-8 p-0"
                          >
                            <Edit className="w-4 h-4" />
                          </Button>
                          {isAdmin && (
                            <AlertDialog>
                              <AlertDialogTrigger asChild>
                                <Button variant="outline" size="sm" className="h-8 w-8 p-0">
                                  <Trash2 className="w-4 h-4" />
                                </Button>
                              </AlertDialogTrigger>
                              <AlertDialogContent>
                                <AlertDialogHeader>
                                  <AlertDialogTitle>Confirmar exclusão</AlertDialogTitle>
                                  <AlertDialogDescription>
                                    Tem certeza que deseja excluir a empresa "{company.name}"?
                                    Esta ação não pode ser desfeita e todos os dados relacionados serão perdidos.
                                  </AlertDialogDescription>
                                </AlertDialogHeader>
                                <AlertDialogFooter>
                                  <AlertDialogCancel>Cancelar</AlertDialogCancel>
                                  <AlertDialogAction
                                    onClick={() => handleDeleteCompany(company.id)}
                                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                  >
                                    Excluir
                                  </AlertDialogAction>
                                </AlertDialogFooter>
                              </AlertDialogContent>
                            </AlertDialog>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
