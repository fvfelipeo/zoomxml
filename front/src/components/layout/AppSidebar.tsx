'use client';

import { useAuth } from '@/contexts/AuthContext';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import {
  Building2,
  Users,
  Key,
  FileText,
  Settings,
  LogOut,
  Home,
  Shield,
  User
} from 'lucide-react';
import Link from 'next/link';

const menuItems = [
  {
    title: 'Dashboard',
    url: '/dashboard',
    icon: Home,
    requireAuth: true,
  },
  {
    title: 'Empresas',
    url: '/companies',
    icon: Building2,
    requireAuth: true,
  },
  {
    title: 'Usuários',
    url: '/users',
    icon: Users,
    requireAdmin: true,
  },
  {
    title: 'Credenciais',
    url: '/credentials',
    icon: Key,
    requireAuth: true,
  },
  {
    title: 'Documentos NFSe',
    url: '/nfse',
    icon: FileText,
    requireAuth: true,
  },
  {
    title: 'Perfil',
    url: '/profile',
    icon: User,
    requireAuth: true,
  },
  {
    title: 'Configurações',
    url: '/settings',
    icon: Settings,
    requireAuth: true,
  },
];

export function AppSidebar() {
  const { user, isAuthenticated, isAdmin, logout } = useAuth();

  const filteredMenuItems = menuItems.filter(item => {
    if (item.requireAdmin && !isAdmin) return false;
    if (item.requireAuth && !isAuthenticated) return false;
    return true;
  });

  const getUserInitials = (name: string) => {
    return name
      .split(' ')
      .map(word => word.charAt(0))
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  return (
    <Sidebar>
      <SidebarHeader className="border-b">
        <div className="flex items-center gap-2 px-4 py-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Shield className="h-4 w-4" />
          </div>
          <span className="font-semibold">ZoomXML</span>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Menu Principal</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {filteredMenuItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild>
                    <Link href={item.url}>
                      <item.icon className="h-4 w-4" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-t">
        {isAuthenticated && user && (
          <div className="flex items-center gap-3 px-4 py-3">
            <Avatar className="h-8 w-8">
              <AvatarFallback className="text-xs">
                {getUserInitials(user.name)}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">{user.name}</p>
              <p className="text-xs text-muted-foreground truncate">{user.email}</p>
              <p className="text-xs text-muted-foreground">
                {user.role === 'admin' ? 'Administrador' : 'Usuário'}
              </p>
            </div>
          </div>
        )}
        
        {isAuthenticated && (
          <div className="px-4 pb-3">
            <Button 
              variant="outline" 
              size="sm" 
              onClick={logout}
              className="w-full"
            >
              <LogOut className="h-4 w-4 mr-2" />
              Sair
            </Button>
          </div>
        )}
      </SidebarFooter>
    </Sidebar>
  );
}
