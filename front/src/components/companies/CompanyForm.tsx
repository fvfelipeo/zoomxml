'use client';

import { useState, useEffect } from 'react';
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
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
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
  name: z.string().min(2, 'Nome deve ter pelo menos 2 caracteres'),
  cnpj: z.string().min(14, 'CNPJ deve ter 14 dígitos').max(14, 'CNPJ deve ter 14 dígitos'),
  trade_name: z.string().optional(),
  address: z.string().optional(),
  number: z.string().optional(),
  complement: z.string().optional(),
  district: z.string().optional(),
  city: z.string().optional(),
  state: z.string().optional(),
  zip_code: z.string().optional(),
  phone: z.string().optional(),
  email: z.string().email('Email inválido').optional().or(z.literal('')),
  company_size: z.string().optional(),
  main_activity: z.string().optional(),
  secondary_activity: z.string().optional(),
  legal_nature: z.string().optional(),
  opening_date: z.string().optional(),
  registration_status: z.string().optional(),
  restricted: z.boolean(),
  auto_fetch: z.boolean(),
  active: z.boolean().optional(),
  // Campos de credencial para criação
  credential_type: z.string().optional(),
  credential_name: z.string().optional(),
  credential_description: z.string().optional(),
  credential_login: z.string().optional(),
  credential_password: z.string().optional(),
  credential_token: z.string().optional(),
  credential_environment: z.string().optional(),
});

type CompanyFormData = z.infer<typeof companySchema>;

interface CompanyFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  company?: Company | null;
  onSuccess: () => void;
}

