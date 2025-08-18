'use client';

import { useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { 
  User, 
  Mail, 
  Shield, 
  Key, 
  Settings,
  LogOut,
  Save,
  Eye,
  EyeOff
} from 'lucide-react';
import { toast } from 'sonner';
import { format } from 'date-fns';
import { ptBR } from 'date-fns/locale';

export default function ProfilePage() {
  const { user, logout } = useAuth();
  const [showToken, setShowToken] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({
    name: user?.name || '',
    email: user?.email || '',
  });

  const handleSave = async () => {
    try {
      // TODO: Implementar atualização do perfil
      toast.success('Perfil atualizado com sucesso');
      setIsEditing(false);
    } catch (error) {
      toast.error('Erro ao atualizar perfil');
    }
  };

  const handleLogout = () => {
    logout();
    toast.success('Logout realizado com sucesso');
  };

  const copyToken = () => {
    if (user?.token) {
      navigator.clipboard.writeText(user.token);
      toast.success('Token copiado para a área de transferência');
    }
  };

  return (
    <ProtectedRoute>
      <DashboardLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Perfil</h1>
            <p className="text-muted-foreground">
              Gerencie suas informações pessoais e configurações
            </p>
          </div>

          {/* Informações Básicas */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <User className="h-5 w-5" />
                Informações Pessoais
              </CardTitle>
              <CardDescription>
                Suas informações básicas no sistema
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Nome</Label>
                  {isEditing ? (
                    <Input
                      id="name"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    />
                  ) : (
                    <p className="text-lg font-medium">{user?.name}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  {isEditing ? (
                    <Input
                      id="email"
                      type="email"
                      value={formData.email}
                      onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    />
                  ) : (
                    <p className="text-lg font-medium">{user?.email}</p>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-4">
                <div>
                  <Label>Papel no Sistema</Label>
                  <div className="mt-1">
                    <Badge variant={user?.role === 'admin' ? 'destructive' : 'secondary'}>
                      {user?.role === 'admin' ? 'Administrador' : 'Usuário'}
                    </Badge>
                  </div>
                </div>

                <div>
                  <Label>Status</Label>
                  <div className="mt-1">
                    <Badge variant={user?.active ? 'default' : 'outline'}>
                      {user?.active ? 'Ativo' : 'Inativo'}
                    </Badge>
                  </div>
                </div>
              </div>

              <div className="flex gap-2">
                {isEditing ? (
                  <>
                    <Button onClick={handleSave} className="flex items-center gap-2">
                      <Save className="h-4 w-4" />
                      Salvar
                    </Button>
                    <Button 
                      variant="outline" 
                      onClick={() => {
                        setIsEditing(false);
                        setFormData({
                          name: user?.name || '',
                          email: user?.email || '',
                        });
                      }}
                    >
                      Cancelar
                    </Button>
                  </>
                ) : (
                  <Button onClick={() => setIsEditing(true)} className="flex items-center gap-2">
                    <Settings className="h-4 w-4" />
                    Editar Perfil
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Token de Acesso */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Key className="h-5 w-5" />
                Token de Acesso
              </CardTitle>
              <CardDescription>
                Seu token pessoal para acesso à API
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>Token</Label>
                <div className="flex items-center gap-2">
                  <Input
                    type={showToken ? 'text' : 'password'}
                    value={user?.token || ''}
                    readOnly
                    className="font-mono"
                  />
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setShowToken(!showToken)}
                  >
                    {showToken ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </Button>
                  <Button
                    variant="outline"
                    onClick={copyToken}
                  >
                    Copiar
                  </Button>
                </div>
              </div>

              <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                <div className="flex items-start gap-2">
                  <Shield className="h-5 w-5 text-yellow-600 mt-0.5" />
                  <div>
                    <h4 className="font-medium text-yellow-800">Importante</h4>
                    <p className="text-sm text-yellow-700">
                      Mantenha seu token seguro. Não compartilhe com terceiros e use-o apenas 
                      em aplicações confiáveis.
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Informações da Conta */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Mail className="h-5 w-5" />
                Informações da Conta
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <Label>Conta criada em</Label>
                  <p className="text-lg font-medium">
                    {user?.created_at 
                      ? format(new Date(user.created_at), 'dd/MM/yyyy \'às\' HH:mm', { locale: ptBR })
                      : 'Não informado'
                    }
                  </p>
                </div>

                <div>
                  <Label>Última atualização</Label>
                  <p className="text-lg font-medium">
                    {user?.updated_at 
                      ? format(new Date(user.updated_at), 'dd/MM/yyyy \'às\' HH:mm', { locale: ptBR })
                      : 'Não informado'
                    }
                  </p>
                </div>
              </div>

              <div>
                <Label>ID do Usuário</Label>
                <p className="text-lg font-medium font-mono">{user?.id}</p>
              </div>
            </CardContent>
          </Card>

          {/* Ações da Conta */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings className="h-5 w-5" />
                Ações da Conta
              </CardTitle>
              <CardDescription>
                Ações relacionadas à sua conta
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <Separator />
                
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="font-medium">Sair da Conta</h4>
                    <p className="text-sm text-muted-foreground">
                      Encerrar sua sessão atual no sistema
                    </p>
                  </div>
                  <Button 
                    variant="outline" 
                    onClick={handleLogout}
                    className="flex items-center gap-2"
                  >
                    <LogOut className="h-4 w-4" />
                    Sair
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </DashboardLayout>
    </ProtectedRoute>
  );
}
