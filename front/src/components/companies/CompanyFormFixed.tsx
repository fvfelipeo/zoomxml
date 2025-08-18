'use client';

import React, { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Company, CreateCompanyRequest, UpdateCompanyRequest, CompanyCredential, CreateCredentialRequest } from '@/types/api';
import { createCompany, updateCompany, consultarCNPJ, getCredentials, createCredential, deleteCredential } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Loader2, Plus, Trash2 } from 'lucide-react';
import { toast } from 'sonner';

const companySchema = z.object({
  cnpj: z.string().min(14, 'CNPJ deve ter 14 dígitos'),
  name: z.string().min(1, 'Nome é obrigatório'),
  trade_name: z.string().optional(),
  address: z.string().optional(),
  number: z.string().optional(),
  district: z.string().optional(),
  city: z.string().optional(),
  state: z.string().optional(),
  zip_code: z.string().optional(),
  phone: z.string().optional(),
  email: z.string().email('Email inválido').optional().or(z.literal('')),
  company_size: z.string().optional(),
  main_activity: z.string().optional(),
  secondary_activity: z.string().optional(),
  restricted: z.boolean().default(false),
  auto_fetch: z.boolean().default(true),
  active: z.boolean().default(true),
});

const credentialSchema = z.object({
  name: z.string().optional(),
  description: z.string().optional(),
  credential_type: z.string().optional(),
  username: z.string().optional(),
  password: z.string().optional(),
  token: z.string().optional(),
  environment: z.string().optional(),
});

type CompanyFormData = z.infer<typeof companySchema>;
type CredentialFormData = z.infer<typeof credentialSchema>;

interface CompanyFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  company?: Company;
  onSuccess: () => void;
}

const formatCNPJ = (value: string) => {
  return value.replace(/\D/g, '').slice(0, 14);
};