export function CompanyForm({ open, onOpenChange, company, onSuccess }: CompanyFormProps) {
  const [loading, setLoading] = useState(false);
  const [cnpjLoading, setCnpjLoading] = useState(false);
  const [credentials, setCredentials] = useState<CompanyCredential[]>([]);
  const [credentialsLoading, setCredentialsLoading] = useState(false);
  const [newCredential, setNewCredential] = useState<Partial<CreateCredentialRequest>>({
    type: 'prefeitura_token',
    name: '',
    description: '',
    login: '',
    password: '',
    token: '',
    environment: 'production',
  });
  const isEditing = !!company;

  const form = useForm<CompanyFormData>({
    resolver: zodResolver(companySchema),
    defaultValues: {
      name: '',
      cnpj: '',
      trade_name: '',
      address: '',
      number: '',
      complement: '',
      district: '',
      city: '',
      state: '',
      zip_code: '',
      phone: '',
      email: '',
      company_size: '',
      main_activity: '',
      secondary_activity: '',
      legal_nature: '',
      opening_date: '',
      registration_status: '',
      restricted: false,
      auto_fetch: false,
      active: true,
      // Valores padrão para credencial
      credential_type: 'prefeitura_token',
      credential_name: '',
      credential_description: '',
      credential_login: '',
      credential_password: '',
      credential_token: '',
      credential_environment: 'production',
    },
  });

  // Função para carregar credenciais
  const loadCredentials = async (companyId: number) => {
    console.log('Carregando credenciais para empresa:', companyId);
    try {
      setCredentialsLoading(true);
      const credentialsList = await getCredentials(companyId);
      console.log('Credenciais carregadas:', credentialsList);
      setCredentials(credentialsList);
    } catch (error) {
      console.error('Erro ao carregar credenciais:', error);
      toast.error('Erro ao carregar credenciais');
    } finally {
      setCredentialsLoading(false);
    }
  };

  useEffect(() => {
    if (company) {
      form.reset({
        name: company.name,
        cnpj: company.cnpj,
        trade_name: company.trade_name || '',
        address: company.address || '',
        number: company.number || '',
        complement: company.complement || '',
        district: company.district || '',
        city: company.city || '',
        state: company.state || '',
        zip_code: company.zip_code || '',
        phone: company.phone || '',
        email: company.email || '',
        company_size: company.company_size || '',
        main_activity: company.main_activity || '',
        secondary_activity: company.secondary_activity || '',
        legal_nature: company.legal_nature || '',
        opening_date: company.opening_date || '',
        registration_status: company.registration_status || '',
        restricted: company.restricted,
        auto_fetch: company.auto_fetch,
        active: company.active,
      });

      // Carregar credenciais se estiver editando
      loadCredentials(company.id);
    } else {
      form.reset({
        name: '',
        cnpj: '',
        trade_name: '',
        address: '',
        number: '',
        complement: '',
        district: '',
        city: '',
        state: '',
        zip_code: '',
        phone: '',
        email: '',
        company_size: '',
        main_activity: '',
        secondary_activity: '',
        legal_nature: '',
        opening_date: '',
        registration_status: '',
        restricted: false,
        auto_fetch: false,
        active: true,
        // Resetar campos de credencial
        credential_type: 'prefeitura_token',
        credential_name: '',
        credential_description: '',
        credential_login: '',
        credential_password: '',
        credential_token: '',
        credential_environment: 'production',
      });

      // Limpar credenciais se não estiver editando
      setCredentials([]);
    }
  }, [company, form]);

  // Funções para gerenciar credenciais
  const handleCreateCredential = async () => {
    if (!company || !newCredential.type || !newCredential.name) {
      toast.error('Preencha pelo menos o tipo e nome da credencial');
      return;
    }

    try {
      setCredentialsLoading(true);
      const credentialData: CreateCredentialRequest = {
        type: newCredential.type as 'prefeitura_user_pass' | 'prefeitura_token' | 'prefeitura_mixed',
        name: newCredential.name,
        description: newCredential.description || '',
        login: newCredential.login || '',
        password: newCredential.password || '',
        token: newCredential.token || '',
        environment: newCredential.environment as 'production' | 'staging' | 'development' || 'production',
      };

      await createCredential(company.id, credentialData);
      await loadCredentials(company.id);

      // Limpar formulário
      setNewCredential({
        type: 'prefeitura_token',
        name: '',
        description: '',
        login: '',
        password: '',
        token: '',
        environment: 'production',
      });

      toast.success('Credencial criada com sucesso!');
    } catch (error) {
      console.error('Erro ao criar credencial:', error);
      toast.error('Erro ao criar credencial');
    } finally {
      setCredentialsLoading(false);
    }
  };

  const handleDeleteCredential = async (credentialId: number) => {
    if (!company) return;

    try {
      setCredentialsLoading(true);
      await deleteCredential(company.id, credentialId);
      await loadCredentials(company.id);
      toast.success('Credencial removida com sucesso!');
    } catch (error) {
      console.error('Erro ao deletar credencial:', error);
      toast.error('Erro ao deletar credencial');
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
          cnpj: data.cnpj,
          trade_name: data.trade_name || undefined,
          address: data.address || undefined,
          number: data.number || undefined,
          complement: data.complement || undefined,
          district: data.district || undefined,
          city: data.city || undefined,
          state: data.state || undefined,
          zip_code: data.zip_code || undefined,
          phone: data.phone || undefined,
          email: data.email || undefined,
          company_size: data.company_size || undefined,
          main_activity: data.main_activity || undefined,
          secondary_activity: data.secondary_activity || undefined,
          legal_nature: data.legal_nature || undefined,
          opening_date: data.opening_date || undefined,
          registration_status: data.registration_status || undefined,
          restricted: data.restricted,
          auto_fetch: data.auto_fetch,
          active: data.active,
        };

        await updateCompany(company.id, updateData);
        toast.success('Empresa atualizada com sucesso');
      } else {
        const createData: CreateCompanyRequest = {
          name: data.name,
          cnpj: data.cnpj,
          trade_name: data.trade_name || undefined,
          address: data.address || undefined,
          number: data.number || undefined,
          complement: data.complement || undefined,
          district: data.district || undefined,
          city: data.city || undefined,
          state: data.state || undefined,
          zip_code: data.zip_code || undefined,
          phone: data.phone || undefined,
          email: data.email || undefined,
          company_size: data.company_size || undefined,
          main_activity: data.main_activity || undefined,
          secondary_activity: data.secondary_activity || undefined,
          legal_nature: data.legal_nature || undefined,
          opening_date: data.opening_date || undefined,
          registration_status: data.registration_status || undefined,
          restricted: data.restricted,
          auto_fetch: data.auto_fetch,
        };

        const newCompany = await createCompany(createData);
        toast.success('Empresa criada com sucesso');

        // Criar credencial automaticamente se preenchida
        if (data.credential_name && data.credential_type && (data.credential_token || data.credential_password)) {
          try {
            const credentialData: CreateCredentialRequest = {
              type: data.credential_type as 'prefeitura_user_pass' | 'prefeitura_token' | 'prefeitura_mixed',
              name: data.credential_name,
              description: data.credential_description || '',
              login: data.credential_login || '',
              password: data.credential_password || '',
              token: data.credential_token || '',
              environment: data.credential_environment as 'production' | 'staging' | 'development' || 'production',
            };

            await createCredential(newCompany.id, credentialData);
            toast.success('Credencial criada automaticamente!');
          } catch (credentialError) {
            console.error('Erro ao criar credencial:', credentialError);
            toast.error('Empresa criada, mas houve erro ao criar a credencial');
          }
        }
      }

      onSuccess();
      onOpenChange(false);
      form.reset();
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || 'Erro ao salvar empresa';
      toast.error(errorMessage);
      console.error('Erro ao salvar empresa:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatCNPJ = (value: string) => {
    // Remove tudo que não é dígito
    const digits = value.replace(/\D/g, '');
    // Limita a 14 dígitos
    return digits.slice(0, 14);
  };

  const buscarDadosCNPJ = async (cnpj: string) => {
    if (cnpj.length !== 14 || isEditing) return;

    try {
      setCnpjLoading(true);
      const cnpjData = await consultarCNPJ(cnpj);

      // Preencher o formulário com os dados retornados
      form.setValue('name', cnpjData.name || '');
      form.setValue('trade_name', cnpjData.trade_name || '');
      form.setValue('address', cnpjData.address || '');
      form.setValue('number', cnpjData.number || '');
      form.setValue('complement', cnpjData.complement || '');
      form.setValue('district', cnpjData.district || '');
      form.setValue('city', cnpjData.city || '');
      form.setValue('state', cnpjData.state || '');
      form.setValue('zip_code', cnpjData.zip_code || '');
      form.setValue('phone', cnpjData.phone || '');
      form.setValue('email', cnpjData.email || '');
      form.setValue('company_size', cnpjData.company_size || '');
      form.setValue('main_activity', cnpjData.main_activity || '');
      form.setValue('secondary_activity', cnpjData.secondary_activities?.[0] || '');
      form.setValue('legal_nature', cnpjData.legal_nature || '');
      form.setValue('opening_date', cnpjData.opening_date || '');
      form.setValue('registration_status', cnpjData.registration_status || '');

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

              {/* Abas do formulário */}
              <Tabs defaultValue="basic" className="w-full">
                <TabsList className="grid w-full grid-cols-2 h-auto">
                  <TabsTrigger value="basic" className="text-sm py-3">
                    <span className="hidden sm:inline">Informações Básicas</span>
                    <span className="sm:hidden">Básicas</span>
                  </TabsTrigger>
                  <TabsTrigger value="credentials" className="text-sm py-3">
                    <span className="hidden sm:inline">Credenciais</span>
                    <span className="sm:hidden">Credenciais</span>
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="basic" className="space-y-6 mt-6">
                  {/* Informações Básicas */}
                  <div className="space-y-4">
                    <h3 className="text-lg font-medium text-gray-900 border-b pb-2">
                      Informações da Empresa
                    </h3>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <FormField
                        control={form.control}
                        name="name"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Razão Social *</FormLabel>
                            <FormControl>
                              <Input placeholder="Nome da empresa" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="trade_name"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Nome Fantasia</FormLabel>
                            <FormControl>
                              <Input placeholder="Nome fantasia" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </div>
                  </div>

                  {/* Endereço */}
                  <div className="space-y-4">
                    <h3 className="text-lg font-medium text-gray-900 border-b pb-2">
                      Endereço
                    </h3>

                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                      <div className="md:col-span-3">
                        <FormField
                          control={form.control}
                          name="address"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Endereço</FormLabel>
                              <FormControl>
                                <Input placeholder="Rua, avenida..." {...field} />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      </div>
                      <FormField
                        control={form.control}
                        name="number"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Número</FormLabel>
                            <FormControl>
                              <Input placeholder="123" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                      <FormField
                        control={form.control}
                        name="district"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Bairro</FormLabel>
                            <FormControl>
                              <Input placeholder="Bairro" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="city"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Cidade</FormLabel>
                            <FormControl>
                              <Input placeholder="Cidade" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="state"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Estado</FormLabel>
                            <FormControl>
                              <Input placeholder="UF" {...field} maxLength={2} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </div>
                  </div>

                  {/* Contato */}
                  <div className="space-y-4">
                    <h3 className="text-lg font-medium text-gray-900 border-b pb-2">
                      Contato
                    </h3>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <FormField
                        control={form.control}
                        name="phone"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Telefone</FormLabel>
                            <FormControl>
                              <Input placeholder="(11) 99999-9999" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="email"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Email</FormLabel>
                            <FormControl>
                              <Input type="email" placeholder="email@empresa.com" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                    </div>
                  </div>

                  {/* Configurações */}
                  <div className="space-y-4">
                    <h3 className="text-lg font-medium text-gray-900 border-b pb-2">
                      Configurações
                    </h3>

                    <div className="space-y-4">
                  <FormField
                    control={form.control}
                    name="restricted"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                        <div className="space-y-0.5">
                          <FormLabel>Empresa Restrita</FormLabel>
                          <div className="text-sm text-muted-foreground">
                            Apenas usuários específicos podem acessar esta empresa
                          </div>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="auto_fetch"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                        <div className="space-y-0.5">
                          <FormLabel>Busca Automática</FormLabel>
                          <div className="text-sm text-muted-foreground">
                            Buscar documentos NFSe automaticamente
                          </div>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  {isEditing && (
                    <FormField
                      control={form.control}
                      name="active"
                      render={({ field }) => (
                        <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                          <div className="space-y-0.5">
                            <FormLabel>Empresa Ativa</FormLabel>
                            <div className="text-sm text-muted-foreground">
                              Empresas inativas não aparecem nas listagens
                            </div>
                          </div>
                          <FormControl>
                            <Switch
                              checked={field.value}
                              onCheckedChange={field.onChange}
                            />
                          </FormControl>
                        </FormItem>
                      )}
                    />
                  )}
                    </div>
                  </div>
                </TabsContent>

              <TabsContent value="credentials" className="space-y-4 mt-6">
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="text-lg font-medium">Credenciais</h3>
                      <p className="text-sm text-muted-foreground">
                        Configure as credenciais para integração com sistemas externos.
                      </p>
                    </div>
                  </div>

                  {/* Lista de credenciais existentes */}
                  {isEditing && (
                    <div className="space-y-3">
                      <h4 className="text-sm font-medium">Credenciais Existentes</h4>
                      {credentialsLoading ? (
                        <div className="flex items-center justify-center p-4">
                          <Loader2 className="h-4 w-4 animate-spin" />
                          <span className="ml-2 text-sm">Carregando credenciais...</span>
                        </div>
                      ) : credentials.length > 0 ? (
                        <div className="space-y-2">
                          {credentials.map((credential) => (
                            <div key={credential.id} className="flex items-center justify-between p-3 border rounded-lg">
                              <div className="flex-1">
                                <div className="flex items-center gap-2">
                                  <span className="font-medium">{credential.name}</span>
                                  <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                                    {credential.type}
                                  </span>
                                  <span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
                                    {credential.environment}
                                  </span>
                                </div>
                                {credential.description && (
                                  <p className="text-sm text-muted-foreground mt-1">{credential.description}</p>
                                )}
                                {credential.login && (
                                  <p className="text-xs text-muted-foreground">Login: {credential.login}</p>
                                )}
                              </div>
                              <div className="flex items-center gap-2">
                                <Button
                                  type="button"
                                  variant="outline"
                                  size="sm"
                                  onClick={() => handleDeleteCredential(credential.id)}
                                  disabled={credentialsLoading}
                                >
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              </div>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <p className="text-sm text-muted-foreground p-4 text-center border rounded-lg">
                          Nenhuma credencial configurada
                        </p>
                      )}
                    </div>
                  )}

                  {/* Formulário para nova credencial */}
                  {isEditing && (
                    <div className="space-y-4 border-t pt-4">
                      <h4 className="text-sm font-medium">Adicionar Nova Credencial</h4>

                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <label className="text-sm font-medium">Tipo de Credencial</label>
                          <Select
                            value={newCredential.type}
                            onValueChange={(value) => setNewCredential(prev => ({ ...prev, type: value as any }))}
                          >
                            <SelectTrigger>
                              <SelectValue placeholder="Selecione o tipo" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="prefeitura_token">Token da Prefeitura</SelectItem>
                              <SelectItem value="prefeitura_user_pass">Usuário e Senha</SelectItem>
                              <SelectItem value="prefeitura_mixed">Misto (Token + Login)</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>

                        <div>
                          <label className="text-sm font-medium">Nome da Credencial</label>
                          <Input
                            placeholder="Ex: Token NFSe Imperatriz"
                            value={newCredential.name || ''}
                            onChange={(e) => setNewCredential(prev => ({ ...prev, name: e.target.value }))}
                          />
                        </div>
                      </div>

                      <div>
                        <label className="text-sm font-medium">Descrição</label>
                        <Input
                          placeholder="Descrição da credencial"
                          value={newCredential.description || ''}
                          onChange={(e) => setNewCredential(prev => ({ ...prev, description: e.target.value }))}
                        />
                      </div>

                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <label className="text-sm font-medium">Login/Usuário</label>
                          <Input
                            placeholder="Login ou usuário"
                            value={newCredential.login || ''}
                            onChange={(e) => setNewCredential(prev => ({ ...prev, login: e.target.value }))}
                          />
                        </div>

                        <div>
                          <label className="text-sm font-medium">Ambiente</label>
                          <Select
                            value={newCredential.environment}
                            onValueChange={(value) => setNewCredential(prev => ({ ...prev, environment: value as any }))}
                          >
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="production">Produção</SelectItem>
                              <SelectItem value="staging">Homologação</SelectItem>
                              <SelectItem value="development">Desenvolvimento</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>

                      {(newCredential.type === 'prefeitura_user_pass' || newCredential.type === 'prefeitura_mixed') && (
                        <div>
                          <label className="text-sm font-medium">Senha</label>
                          <Input
                            type="password"
                            placeholder="Senha"
                            value={newCredential.password || ''}
                            onChange={(e) => setNewCredential(prev => ({ ...prev, password: e.target.value }))}
                          />
                        </div>
                      )}

                      {(newCredential.type === 'prefeitura_token' || newCredential.type === 'prefeitura_mixed') && (
                        <div>
                          <label className="text-sm font-medium">Token</label>
                          <Input
                            type="password"
                            placeholder="Token de acesso"
                            value={newCredential.token || ''}
                            onChange={(e) => setNewCredential(prev => ({ ...prev, token: e.target.value }))}
                          />
                        </div>
                      )}

                      <Button
                        type="button"
                        onClick={handleCreateCredential}
                        disabled={credentialsLoading || !newCredential.name || !newCredential.type}
                        className="w-full"
                      >
                        {credentialsLoading ? (
                          <>
                            <Loader2 className="h-4 w-4 animate-spin mr-2" />
                            Salvando...
                          </>
                        ) : (
                          <>
                            <Plus className="h-4 w-4 mr-2" />
                            Adicionar Credencial
                          </>
                        )}
                      </Button>

                      <div className="bg-blue-50 p-4 rounded-lg">
                        <p className="text-sm text-blue-800">
                          💡 <strong>Dica:</strong> As credenciais são criptografadas e armazenadas com segurança.
                          Elas são necessárias para a busca automática de documentos NFSe.
                        </p>
                      </div>
                    </div>
                  )}

                  {/* Formulário para credencial na criação */}
                  {!isEditing && (
                    <div className="space-y-4">
                      <div>
                        <h4 className="text-sm font-medium mb-3">Configurar Credencial (Opcional)</h4>
                        <p className="text-sm text-muted-foreground mb-4">
                          Configure uma credencial que será criada automaticamente junto com a empresa.
                        </p>
                      </div>

                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <FormField
                          control={form.control}
                          name="credential_type"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Tipo de Credencial</FormLabel>
                              <Select onValueChange={field.onChange} defaultValue={field.value}>
                                <FormControl>
                                  <SelectTrigger>
                                    <SelectValue placeholder="Selecione o tipo" />
                                  </SelectTrigger>
                                </FormControl>
                                <SelectContent>
                                  <SelectItem value="prefeitura_token">Token da Prefeitura</SelectItem>
                                  <SelectItem value="prefeitura_user_pass">Usuário e Senha</SelectItem>
                                  <SelectItem value="prefeitura_mixed">Misto (Token + Login)</SelectItem>
                                </SelectContent>
                              </Select>
                              <FormMessage />
                            </FormItem>
                          )}
                        />

                        <FormField
                          control={form.control}
                          name="credential_name"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Nome da Credencial</FormLabel>
                              <FormControl>
                                <Input placeholder="Ex: Token NFSe Imperatriz" {...field} />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      </div>

                      <FormField
                        control={form.control}
                        name="credential_description"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Descrição</FormLabel>
                            <FormControl>
                              <Input placeholder="Descrição da credencial" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <FormField
                          control={form.control}
                          name="credential_login"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Login/Usuário</FormLabel>
                              <FormControl>
                                <Input placeholder="Login ou usuário" {...field} />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />

                        <FormField
                          control={form.control}
                          name="credential_environment"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Ambiente</FormLabel>
                              <Select onValueChange={field.onChange} defaultValue={field.value}>
                                <FormControl>
                                  <SelectTrigger>
                                    <SelectValue />
                                  </SelectTrigger>
                                </FormControl>
                                <SelectContent>
                                  <SelectItem value="production">Produção</SelectItem>
                                  <SelectItem value="staging">Homologação</SelectItem>
                                  <SelectItem value="development">Desenvolvimento</SelectItem>
                                </SelectContent>
                              </Select>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      </div>

                      {(form.watch('credential_type') === 'prefeitura_user_pass' || form.watch('credential_type') === 'prefeitura_mixed') && (
                        <FormField
                          control={form.control}
                          name="credential_password"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Senha</FormLabel>
                              <FormControl>
                                <Input type="password" placeholder="Senha" {...field} />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      )}

                      {(form.watch('credential_type') === 'prefeitura_token' || form.watch('credential_type') === 'prefeitura_mixed') && (
                        <FormField
                          control={form.control}
                          name="credential_token"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Token</FormLabel>
                              <FormControl>
                                <Input type="password" placeholder="Token de acesso" {...field} />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      )}

                      <div className="bg-blue-50 p-4 rounded-lg">
                        <p className="text-sm text-blue-800">
                          💡 <strong>Dica:</strong> Preencha os campos de credencial para que ela seja criada automaticamente junto com a empresa.
                          As credenciais são criptografadas e armazenadas com segurança.
                        </p>
                      </div>
                    </div>
                  )}
                </div>
              </TabsContent>
            </Tabs>
          </form>
        </Form>
        </div>

        <div className="border-t px-6 py-4">
          <div className="flex flex-col sm:flex-row gap-3 sm:justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={loading}
              className="w-full sm:w-auto"
            >
              Cancelar
            </Button>
            <Button
              type="submit"
              disabled={loading}
              onClick={form.handleSubmit(onSubmit)}
              className="w-full sm:w-auto"
            >
              {loading ? 'Salvando...' : (isEditing ? 'Atualizar' : 'Criar')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
