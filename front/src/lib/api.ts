import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import {
  User,
  Company,
  CompanyCredential,
  Document,
  CreateUserRequest,
  UpdateUserRequest,
  CreateCompanyRequest,
  UpdateCompanyRequest,
  CreateCredentialRequest,
  UpdateCredentialRequest,
  FetchNFSeRequest,
  FetchNFSeResponse,
  PaginatedResponse,
  UserFilters,
  CompanyFilters,
  DocumentFilters,
  CNPJData
} from '@/types/api';

class ApiClient {
  private client: AxiosInstance;
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000';
    console.log('API Base URL:', this.baseURL);

    this.client = axios.create({
      baseURL: this.baseURL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Interceptor para adicionar tokens automaticamente
    this.client.interceptors.request.use((config) => {
      console.log('Making request to:', config.baseURL + config.url);

      const adminToken = localStorage.getItem('zoomxml_admin_token');
      const userToken = localStorage.getItem('zoomxml_user_token');

      // Priorizar token admin se disponível
      const token = adminToken || userToken;

      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
        console.log('Using token:', token.substring(0, 10) + '...');
      }

      return config;
    });

    // Interceptor para tratar erros de autenticação
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Token inválido, limpar localStorage
          localStorage.removeItem('zoomxml_user');
          localStorage.removeItem('zoomxml_admin_token');
          localStorage.removeItem('zoomxml_user_token');
          
          // Redirecionar para login se não estiver já lá
          if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
            window.location.href = '/login';
          }
        }
        return Promise.reject(error);
      }
    );
  }

  // Métodos para configurar tokens manualmente
  setAdminToken(token: string) {
    this.client.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  }

  setUserToken(token: string) {
    this.client.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  }

  clearTokens() {
    delete this.client.defaults.headers.common['Authorization'];
  }

  // ===== USUÁRIOS =====
  async getUsers(filters?: UserFilters): Promise<PaginatedResponse<User>> {
    const params = new URLSearchParams();
    if (filters?.role) params.append('role', filters.role);
    if (filters?.active !== undefined) params.append('active', filters.active.toString());
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.limit) params.append('limit', filters.limit.toString());

    const response = await this.client.get(`/api/users?${params.toString()}`);

    // Mapear a resposta da API para o formato esperado pelo frontend
    const apiResponse = response.data;
    return {
      data: apiResponse.users || [],
      page: apiResponse.pagination?.page || 1,
      limit: apiResponse.pagination?.limit || 20,
      total: apiResponse.pagination?.total || 0,
      total_pages: Math.ceil((apiResponse.pagination?.total || 0) / (apiResponse.pagination?.limit || 20))
    };
  }

  async getUser(id: number): Promise<User> {
    const response = await this.client.get(`/api/users/${id}`);
    return response.data;
  }

  async createUser(data: CreateUserRequest): Promise<User> {
    const response = await this.client.post('/api/users', data);
    return response.data;
  }

  async updateUser(id: number, data: UpdateUserRequest): Promise<User> {
    const response = await this.client.patch(`/api/users/${id}`, data);
    return response.data;
  }

  async deleteUser(id: number): Promise<void> {
    await this.client.delete(`/api/users/${id}`);
  }

  // ===== EMPRESAS =====
  async getCompanies(filters?: CompanyFilters): Promise<PaginatedResponse<Company>> {
    const params = new URLSearchParams();
    if (filters?.restricted !== undefined) params.append('restricted', filters.restricted.toString());
    if (filters?.active !== undefined) params.append('active', filters.active.toString());
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.limit) params.append('limit', filters.limit.toString());

    const response = await this.client.get(`/api/companies?${params.toString()}`);

    // Mapear a resposta da API para o formato esperado pelo frontend
    const apiResponse = response.data;
    return {
      data: apiResponse.companies || [],
      page: apiResponse.pagination?.page || 1,
      limit: apiResponse.pagination?.limit || 20,
      total: apiResponse.pagination?.total || 0,
      total_pages: Math.ceil((apiResponse.pagination?.total || 0) / (apiResponse.pagination?.limit || 20))
    };
  }

  async getCompany(id: number): Promise<Company> {
    const response = await this.client.get(`/api/companies/${id}`);
    return response.data;
  }

  async createCompany(data: CreateCompanyRequest): Promise<Company> {
    const response = await this.client.post('/api/companies', data);
    return response.data;
  }

  async updateCompany(id: number, data: UpdateCompanyRequest): Promise<Company> {
    const response = await this.client.patch(`/api/companies/${id}`, data);
    return response.data;
  }

  async deleteCompany(id: number): Promise<void> {
    await this.client.delete(`/api/companies/${id}`);
  }

  // ===== CREDENCIAIS =====
  async getCredentials(companyId: number): Promise<CompanyCredential[]> {
    const response = await this.client.get(`/api/companies/${companyId}/credentials`);
    return Array.isArray(response.data) ? response.data : [];
  }

  async createCredential(companyId: number, data: CreateCredentialRequest): Promise<CompanyCredential> {
    const response = await this.client.post(`/api/companies/${companyId}/credentials`, data);
    return response.data;
  }

  async updateCredential(companyId: number, credentialId: number, data: UpdateCredentialRequest): Promise<CompanyCredential> {
    const response = await this.client.patch(`/api/companies/${companyId}/credentials/${credentialId}`, data);
    return response.data;
  }

  async deleteCredential(companyId: number, credentialId: number): Promise<void> {
    await this.client.delete(`/api/companies/${companyId}/credentials/${credentialId}`);
  }

  // ===== NFSe =====
  async fetchNFSeDocuments(companyId: number, data: FetchNFSeRequest): Promise<FetchNFSeResponse> {
    const response = await this.client.post(`/api/companies/${companyId}/nfse/fetch`, data);
    return response.data;
  }

  async getNFSeDocuments(companyId: number, filters?: DocumentFilters): Promise<PaginatedResponse<Document>> {
    const params = new URLSearchParams();
    if (filters?.status) params.append('status', filters.status);
    if (filters?.start_date) params.append('start_date', filters.start_date);
    if (filters?.end_date) params.append('end_date', filters.end_date);
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.limit) params.append('limit', filters.limit.toString());

    const response = await this.client.get(`/api/companies/${companyId}/nfse?${params.toString()}`);

    // Mapear a resposta da API para o formato esperado pelo frontend
    const apiResponse = response.data;
    return {
      data: apiResponse.documents || [],
      page: apiResponse.pagination?.page || 1,
      limit: apiResponse.pagination?.limit || 20,
      total: apiResponse.pagination?.total || 0,
      total_pages: Math.ceil((apiResponse.pagination?.total || 0) / (apiResponse.pagination?.limit || 20))
    };
  }

  // ===== AUTENTICAÇÃO =====
  async validateAdminToken(token: string): Promise<boolean> {
    try {
      const response = await this.client.get('/api/users', {
        headers: { Authorization: `Bearer ${token}` }
      });
      return response.status === 200;
    } catch {
      return false;
    }
  }

  async validateUserToken(token: string): Promise<User | null> {
    try {
      const response = await this.client.get('/api/auth/me', {
        headers: { Authorization: `Bearer ${token}` }
      });
      return response.data;
    } catch {
      return null;
    }
  }

  async consultarCNPJ(cnpj: string): Promise<CNPJData> {
    const response = await this.client.get(`/api/cnpj/${cnpj}`);
    return response.data;
  }
}

