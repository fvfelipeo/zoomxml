'use client';

import { useState } from 'react';
import { Company } from '@/types/api';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { CompaniesList } from '@/components/companies/CompaniesList';
import { CompanyForm } from '@/components/companies/CompanyForm';
import { CompanyDetails } from '@/components/companies/CompanyDetails';

export default function CompaniesPage() {
  const [selectedCompany, setSelectedCompany] = useState<Company | null>(null);
  const [viewCompany, setViewCompany] = useState<Company | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleCreateCompany = () => {
    setSelectedCompany(null);
    setFormOpen(true);
  };

  const handleEditCompany = (company: Company) => {
    setSelectedCompany(company);
    setFormOpen(true);
  };

  const handleViewCompany = (company: Company) => {
    setViewCompany(company);
    setDetailsOpen(true);
  };

  const handleFormSuccess = () => {
    setRefreshKey(prev => prev + 1);
  };

  return (
    <ProtectedRoute>
      <DashboardLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Empresas</h1>
            <p className="text-muted-foreground">
              Gerencie as empresas do sistema
            </p>
          </div>

          <CompaniesList
            key={refreshKey}
            onCreateCompany={handleCreateCompany}
            onEditCompany={handleEditCompany}
            onViewCompany={handleViewCompany}
          />

          <CompanyForm
            open={formOpen}
            onOpenChange={setFormOpen}
            company={selectedCompany}
            onSuccess={handleFormSuccess}
          />

          <CompanyDetails
            open={detailsOpen}
            onOpenChange={setDetailsOpen}
            company={viewCompany}
          />
        </div>
      </DashboardLayout>
    </ProtectedRoute>
  );
}
