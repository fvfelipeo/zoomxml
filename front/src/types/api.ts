// Tipos baseados nos modelos da API ZoomXML

export interface User {
  id: number;
  name: string;
  email: string;
  role: 'admin' | 'user';
  active: boolean;
  created_at: string;
  updated_at: string;
  company_members?: CompanyMember[];
}

export interface Company {
  id: number;
  name: string;
  cnpj: string;
  trade_name?: string;
  
  // Endereço
  address?: string;
  number?: string;
  complement?: string;
  district?: string;
  city?: string;
  state?: string;
  zip_code?: string;
  
  // Contato
  phone?: string;
  email?: string;
  
  // Dados empresariais
  company_size?: string;
  main_activity?: string;
  secondary_activity?: string;
  legal_nature?: string;
  opening_date?: string;
  registration_status?: string;
  
  // Configurações
  restricted: boolean;
  auto_fetch: boolean;
  active: boolean;
  created_at: string;
  updated_at: string;
  
  // Relacionamentos
  members?: CompanyMember[];
  credentials?: CompanyCredential[];
  documents?: Document[];
}

export interface CompanyMember {
  id: number;
  user_id: number;
  company_id: number;
  created_at: string;
  updated_at: string;
  user?: User;
  company?: Company;
}

export interface CompanyCredential {
  id: number;
  company_id: number;
  type: 'prefeitura_user_pass' | 'prefeitura_token' | 'prefeitura_mixed';
  name: string;
  description?: string;
  login?: string;
  environment?: 'production' | 'staging' | 'development';
  active: boolean;
  created_at: string;
  updated_at: string;
  company?: Company;
}

export interface Document {
  id: number;
  company_id: number;
  type: string;
  key?: string;
  number?: string;
  series?: string;
  issue_date?: string;
  due_date?: string;
  amount?: number;
  status: 'pending' | 'processed' | 'error';
  storage_key?: string;
  hash?: string;
  metadata?: string;
  
  // Campos específicos NFSe
  verification_code?: string;
  provider_cnpj?: string;
  taker_cnpj?: string;
  service_value?: number;
  service_code?: string;
  municipal_registration?: string;
  document_hash?: string;
  is_cancelled?: boolean;
  is_substituted?: boolean;
  processing_date?: string;
  competence?: string;
  rps_issue_date?: string;
  taker_name?: string;
  provider_name?: string;
  provider_trade_name?: string;
  
  created_at: string;
  updated_at: string;
  company?: Company;
}

// Tipos para requests da API
export interface CreateUserRequest {
  name: string;
  email: string;
  password: string;
  token?: string;
  role: 'admin' | 'user';
}

export interface UpdateUserRequest {
  name?: string;
  email?: string;
  password?: string;
  token?: string;
  role?: 'admin' | 'user';
  active?: boolean;
}

export interface CreateCompanyRequest {
  name: string;
  cnpj: string;
  trade_name?: string;
  address?: string;
  number?: string;
  complement?: string;
  district?: string;
  city?: string;
  state?: string;
  zip_code?: string;
  phone?: string;
  email?: string;
  company_size?: string;
  main_activity?: string;
  secondary_activity?: string;
  legal_nature?: string;
  opening_date?: string;
  registration_status?: string;
  restricted: boolean;
  auto_fetch: boolean;
}

export interface UpdateCompanyRequest {
  name?: string;
  cnpj?: string;
  trade_name?: string;
  address?: string;
  number?: string;
  complement?: string;
  district?: string;
  city?: string;
  state?: string;
  zip_code?: string;
  phone?: string;
  email?: string;
  company_size?: string;
  main_activity?: string;
  secondary_activity?: string;
  legal_nature?: string;
  opening_date?: string;
  registration_status?: string;
  restricted?: boolean;
  auto_fetch?: boolean;
  active?: boolean;
}

export interface CreateCredentialRequest {
  type: 'prefeitura_user_pass' | 'prefeitura_token' | 'prefeitura_mixed';
  name: string;
  description?: string;
  login?: string;
  password?: string;
  token?: string;
  environment?: 'production' | 'staging' | 'development';
}

export interface UpdateCredentialRequest {
  name?: string;
  description?: string;
  login?: string;
  password?: string;
  token?: string;
  environment?: 'production' | 'staging' | 'development';
  active?: boolean;
}

export interface FetchNFSeRequest {
  start_date: string; // Format: YYYY-MM-DD
  end_date: string;   // Format: YYYY-MM-DD
  page?: number;
}

export interface FetchNFSeResponse {
  success: boolean;
  message: string;
  documents_count: number;
  documents?: NFSeDocument[];
  error?: string;
}

export interface NFSeDocument {
  file_name: string;
  xml_content: string;
  processed_at: string;
}

// Tipos para respostas paginadas
export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

// Tipos para filtros
export interface UserFilters {
  role?: 'admin' | 'user';
  active?: boolean;
  page?: number;
  limit?: number;
}

export interface CompanyFilters {
  restricted?: boolean;
  active?: boolean;
  page?: number;
  limit?: number;
}

export interface DocumentFilters {
  type?: string;
  status?: 'pending' | 'processed' | 'error';
  start_date?: string;
  end_date?: string;
  page?: number;
  limit?: number;
}

// Tipos para autenticação
export interface AuthTokens {
  adminToken?: string;
  userToken?: string;
}

export interface ApiError {
  error: string;
  details?: any;
}

// CNPJ Data
export interface CNPJData {
  cnpj: string;
  name: string;
  trade_name?: string;
  address?: string;
  number?: string;
  complement?: string;
  district?: string;
  city?: string;
  state?: string;
  zip_code?: string;
  phone?: string;
  email?: string;
  company_size?: string;
  main_activity?: string;
  secondary_activities?: string[];
  legal_nature?: string;
  opening_date?: string;
  registration_status?: string;
}
