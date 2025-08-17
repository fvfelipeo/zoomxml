-- NFS-e Metadata Database Initialization Script
-- PostgreSQL version

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schemas
CREATE SCHEMA IF NOT EXISTS nfse;
CREATE SCHEMA IF NOT EXISTS audit;

-- Set search path
SET search_path TO nfse, public;

-- NFS-e metadata table
CREATE TABLE IF NOT EXISTS nfse_metadata (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    numero_nfse VARCHAR(50) NOT NULL,
    content_hash VARCHAR(64) UNIQUE NOT NULL,
    file_path TEXT NOT NULL,
    source_zip_file TEXT NOT NULL,
    data_emissao TIMESTAMP NOT NULL,
    competencia TEXT NOT NULL,
    competencia_formatada VARCHAR(7) NOT NULL, -- YYYY-MM format
    prestador_cnpj VARCHAR(14) NOT NULL,
    prestador_razao TEXT NOT NULL,
    prestador_nome_fantasia TEXT,
    prestador_endereco JSONB,
    tomador_cnpj VARCHAR(14),
    tomador_cpf VARCHAR(11),
    tomador_razao TEXT,
    tomador_endereco JSONB,
    valor_servicos DECIMAL(15,2),
    valor_iss DECIMAL(15,2),
    valor_liquido DECIMAL(15,2),
    aliquota DECIMAL(5,4),
    base_calculo DECIMAL(15,2),
    codigo_verificacao VARCHAR(100),
    natureza_operacao INTEGER,
    item_lista_servico VARCHAR(10),
    codigo_cnae VARCHAR(20),
    discriminacao TEXT,
    file_size BIGINT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'active',
    error_message TEXT,
    metadata_json JSONB, -- Store additional metadata as JSON
    CONSTRAINT unique_nfse_prestador_competencia UNIQUE(numero_nfse, prestador_cnpj, competencia_formatada)
);

-- Processing logs table
CREATE TABLE IF NOT EXISTS processing_logs (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    operation VARCHAR(50) NOT NULL,
    source_file TEXT NOT NULL,
    target_file TEXT,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    error_details JSONB,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    duration_ms BIGINT,
    batch_id UUID,
    user_agent TEXT,
    ip_address INET
);

-- Prestador cache table
CREATE TABLE IF NOT EXISTS prestador_cache (
    cnpj VARCHAR(14) PRIMARY KEY,
    razao_social TEXT NOT NULL,
    nome_fantasia TEXT,
    endereco JSONB,
    inscricao_municipal VARCHAR(50),
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    nfse_count INTEGER DEFAULT 0,
    total_valor_servicos DECIMAL(15,2) DEFAULT 0,
    total_valor_iss DECIMAL(15,2) DEFAULT 0,
    competencias TEXT[], -- Array of competencias
    status VARCHAR(20) DEFAULT 'active',
    metadata_json JSONB
);