export function CompanyForm({ open, onOpenChange, company, onSuccess }: CompanyFormProps) {
  const [loading, setLoading] = useState(false);
  const [cnpjLoading, setCnpjLoading] = useState(false);
  const [credentials, setCredentials] = useState<CompanyCredential[]>([]);
  const [credentialsLoading, setCredentialsLoading] = useState(false);
  const [newCredential, setNewCredential] = useState<CredentialFormData>({});
  
  const isEditing = !!company;

  const form = useForm<CompanyFormData>({
    resolver: zodResolver(companySchema),
    defaultValues: {
      cnpj: '',
      name: '',
      trade_name: '',
      address: '',
      number: '',
      district: '',
      city: '',
      state: '',
      zip_code: '',
      phone: '',
      email: '',
      company_size: '',
      main_activity: '',
      secondary_activity: '',
      restricted: false,
      auto_fetch: true,
      active: true,
    },
  });

  // Reset form when company changes
  useEffect(() => {
    if (company) {
      form.reset({
        cnpj: company.cnpj || '',
        name: company.name || '',
        trade_name: company.trade_name || '',
        address: company.address || '',
        number: company.number || '',
        district: company.district || '',
        city: company.city || '',
        state: company.state || '',
        zip_code: company.zip_code || '',
        phone: company.phone || '',
        email: company.email || '',
        company_size: company.company_size || '',
        main_activity: company.main_activity || '',
        secondary_activity: company.secondary_activity || '',
        restricted: company.restricted || false,
        auto_fetch: company.auto_fetch !== false,
        active: company.active !== false,
      });
      
      // Load credentials if editing
      loadCredentials(company.id);
    } else {
      form.reset({
        cnpj: '',
        name: '',
        trade_name: '',
        address: '',
        number: '',
        district: '',
        city: '',
        state: '',
        zip_code: '',
        phone: '',
        email: '',
        company_size: '',
        main_activity: '',
        secondary_activity: '',
        restricted: false,
        auto_fetch: true,
        active: true,
      });
      setCredentials([]);
      setNewCredential({});
    }
  }, [company, form]);

  const loadCredentials = async (companyId: number) => {
    try {
      setCredentialsLoading(true);
      const response = await getCredentials(companyId);
      setCredentials(response.data || []);
    } catch (error) {
      console.error('Erro ao carregar credenciais:', error);
      toast.error('Erro ao carregar credenciais');
    } finally {
      setCredentialsLoading(false);
    }
  };

  const onSubmit = async (data: CompanyFormData) => {
    try {
      setLoading(true);
      
      if (isEditing && company) {
        const updateData: UpdateCompanyRequest = {
          name: data.name,
          trade_name: data.trade_name,
          address: data.address,
          number: data.number,
          district: data.district,
          city: data.city,
          state: data.state,
          zip_code: data.zip_code,
          phone: data.phone,
          email: data.email,
          company_size: data.company_size,
          main_activity: data.main_activity,
          secondary_activity: data.secondary_activity,
          restricted: data.restricted,
          auto_fetch: data.auto_fetch,
          active: data.active,
        };
        
        await updateCompany(company.id, updateData);
        toast.success('Empresa atualizada com sucesso!');
      } else {
        const createData: CreateCompanyRequest = {
          cnpj: data.cnpj,
          name: data.name,
          trade_name: data.trade_name,
          address: data.address,
          number: data.number,
          district: data.district,
          city: data.city,
          state: data.state,
          zip_code: data.zip_code,
          phone: data.phone,
          email: data.email,
          company_size: data.company_size,
          main_activity: data.main_activity,
          secondary_activity: data.secondary_activity,
          restricted: data.restricted,
          auto_fetch: data.auto_fetch,
          active: data.active,
        };
        
        const response = await createCompany(createData);
        toast.success('Empresa criada com sucesso!');
        
        // Create credential if provided
        if (newCredential.name && newCredential.credential_type && response.data?.id) {
          try {
            const credentialData: CreateCredentialRequest = {
              company_id: response.data.id,
              name: newCredential.name,
              description: newCredential.description || '',
              credential_type: newCredential.credential_type,
              username: newCredential.username || '',
              password: newCredential.password || '',
              token: newCredential.token || '',
              environment: newCredential.environment || 'production',
            };
            
            await createCredential(credentialData);
            toast.success('Credencial criada automaticamente!');
          } catch (credError) {
            console.error('Erro ao criar credencial:', credError);
            toast.error('Empresa criada, mas houve erro ao criar a credencial');
          }
        }
      }
      
      onSuccess();
      onOpenChange(false);
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || 'Erro ao salvar empresa';
      toast.error(errorMessage);
      console.error('Erro ao salvar empresa:', error);
    } finally {
      setLoading(false);
    }
  };

  const buscarDadosCNPJ = async (cnpj: string) => {
    try {
      setCnpjLoading(true);
      const response = await consultarCNPJ(cnpj);
      
      if (response.data) {
        const data = response.data;
        form.setValue('name', data.nome || '');
        form.setValue('trade_name', data.fantasia || '');
        form.setValue('address', data.logradouro || '');
        form.setValue('number', data.numero || '');
        form.setValue('district', data.bairro || '');
        form.setValue('city', data.municipio || '');
        form.setValue('state', data.uf || '');
        form.setValue('zip_code', data.cep?.replace(/\D/g, '') || '');
        form.setValue('phone', data.telefone || '');
        form.setValue('email', data.email || '');
      }

      toast.success('Dados do CNPJ carregados automaticamente!');
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || 'CNPJ não encontrado';
      toast.error(errorMessage);
      console.error('Erro ao consultar CNPJ:', error);
    } finally {
      setCnpjLoading(false);
    }
  };

  // Watch do campo CNPJ para busca automática
  const cnpjValue = form.watch('cnpj');

  useEffect(() => {
    if (cnpjValue && cnpjValue.length === 14 && !isEditing) {
      const timeoutId = setTimeout(() => {
        buscarDadosCNPJ(cnpjValue);
      }, 500); // Debounce de 500ms

      return () => clearTimeout(timeoutId);
    }
  }, [cnpjValue, isEditing]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[95vw] max-w-[1200px] h-[95vh] max-h-[900px] overflow-hidden flex flex-col p-0">
        <DialogHeader className="px-6 py-4 border-b">
          <DialogTitle className="text-xl font-semibold">
            {isEditing ? 'Editar Empresa' : 'Nova Empresa'}
          </DialogTitle>
          <DialogDescription className="text-sm text-muted-foreground">
            {isEditing 
              ? 'Edite as informações da empresa abaixo.'
              : 'Preencha as informações para criar uma nova empresa.'
            }
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto px-6">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6 py-4">
              {/* CNPJ como primeiro campo - sempre em destaque */}
              <div className="bg-blue-50 p-4 rounded-lg border border-blue-200">
                <FormField
                  control={form.control}
                  name="cnpj"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="flex items-center gap-2 text-blue-900 font-medium">
                        CNPJ *
                        {cnpjLoading && (
                          <Loader2 className="h-4 w-4 animate-spin text-blue-500" />
                        )}
                      </FormLabel>
                      <FormControl>
                        <div className="relative">
                          <Input
                            placeholder="50029654000116"
                            {...field}
                            onChange={(e) => field.onChange(formatCNPJ(e.target.value))}
                            maxLength={14}
                            disabled={cnpjLoading}
                            className="text-lg font-mono bg-white border-blue-300 focus:border-blue-500"
                          />
                          {cnpjLoading && (
                            <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                              <Loader2 className="h-4 w-4 animate-spin text-blue-500" />
                            </div>
                          )}
                        </div>
                      </FormControl>
                      <FormMessage />
                      {!isEditing && (
                        <p className="text-xs text-blue-700">
                          Digite 14 dígitos para buscar automaticamente
                        </p>
                      )}
                    </FormItem>
                  )}
                />
              </div>
