'use client';

import { useState } from 'react';
import { User } from '@/types/api';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { UsersList } from '@/components/users/UsersList';
import { UserForm } from '@/components/users/UserForm';

export default function UsersPage() {
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleCreateUser = () => {
    setSelectedUser(null);
    setFormOpen(true);
  };

  const handleEditUser = (user: User) => {
    setSelectedUser(user);
    setFormOpen(true);
  };

  const handleFormSuccess = () => {
    setRefreshKey(prev => prev + 1);
  };

  return (
    <ProtectedRoute requireAdmin>
      <DashboardLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Usuários</h1>
            <p className="text-muted-foreground">
              Gerencie os usuários do sistema
            </p>
          </div>

          <UsersList
            key={refreshKey}
            onCreateUser={handleCreateUser}
            onEditUser={handleEditUser}
          />

          <UserForm
            open={formOpen}
            onOpenChange={setFormOpen}
            user={selectedUser}
            onSuccess={handleFormSuccess}
          />
        </div>
      </DashboardLayout>
    </ProtectedRoute>
  );
}
