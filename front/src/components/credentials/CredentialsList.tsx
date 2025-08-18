'use client';

import { useState, useEffect } from 'react';
import { CompanyCredential } from '@/types/api';
import { getCredentials, deleteCredential } from '@/lib/api';
import { Button } from '@/components/ui/button';
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
import { Plus, Edit, Trash2, Key, Shield, Globe, TestTube } from 'lucide-react';
import { toast } from 'sonner';
import { format } from 'date-fns';
import { ptBR } from 'date-fns/locale';

interface CredentialsListProps {
  companyId: number;
  companyName: string;
  onCreateCredential: () => void;
  onEditCredential: (credential: CompanyCredential) => void;
}

export function CredentialsList({ 
  companyId, 
  companyName, 
  onCreateCredential, 
  onEditCredential 
}: CredentialsListProps) {
  const [credentials, setCredentials] = useState<CompanyCredential[]>([]);
  const [loading, setLoading] = useState(true);

  const loadCredentials = async () => {
    try {
      setLoading(true);
      const data = await getCredentials(companyId);
      setCredentials(data);
    } catch (error) {
      toast.error('Erro ao carregar credenciais');
      console.error('Erro ao carregar credenciais:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCredentials();
  }, [companyId]);

  const handleDeleteCredential = async (credentialId: number) => {
    try {
      await deleteCredential(companyId, credentialId);
      toast.success('Credencial excluída com sucesso');
      loadCredentials();
    } catch (error) {
      toast.error('Erro ao excluir credencial');
      console.error('Erro ao excluir credencial:', error);
    }
  };

  const getTypeBadge = (type: string) => {
    switch (type) {
      case 'prefeitura_user_pass':
        return (
          <Badge variant="default">
            <Key className="w-3 h-3 mr-1" />
            Usuário/Senha
          </Badge>
        );
      case 'prefeitura_token':
        return (
          <Badge variant="secondary">
            <Shield className="w-3 h-3 mr-1" />
            Token
          </Badge>
        );
      case 'prefeitura_mixed':
        return (
          <Badge variant="outline">
            <Key className="w-3 h-3 mr-1" />
            Misto
          </Badge>
        );
      default:
        return <Badge variant="outline">{type}</Badge>;
    }
  };

  const getEnvironmentBadge = (environment?: string) => {
    if (!environment) return null;
    
    switch (environment) {
      case 'production':
        return (
          <Badge variant="destructive">
            <Globe className="w-3 h-3 mr-1" />
            Produção
          </Badge>
        );
      case 'staging':
        return (
          <Badge variant="default">
            <TestTube className="w-3 h-3 mr-1" />
            Homologação
          </Badge>
        );
      case 'development':
        return (
          <Badge variant="secondary">
            <TestTube className="w-3 h-3 mr-1" />
            Desenvolvimento
          </Badge>
        );
      default:
        return <Badge variant="outline">{environment}</Badge>;
    }
  };

  const getStatusBadge = (active: boolean) => {
    return active ? (
      <Badge variant="default">Ativa</Badge>
    ) : (
      <Badge variant="outline">Inativa</Badge>
    );
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Credenciais - {companyName}</CardTitle>
            <CardDescription>
              Gerencie as credenciais de acesso às APIs municipais
            </CardDescription>
          </div>
          <Button onClick={onCreateCredential}>
            <Plus className="w-4 h-4 mr-2" />
            Nova Credencial
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="text-muted-foreground">Carregando credenciais...</div>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nome</TableHead>
                <TableHead>Tipo</TableHead>
                <TableHead>Ambiente</TableHead>
                <TableHead>Login</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Criada em</TableHead>
                <TableHead className="text-right">Ações</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {credentials.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                    Nenhuma credencial encontrada
                  </TableCell>
                </TableRow>
              ) : (
                credentials.map((credential) => (
                  <TableRow key={credential.id}>
                    <TableCell>
                      <div>
                        <div className="font-medium">{credential.name}</div>
                        {credential.description && (
                          <div className="text-sm text-muted-foreground">
                            {credential.description}
                          </div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>{getTypeBadge(credential.type)}</TableCell>
                    <TableCell>{getEnvironmentBadge(credential.environment)}</TableCell>
                    <TableCell>
                      {credential.login ? (
                        <span className="font-mono text-sm">{credential.login}</span>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell>{getStatusBadge(credential.active)}</TableCell>
                    <TableCell>
                      {format(new Date(credential.created_at), 'dd/MM/yyyy', { locale: ptBR })}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end space-x-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => onEditCredential(credential)}
                        >
                          <Edit className="w-4 h-4" />
                        </Button>
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button variant="outline" size="sm">
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Confirmar exclusão</AlertDialogTitle>
                              <AlertDialogDescription>
                                Tem certeza que deseja excluir a credencial "{credential.name}"? 
                                Esta ação não pode ser desfeita.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancelar</AlertDialogCancel>
                              <AlertDialogAction
                                onClick={() => handleDeleteCredential(credential.id)}
                                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                              >
                                Excluir
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
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
