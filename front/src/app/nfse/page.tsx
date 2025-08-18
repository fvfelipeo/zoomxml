'use client';

import { useState, useEffect } from 'react';
import { Company } from '@/types/api';
import { getCompanies } from '@/lib/api';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { NFSeDocumentsList } from '@/components/nfse/NFSeDocumentsList';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Building2 } from 'lucide-react';
import { toast } from 'sonner';

export default function NFSePage() {
  const [companies, setCompanies] = useState<Company[]>([]);
  const [selectedCompanyId, setSelectedCompanyId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);

  const loadCompanies = async () => {
    try {
      setLoading(true);
      const response = await getCompanies({ limit: 100 });
      setCompanies(response.data);
      
      // Selecionar a primeira empresa automaticamente se houver apenas uma
      if (response.data.length === 1) {
        setSelectedCompanyId(response.data[0].id);
      }
    } catch (error) {
      toast.error('Erro ao carregar empresas');
      console.error('Erro ao carregar empresas:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCompanies();
  }, []);

  const selectedCompany = companies.find(c => c.id === selectedCompanyId);

  return (
    <ProtectedRoute>
      <DashboardLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Documentos NFSe</h1>
            <p className="text-muted-foreground">
              Visualize e gerencie os documentos NFSe das empresas
            </p>
          </div>

          {/* Seletor de Empresa */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Building2 className="h-5 w-5" />
                Selecionar Empresa
              </CardTitle>
              <CardDescription>
                Escolha uma empresa para visualizar seus documentos NFSe
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="text-muted-foreground">Carregando empresas...</div>
              ) : (
                <Select
                  value={selectedCompanyId?.toString() || ''}
                  onValueChange={(value) => setSelectedCompanyId(parseInt(value))}
                >
                  <SelectTrigger className="w-full max-w-md">
                    <SelectValue placeholder="Selecione uma empresa" />
                  </SelectTrigger>
                  <SelectContent>
                    {companies.map((company) => (
                      <SelectItem key={company.id} value={company.id.toString()}>
                        <div className="flex flex-col">
                          <span className="font-medium">{company.name}</span>
                          {company.trade_name && (
                            <span className="text-sm text-muted-foreground">
                              {company.trade_name}
                            </span>
                          )}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </CardContent>
          </Card>

          {/* Lista de Documentos NFSe */}
          {selectedCompany && (
            <NFSeDocumentsList
              companyId={selectedCompany.id}
              companyName={selectedCompany.name}
            />
          )}

          {!selectedCompany && !loading && (
            <Card>
              <CardContent className="flex items-center justify-center py-12">
                <div className="text-center">
                  <Building2 className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-lg font-semibold mb-2">Nenhuma empresa selecionada</h3>
                  <p className="text-muted-foreground">
                    Selecione uma empresa acima para visualizar seus documentos NFSe
                  </p>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </DashboardLayout>
    </ProtectedRoute>
  );
}
