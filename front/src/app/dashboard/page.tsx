'use client';

import { useAuth } from '@/contexts/AuthContext';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Building2, Users, Key, FileText, Shield, Activity } from 'lucide-react';

export default function DashboardPage() {
  const { user, isAdmin } = useAuth();

  const stats = [
    {
      title: 'Empresas',
      value: '12',
      description: 'Empresas cadastradas',
      icon: Building2,
      color: 'text-blue-600',
    },
    {
      title: 'Usuários',
      value: '8',
      description: 'Usuários ativos',
      icon: Users,
      color: 'text-green-600',
      adminOnly: true,
    },
    {
      title: 'Credenciais',
      value: '24',
      description: 'Credenciais configuradas',
      icon: Key,
      color: 'text-yellow-600',
    },
    {
      title: 'Documentos NFSe',
      value: '1,234',
      description: 'Documentos processados',
      icon: FileText,
      color: 'text-purple-600',
    },
  ];

  const filteredStats = stats.filter(stat => !stat.adminOnly || isAdmin);

  return (
    <ProtectedRoute>
      <DashboardLayout>
        <div className="space-y-6">
          {/* Header */}
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
            <p className="text-muted-foreground">
              Bem-vindo ao ZoomXML, {user?.name}
            </p>
          </div>

          {/* User Info */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Shield className="h-5 w-5" />
                Informações do Usuário
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Nome</p>
                  <p className="text-lg font-semibold">{user?.name}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Email</p>
                  <p className="text-lg font-semibold">{user?.email}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Papel</p>
                  <div className="flex items-center gap-2">
                    <Badge variant={user?.role === 'admin' ? 'destructive' : 'secondary'}>
                      {user?.role === 'admin' ? 'Administrador' : 'Usuário'}
                    </Badge>
                    <Badge variant={user?.active ? 'default' : 'outline'}>
                      {user?.active ? 'Ativo' : 'Inativo'}
                    </Badge>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Stats Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {filteredStats.map((stat) => (
              <Card key={stat.title}>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">
                    {stat.title}
                  </CardTitle>
                  <stat.icon className={`h-4 w-4 ${stat.color}`} />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{stat.value}</div>
                  <p className="text-xs text-muted-foreground">
                    {stat.description}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Recent Activity */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5" />
                Atividade Recente
              </CardTitle>
              <CardDescription>
                Últimas ações realizadas no sistema
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center gap-4 p-3 rounded-lg border">
                  <div className="h-2 w-2 bg-green-500 rounded-full"></div>
                  <div className="flex-1">
                    <p className="text-sm font-medium">Documentos NFSe processados</p>
                    <p className="text-xs text-muted-foreground">Há 2 horas</p>
                  </div>
                </div>
                <div className="flex items-center gap-4 p-3 rounded-lg border">
                  <div className="h-2 w-2 bg-blue-500 rounded-full"></div>
                  <div className="flex-1">
                    <p className="text-sm font-medium">Nova empresa cadastrada</p>
                    <p className="text-xs text-muted-foreground">Há 4 horas</p>
                  </div>
                </div>
                <div className="flex items-center gap-4 p-3 rounded-lg border">
                  <div className="h-2 w-2 bg-yellow-500 rounded-full"></div>
                  <div className="flex-1">
                    <p className="text-sm font-medium">Credenciais atualizadas</p>
                    <p className="text-xs text-muted-foreground">Ontem</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </DashboardLayout>
    </ProtectedRoute>
  );
}
