'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { CompanyCredential, CreateCredentialRequest, UpdateCredentialRequest } from '@/types/api';
import { createCredential, updateCredential } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { toast } from 'sonner';

const credentialSchema = z.object({
  type: z.enum(['prefeitura_user_pass', 'prefeitura_token', 'prefeitura_mixed']),
  name: z.string().min(2, 'Nome deve ter pelo menos 2 caracteres'),
  description: z.string().optional(),
  login: z.string().optional(),
  password: z.string().optional(),
  token: z.string().optional(),
  environment: z.enum(['production', 'staging', 'development']).optional(),
  active: z.boolean().optional(),
});

type CredentialFormData = z.infer<typeof credentialSchema>;

interface CredentialFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  companyId: number;
  credential?: CompanyCredential | null;
  onSuccess: () => void;
}

export function CredentialForm({ 
  open, 
  onOpenChange, 
  companyId, 
  credential, 
  onSuccess 
}: CredentialFormProps) {
  const [loading, setLoading] = useState(false);
  const isEditing = !!credential;

  const form = useForm<CredentialFormData>({
    resolver: zodResolver(credentialSchema),
    defaultValues: {
      type: 'prefeitura_user_pass',
      name: '',
      description: '',
      login: '',
      password: '',
      token: '',
      environment: 'production',
      active: true,
    },
  });

  const selectedType = form.watch('type');

  useEffect(() => {
    if (credential) {
      form.reset({
        type: credential.type,
        name: credential.name,
        description: credential.description || '',
        login: credential.login || '',
        password: '', // Não preencher senha ao editar
        token: '', // Não preencher token ao editar
        environment: credential.environment || 'production',
        active: credential.active,
      });
    } else {
      form.reset({
        type: 'prefeitura_user_pass',
        name: '',
        description: '',
        login: '',
        password: '',
        token: '',
        environment: 'production',
        active: true,
      });
    }
  }, [credential, form]);

  const onSubmit = async (data: CredentialFormData) => {
    try {
      setLoading(true);

      if (isEditing && credential) {
        const updateData: UpdateCredentialRequest = {
          name: data.name,
          description: data.description || undefined,
          login: data.login || undefined,
          environment: data.environment,
          active: data.active,
        };

        // Só incluir senha se foi fornecida
        if (data.password) {
          updateData.password = data.password;
        }

        // Só incluir token se foi fornecido
        if (data.token) {
          updateData.token = data.token;
        }

        await updateCredential(companyId, credential.id, updateData);
        toast.success('Credencial atualizada com sucesso');
      } else {
        const createData: CreateCredentialRequest = {
          type: data.type,
          name: data.name,
          description: data.description || undefined,
          login: data.login || undefined,
          password: data.password || undefined,
          token: data.token || undefined,
          environment: data.environment,
        };

        await createCredential(companyId, createData);
        toast.success('Credencial criada com sucesso');
      }

      onSuccess();
      onOpenChange(false);
      form.reset();
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || 'Erro ao salvar credencial';
      toast.error(errorMessage);
      console.error('Erro ao salvar credencial:', error);
    } finally {
      setLoading(false);
    }
  };

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'prefeitura_user_pass':
        return 'Usuário e Senha';
      case 'prefeitura_token':
        return 'Token';
      case 'prefeitura_mixed':
        return 'Misto (Usuário/Senha + Token)';
      default:
        return type;
    }
  };

  const showLoginPassword = selectedType === 'prefeitura_user_pass' || selectedType === 'prefeitura_mixed';
  const showToken = selectedType === 'prefeitura_token' || selectedType === 'prefeitura_mixed';

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {isEditing ? 'Editar Credencial' : 'Nova Credencial'}
          </DialogTitle>
          <DialogDescription>
            {isEditing 
              ? 'Edite as informações da credencial abaixo.'
              : 'Preencha as informações para criar uma nova credencial.'
            }
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Nome *</FormLabel>
                  <FormControl>
                    <Input placeholder="Nome da credencial" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Descrição</FormLabel>
                  <FormControl>
                    <Textarea 
                      placeholder="Descrição opcional da credencial"
                      className="resize-none"
                      {...field} 
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Tipo *</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Selecione o tipo" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="prefeitura_user_pass">
                        {getTypeLabel('prefeitura_user_pass')}
                      </SelectItem>
                      <SelectItem value="prefeitura_token">
                        {getTypeLabel('prefeitura_token')}
                      </SelectItem>
                      <SelectItem value="prefeitura_mixed">
                        {getTypeLabel('prefeitura_mixed')}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            {showLoginPassword && (
              <>
                <FormField
                  control={form.control}
                  name="login"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Login</FormLabel>
                      <FormControl>
                        <Input placeholder="Nome de usuário" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {isEditing ? 'Nova Senha (opcional)' : 'Senha'}
                      </FormLabel>
                      <FormControl>
                        <Input 
                          type="password" 
                          placeholder={isEditing ? 'Deixe em branco para manter a atual' : 'Senha'} 
                          {...field} 
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </>
            )}

            {showToken && (
              <FormField
                control={form.control}
                name="token"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {isEditing ? 'Novo Token (opcional)' : 'Token'}
                    </FormLabel>
                    <FormControl>
                      <Input 
                        placeholder={isEditing ? 'Deixe em branco para manter o atual' : 'Token de acesso'} 
                        {...field} 
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}

            <FormField
              control={form.control}
              name="environment"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Ambiente</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Selecione o ambiente" />
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

            {isEditing && (
              <FormField
                control={form.control}
                name="active"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                    <div className="space-y-0.5">
                      <FormLabel>Credencial Ativa</FormLabel>
                      <div className="text-sm text-muted-foreground">
                        Credenciais inativas não são utilizadas
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

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={loading}
              >
                Cancelar
              </Button>
              <Button type="submit" disabled={loading}>
                {loading ? 'Salvando...' : (isEditing ? 'Atualizar' : 'Criar')}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