-- File integrity table
CREATE TABLE IF NOT EXISTS file_integrity (
    id SERIAL PRIMARY KEY,
    file_path TEXT UNIQUE NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    file_size BIGINT NOT NULL,
    last_verified TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verification_status VARCHAR(20) DEFAULT 'verified',
    backup_path TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Competencia index table for fast lookups
CREATE TABLE IF NOT EXISTS competencia_index (
    competencia_formatada VARCHAR(7) PRIMARY KEY,
    ano INTEGER NOT NULL,
    mes INTEGER NOT NULL,
    nfse_count INTEGER DEFAULT 0,
    prestador_count INTEGER DEFAULT 0,
    total_valor_servicos DECIMAL(15,2) DEFAULT 0,
    total_valor_iss DECIMAL(15,2) DEFAULT 0,
    first_nfse_date TIMESTAMP,
    last_nfse_date TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit schema tables
CREATE TABLE IF NOT EXISTS audit.nfse_changes (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(50) NOT NULL,
    record_id INTEGER NOT NULL,
    operation VARCHAR(10) NOT NULL, -- INSERT, UPDATE, DELETE
    old_values JSONB,
    new_values JSONB,
    changed_by VARCHAR(100),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    change_reason TEXT
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_nfse_numero ON nfse_metadata(numero_nfse);
CREATE INDEX IF NOT EXISTS idx_nfse_hash ON nfse_metadata(content_hash);
CREATE INDEX IF NOT EXISTS idx_nfse_prestador ON nfse_metadata(prestador_cnpj);
CREATE INDEX IF NOT EXISTS idx_nfse_competencia ON nfse_metadata(competencia_formatada);
CREATE INDEX IF NOT EXISTS idx_nfse_data_emissao ON nfse_metadata(data_emissao);
CREATE INDEX IF NOT EXISTS idx_nfse_status ON nfse_metadata(status);
CREATE INDEX IF NOT EXISTS idx_nfse_processed_at ON nfse_metadata(processed_at);
CREATE INDEX IF NOT EXISTS idx_nfse_valor_servicos ON nfse_metadata(valor_servicos);

-- Composite indexes
CREATE INDEX IF NOT EXISTS idx_nfse_prestador_competencia ON nfse_metadata(prestador_cnpj, competencia_formatada);
CREATE INDEX IF NOT EXISTS idx_nfse_competencia_status ON nfse_metadata(competencia_formatada, status);

-- Processing logs indexes
CREATE INDEX IF NOT EXISTS idx_logs_operation ON processing_logs(operation);
CREATE INDEX IF NOT EXISTS idx_logs_date ON processing_logs(processed_at);
CREATE INDEX IF NOT EXISTS idx_logs_status ON processing_logs(status);
CREATE INDEX IF NOT EXISTS idx_logs_batch ON processing_logs(batch_id);

-- Prestador cache indexes
CREATE INDEX IF NOT EXISTS idx_prestador_razao ON prestador_cache(razao_social);
CREATE INDEX IF NOT EXISTS idx_prestador_last_seen ON prestador_cache(last_seen);

-- File integrity indexes
CREATE INDEX IF NOT EXISTS idx_file_hash ON file_integrity(content_hash);
CREATE INDEX IF NOT EXISTS idx_file_verified ON file_integrity(last_verified);

-- Create functions for automatic updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers
CREATE TRIGGER update_nfse_metadata_updated_at 
    BEFORE UPDATE ON nfse_metadata 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_competencia_index_updated_at 
    BEFORE UPDATE ON competencia_index 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to update competencia statistics
CREATE OR REPLACE FUNCTION update_competencia_stats()
RETURNS TRIGGER AS $$
BEGIN
    -- Update competencia_index when nfse_metadata changes
    INSERT INTO competencia_index (
        competencia_formatada, ano, mes, nfse_count, prestador_count,
        total_valor_servicos, total_valor_iss, first_nfse_date, last_nfse_date
    )
    SELECT 
        NEW.competencia_formatada,
        EXTRACT(YEAR FROM NEW.data_emissao)::INTEGER,
        EXTRACT(MONTH FROM NEW.data_emissao)::INTEGER,
        1,
        1,
        COALESCE(NEW.valor_servicos, 0),
        COALESCE(NEW.valor_iss, 0),
        NEW.data_emissao,
        NEW.data_emissao
    ON CONFLICT (competencia_formatada) DO UPDATE SET
        nfse_count = competencia_index.nfse_count + 1,
        prestador_count = (
            SELECT COUNT(DISTINCT prestador_cnpj) 
            FROM nfse_metadata 
            WHERE competencia_formatada = NEW.competencia_formatada 
            AND status = 'active'
        ),
        total_valor_servicos = competencia_index.total_valor_servicos + COALESCE(NEW.valor_servicos, 0),
        total_valor_iss = competencia_index.total_valor_iss + COALESCE(NEW.valor_iss, 0),
        first_nfse_date = LEAST(competencia_index.first_nfse_date, NEW.data_emissao),
        last_nfse_date = GREATEST(competencia_index.last_nfse_date, NEW.data_emissao),
        updated_at = NOW();
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for competencia stats
CREATE TRIGGER update_competencia_stats_trigger
    AFTER INSERT ON nfse_metadata
    FOR EACH ROW EXECUTE FUNCTION update_competencia_stats();

-- Function to update prestador cache
CREATE OR REPLACE FUNCTION update_prestador_cache()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO prestador_cache (
        cnpj, razao_social, nome_fantasia, nfse_count,
        total_valor_servicos, total_valor_iss, competencias
    )
    VALUES (
        NEW.prestador_cnpj,
        NEW.prestador_razao,
        NEW.prestador_nome_fantasia,
        1,
        COALESCE(NEW.valor_servicos, 0),
        COALESCE(NEW.valor_iss, 0),
        ARRAY[NEW.competencia_formatada]
    )
    ON CONFLICT (cnpj) DO UPDATE SET
        razao_social = NEW.prestador_razao,
        nome_fantasia = COALESCE(NEW.prestador_nome_fantasia, prestador_cache.nome_fantasia),
        last_seen = NOW(),
        nfse_count = prestador_cache.nfse_count + 1,
        total_valor_servicos = prestador_cache.total_valor_servicos + COALESCE(NEW.valor_servicos, 0),
        total_valor_iss = prestador_cache.total_valor_iss + COALESCE(NEW.valor_iss, 0),
        competencias = array_append(
            CASE WHEN NEW.competencia_formatada = ANY(prestador_cache.competencias) 
                 THEN prestador_cache.competencias 
                 ELSE prestador_cache.competencias 
            END, 
            NEW.competencia_formatada
        );
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for prestador cache
CREATE TRIGGER update_prestador_cache_trigger
    AFTER INSERT ON nfse_metadata
    FOR EACH ROW EXECUTE FUNCTION update_prestador_cache();

-- Create views for common queries
CREATE OR REPLACE VIEW v_nfse_summary AS
SELECT 
    n.competencia_formatada,
    n.prestador_cnpj,
    n.prestador_razao,
    COUNT(*) as total_nfse,
    SUM(n.valor_servicos) as total_valor_servicos,
    SUM(n.valor_iss) as total_valor_iss,
    MIN(n.data_emissao) as primeira_emissao,
    MAX(n.data_emissao) as ultima_emissao
FROM nfse_metadata n
WHERE n.status = 'active'
GROUP BY n.competencia_formatada, n.prestador_cnpj, n.prestador_razao
ORDER BY n.competencia_formatada DESC, n.prestador_razao;

-- Create view for processing statistics
CREATE OR REPLACE VIEW v_processing_stats AS
SELECT 
    operation,
    status,
    COUNT(*) as total_operations,
    AVG(duration_ms) as avg_duration_ms,
    MIN(processed_at) as first_operation,
    MAX(processed_at) as last_operation
FROM processing_logs
GROUP BY operation, status
ORDER BY operation, status;

-- Grant permissions (if needed for specific user)
-- GRANT ALL PRIVILEGES ON SCHEMA nfse TO nfse_user;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA nfse TO nfse_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA nfse TO nfse_user;

-- Insert initial data or configuration if needed
INSERT INTO competencia_index (competencia_formatada, ano, mes, nfse_count) 
VALUES ('2025-08', 2025, 8, 0) 
ON CONFLICT (competencia_formatada) DO NOTHING;

-- Create a function to get database statistics
CREATE OR REPLACE FUNCTION get_database_stats()
RETURNS TABLE(
    total_nfse BIGINT,
    total_prestadores BIGINT,
    total_competencias BIGINT,
    total_valor_servicos NUMERIC,
    total_valor_iss NUMERIC,
    last_processed TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        (SELECT COUNT(*) FROM nfse_metadata WHERE status = 'active'),
        (SELECT COUNT(*) FROM prestador_cache WHERE status = 'active'),
        (SELECT COUNT(*) FROM competencia_index),
        (SELECT COALESCE(SUM(valor_servicos), 0) FROM nfse_metadata WHERE status = 'active'),
        (SELECT COALESCE(SUM(valor_iss), 0) FROM nfse_metadata WHERE status = 'active'),
        (SELECT MAX(processed_at) FROM nfse_metadata);
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- SISTEMA MULTI-EMPRESA
-- =====================================================

-- Tabela de empresas
CREATE TABLE IF NOT EXISTS empresas (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    cnpj VARCHAR(14) UNIQUE NOT NULL,
    razao_social VARCHAR(255) NOT NULL,
    nome_fantasia VARCHAR(255),
    municipio VARCHAR(100) NOT NULL,
    security_key TEXT NOT NULL,
    api_endpoint TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    configuracoes JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_sync TIMESTAMP WITH TIME ZONE,
    sync_interval_hours INTEGER DEFAULT 24,
    auto_sync_enabled BOOLEAN DEFAULT true
);

-- Índices para empresas
CREATE INDEX IF NOT EXISTS idx_empresas_cnpj ON empresas(cnpj);
CREATE INDEX IF NOT EXISTS idx_empresas_status ON empresas(status);
CREATE INDEX IF NOT EXISTS idx_empresas_municipio ON empresas(municipio);
CREATE INDEX IF NOT EXISTS idx_empresas_last_sync ON empresas(last_sync);

-- Tabela de tokens JWT para autenticação
CREATE TABLE IF NOT EXISTS auth_tokens (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    empresa_id INTEGER REFERENCES empresas(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true
);

-- Índices para tokens
CREATE INDEX IF NOT EXISTS idx_auth_tokens_empresa ON auth_tokens(empresa_id);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_hash ON auth_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_expires ON auth_tokens(expires_at);

-- Tabela de jobs de processamento
CREATE TABLE IF NOT EXISTS processing_jobs (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    empresa_id INTEGER REFERENCES empresas(id) ON DELETE CASCADE,
    job_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    priority INTEGER DEFAULT 5,
    scheduled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    parameters JSONB DEFAULT '{}',
    result JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índices para jobs
CREATE INDEX IF NOT EXISTS idx_processing_jobs_empresa ON processing_jobs(empresa_id);
CREATE INDEX IF NOT EXISTS idx_processing_jobs_status ON processing_jobs(status);
CREATE INDEX IF NOT EXISTS idx_processing_jobs_type ON processing_jobs(job_type);
CREATE INDEX IF NOT EXISTS idx_processing_jobs_scheduled ON processing_jobs(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_processing_jobs_priority ON processing_jobs(priority DESC);

-- Adicionar empresa_id à tabela nfse_metadata
ALTER TABLE nfse_metadata ADD COLUMN IF NOT EXISTS empresa_id INTEGER REFERENCES empresas(id);
CREATE INDEX IF NOT EXISTS idx_nfse_metadata_empresa ON nfse_metadata(empresa_id);

-- Adicionar empresa_id às outras tabelas
ALTER TABLE processing_logs ADD COLUMN IF NOT EXISTS empresa_id INTEGER REFERENCES empresas(id);
CREATE INDEX IF NOT EXISTS idx_processing_logs_empresa ON processing_logs(empresa_id);

ALTER TABLE prestador_cache ADD COLUMN IF NOT EXISTS empresa_id INTEGER REFERENCES empresas(id);
CREATE INDEX IF NOT EXISTS idx_prestador_cache_empresa ON prestador_cache(empresa_id);

-- Tabela de configurações do sistema
CREATE TABLE IF NOT EXISTS system_config (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Configurações padrão
INSERT INTO system_config (key, value, description) VALUES
('default_sync_interval', '24', 'Intervalo padrão de sincronização em horas'),
('max_concurrent_jobs', '5', 'Máximo de jobs simultâneos'),
('job_timeout_minutes', '60', 'Timeout padrão para jobs em minutos'),
('minio_bucket_name', 'nfse-storage', 'Nome do bucket padrão no MinIO'),
('api_rate_limit', '100', 'Limite de requisições por minuto por empresa')
ON CONFLICT (key) DO NOTHING;

-- Função para atualizar updated_at automaticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers para updated_at
CREATE TRIGGER update_empresas_updated_at BEFORE UPDATE ON empresas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_processing_jobs_updated_at BEFORE UPDATE ON processing_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_config_updated_at BEFORE UPDATE ON system_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Função para obter estatísticas por empresa
CREATE OR REPLACE FUNCTION get_empresa_stats(empresa_uuid UUID)
RETURNS TABLE(
    total_nfse BIGINT,
    total_valor_servicos NUMERIC,
    total_valor_iss NUMERIC,
    last_processed TIMESTAMP WITH TIME ZONE,
    total_prestadores BIGINT,
    total_competencias BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        (SELECT COUNT(*) FROM nfse_metadata nm
         JOIN empresas e ON nm.empresa_id = e.id
         WHERE e.uuid = empresa_uuid AND nm.status = 'active'),
        (SELECT COALESCE(SUM(nm.valor_servicos), 0) FROM nfse_metadata nm
         JOIN empresas e ON nm.empresa_id = e.id
         WHERE e.uuid = empresa_uuid AND nm.status = 'active'),
        (SELECT COALESCE(SUM(nm.valor_iss), 0) FROM nfse_metadata nm
         JOIN empresas e ON nm.empresa_id = e.id
         WHERE e.uuid = empresa_uuid AND nm.status = 'active'),
        (SELECT MAX(nm.processed_at) FROM nfse_metadata nm
         JOIN empresas e ON nm.empresa_id = e.id
         WHERE e.uuid = empresa_uuid),
        (SELECT COUNT(DISTINCT nm.prestador_cnpj) FROM nfse_metadata nm
         JOIN empresas e ON nm.empresa_id = e.id
         WHERE e.uuid = empresa_uuid AND nm.status = 'active'),
        (SELECT COUNT(DISTINCT nm.competencia_formatada) FROM nfse_metadata nm
         JOIN empresas e ON nm.empresa_id = e.id
         WHERE e.uuid = empresa_uuid AND nm.status = 'active');
END;
$$ LANGUAGE plpgsql;