// Instância singleton do cliente API
export const apiClient = new ApiClient();

// Exportar métodos individuais para facilitar o uso
export const getUsers = (filters?: UserFilters) => apiClient.getUsers(filters);
export const getUser = (id: number) => apiClient.getUser(id);
export const createUser = (data: CreateUserRequest) => apiClient.createUser(data);
export const updateUser = (id: number, data: UpdateUserRequest) => apiClient.updateUser(id, data);
export const deleteUser = (id: number) => apiClient.deleteUser(id);

export const getCompanies = (filters?: CompanyFilters) => apiClient.getCompanies(filters);
export const getCompany = (id: number) => apiClient.getCompany(id);
export const createCompany = (data: CreateCompanyRequest) => apiClient.createCompany(data);
export const updateCompany = (id: number, data: UpdateCompanyRequest) => apiClient.updateCompany(id, data);
export const deleteCompany = (id: number) => apiClient.deleteCompany(id);

export const getCredentials = (companyId: number) => apiClient.getCredentials(companyId);
export const createCredential = (companyId: number, data: CreateCredentialRequest) => apiClient.createCredential(companyId, data);
export const updateCredential = (companyId: number, credentialId: number, data: UpdateCredentialRequest) => apiClient.updateCredential(companyId, credentialId, data);
export const deleteCredential = (companyId: number, credentialId: number) => apiClient.deleteCredential(companyId, credentialId);

export const fetchNFSeDocuments = (companyId: number, data: FetchNFSeRequest) => apiClient.fetchNFSeDocuments(companyId, data);
export const getNFSeDocuments = (companyId: number, filters?: DocumentFilters) => apiClient.getNFSeDocuments(companyId, filters);

export const validateAdminToken = (token: string) => apiClient.validateAdminToken(token);
export const validateUserToken = (token: string) => apiClient.validateUserToken(token);

export const consultarCNPJ = (cnpj: string) => apiClient.consultarCNPJ(cnpj);
