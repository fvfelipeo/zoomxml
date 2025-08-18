'use client';

import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';
import { useEffect, ReactNode } from 'react';

interface ProtectedRouteProps {
  children: ReactNode;
  requireAdmin?: boolean;
  requireAuth?: boolean;
}

export function ProtectedRoute({
  children,
  requireAdmin = false,
  requireAuth = true
}: ProtectedRouteProps) {
  const { isAuthenticated, isAdmin, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    // Só redirecionar após o carregamento inicial
    if (!isLoading) {
      if (requireAuth && !isAuthenticated) {
        router.push('/login');
        return;
      }

      if (requireAdmin && !isAdmin) {
        router.push('/unauthorized');
        return;
      }
    }
  }, [isAuthenticated, isAdmin, isLoading, requireAuth, requireAdmin, router]);

  // Mostrar loading enquanto carrega
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-4">ZoomXML</h1>
          <p className="text-muted-foreground">Carregando...</p>
        </div>
      </div>
    );
  }

  // Se requer autenticação e não está autenticado, não renderiza
  if (requireAuth && !isAuthenticated) {
    return null;
  }

  // Se requer admin e não é admin, não renderiza
  if (requireAdmin && !isAdmin) {
    return null;
  }

  return <>{children}</>;
}
