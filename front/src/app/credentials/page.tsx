'use client';

import { useState, useEffect } from 'react';
import { Company, CompanyCredential } from '@/types/api';
import { getCompanies } from '@/lib/api';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { CredentialsList } from '@/components/credentials/CredentialsList';
import { CredentialForm } from '@/components/credentials/CredentialForm';
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

export default function CredentialsPage() {
  const [companies, setCompanies] = useState<Company[]>([]);
  const [selectedCompanyId, setSelectedCompanyId] = useState<number | null>(null);
  const [selectedCredential, setSelectedCredential] = useState<CompanyCredential | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [refreshKey, setRefreshKey] = useState(0);

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

  const handleCreateCredential = () => {
    setSelectedCredential(null);
    setFormOpen(true);
  };

  const handleEditCredential = (credential: CompanyCredential) => {
    setSelectedCredential(credential);
    setFormOpen(true);
  };

  const handleFormSuccess = () => {
    setRefreshKey(prev => prev + 1);
  };

  return (
    <ProtectedRoute>
      <DashboardLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Credenciais</h1>
            <p className="text-muted-foreground">
              Gerencie as credenciais de acesso às APIs municipais
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
                Escolha uma empresa para gerenciar suas credenciais
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

          {/* Lista de Credenciais */}
          {selectedCompany && (
            <CredentialsList
              key={refreshKey}
              companyId={selectedCompany.id}
              companyName={selectedCompany.name}
              onCreateCredential={handleCreateCredential}
              onEditCredential={handleEditCredential}
            />
          )}

          {!selectedCompany && !loading && (
            <Card>
              <CardContent className="flex items-center justify-center py-12">
                <div className="text-center">
                  <Building2 className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-lg font-semibold mb-2">Nenhuma empresa selecionada</h3>
                  <p className="text-muted-foreground">
                    Selecione uma empresa acima para gerenciar suas credenciais
                  </p>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Formulário de Credencial */}
          {selectedCompanyId && (
            <CredentialForm
              open={formOpen}
              onOpenChange={setFormOpen}
              companyId={selectedCompanyId}
              credential={selectedCredential}
              onSuccess={handleFormSuccess}
            />
          )}
        </div>
      </DashboardLayout>
    </ProtectedRoute>
  );
}
