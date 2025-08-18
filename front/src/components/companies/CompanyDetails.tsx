'use client';

import { Company } from '@/types/api';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { 
  Building2, 
  MapPin, 
  Phone, 
  Mail, 
  Calendar,
  Lock,
  Unlock,
  Zap,
  ZapOff,
  CheckCircle,
  XCircle
} from 'lucide-react';
import { format } from 'date-fns';
import { ptBR } from 'date-fns/locale';

interface CompanyDetailsProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  company: Company | null;
}

export function CompanyDetails({ open, onOpenChange, company }: CompanyDetailsProps) {
  if (!company) return null;

  const formatCNPJ = (cnpj: string) => {
    return cnpj.replace(/^(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})$/, '$1.$2.$3/$4-$5');
  };

  const formatPhone = (phone: string) => {
    // Formato básico para telefone brasileiro
    const digits = phone.replace(/\D/g, '');
    if (digits.length === 11) {
      return digits.replace(/^(\d{2})(\d{5})(\d{4})$/, '($1) $2-$3');
    } else if (digits.length === 10) {
      return digits.replace(/^(\d{2})(\d{4})(\d{4})$/, '($1) $2-$3');
    }
    return phone;
  };

  const getStatusBadge = (active: boolean) => {
    return active ? (
      <Badge variant="default" className="flex items-center gap-1">
        <CheckCircle className="w-3 h-3" />
        Ativa
      </Badge>
    ) : (
      <Badge variant="destructive" className="flex items-center gap-1">
        <XCircle className="w-3 h-3" />
        Inativa
      </Badge>
    );
  };

  const getRestrictedBadge = (restricted: boolean) => {
    return restricted ? (
      <Badge variant="destructive" className="flex items-center gap-1">
        <Lock className="w-3 h-3" />
        Restrita
      </Badge>
    ) : (
      <Badge variant="secondary" className="flex items-center gap-1">
        <Unlock className="w-3 h-3" />
        Pública
      </Badge>
    );
  };

  const getAutoFetchBadge = (autoFetch: boolean) => {
    return autoFetch ? (
      <Badge variant="default" className="flex items-center gap-1">
        <Zap className="w-3 h-3" />
        Busca Automática
      </Badge>
    ) : (
      <Badge variant="outline" className="flex items-center gap-1">
        <ZapOff className="w-3 h-3" />
        Busca Manual
      </Badge>
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[700px] max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Detalhes da Empresa
          </DialogTitle>
          <DialogDescription>
            Informações completas da empresa
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Informações Básicas */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Informações Básicas</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Razão Social</p>
                  <p className="text-lg font-semibold">{company.name}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">CNPJ</p>
                  <p className="text-lg font-mono">{formatCNPJ(company.cnpj)}</p>
                </div>
              </div>

              {company.trade_name && (
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Nome Fantasia</p>
                  <p className="text-lg">{company.trade_name}</p>
                </div>
              )}

              <div className="flex flex-wrap gap-2">
                {getStatusBadge(company.active)}
                {getRestrictedBadge(company.restricted)}
                {getAutoFetchBadge(company.auto_fetch)}
              </div>
            </CardContent>
          </Card>

          {/* Endereço */}
          {(company.address || company.city || company.state) && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <MapPin className="h-4 w-4" />
                  Endereço
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {company.address && (
                  <p>
                    {company.address}
                    {company.number && `, ${company.number}`}
                    {company.complement && `, ${company.complement}`}
                  </p>
                )}
                {company.district && <p>{company.district}</p>}
                {(company.city || company.state) && (
                  <p>
                    {company.city}
                    {company.city && company.state && ' - '}
                    {company.state}
                  </p>
                )}
                {company.zip_code && (
                  <p className="text-sm text-muted-foreground">CEP: {company.zip_code}</p>
                )}
              </CardContent>
            </Card>
          )}

          {/* Contato */}
          {(company.phone || company.email) && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Contato</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {company.phone && (
                  <div className="flex items-center gap-2">
                    <Phone className="h-4 w-4 text-muted-foreground" />
                    <span>{formatPhone(company.phone)}</span>
                  </div>
                )}
                {company.email && (
                  <div className="flex items-center gap-2">
                    <Mail className="h-4 w-4 text-muted-foreground" />
                    <span>{company.email}</span>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Dados Empresariais */}
          {(company.company_size || company.main_activity || company.legal_nature) && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Dados Empresariais</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {company.company_size && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Porte da Empresa</p>
                    <p>{company.company_size}</p>
                  </div>
                )}
                {company.main_activity && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Atividade Principal</p>
                    <p>{company.main_activity}</p>
                  </div>
                )}
                {company.secondary_activity && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Atividade Secundária</p>
                    <p>{company.secondary_activity}</p>
                  </div>
                )}
                {company.legal_nature && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Natureza Jurídica</p>
                    <p>{company.legal_nature}</p>
                  </div>
                )}
                {company.opening_date && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Data de Abertura</p>
                    <p>{company.opening_date}</p>
                  </div>
                )}
                {company.registration_status && (
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Situação Cadastral</p>
                    <p>{company.registration_status}</p>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Informações do Sistema */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                Informações do Sistema
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Criada em</p>
                <p>{format(new Date(company.created_at), 'dd/MM/yyyy \'às\' HH:mm', { locale: ptBR })}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">Última atualização</p>
                <p>{format(new Date(company.updated_at), 'dd/MM/yyyy \'às\' HH:mm', { locale: ptBR })}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">ID</p>
                <p className="font-mono text-sm">{company.id}</p>
              </div>
            </CardContent>
          </Card>
        </div>
      </DialogContent>
    </Dialog>
  );
}
