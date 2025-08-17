# 🚀 ZoomXML - Sistema Multi-Empresa de NFS-e

Sistema completo multi-empresa para consulta, organização e gerenciamento de NFS-e com API REST, processamento automático, detecção inteligente de duplicatas e armazenamento em MinIO S3 + PostgreSQL.

## 📁 Estrutura do Projeto

```
zoomxml/
├── cmd/zoomxml/main.go         # 🚀 Serviço principal unificado
├── main.go                     # Aplicação CLI (legacy)
├── go.mod                      # Dependências do Go
├── go.sum                      # Lock file das dependências
├── docker-compose.yml          # PostgreSQL + MinIO + Adminer
├── init.sql                    # Schema multi-empresa
├── .env.example                # Configurações de exemplo
├── internal/                   # Código interno da aplicação
│   ├── api/                    # API REST
│   │   ├── handlers/           # Handlers HTTP
│   │   │   ├── auth.go         # Autenticação
│   │   │   └── empresa.go      # Gestão de empresas
│   │   └── middleware/         # Middlewares
│   │       └── auth.go         # Middleware JWT
│   ├── database/               # Camada de banco de dados
│   │   ├── postgres.go         # Cliente PostgreSQL original
│   │   ├── empresa_repository.go # CRUD empresas
│   │   ├── auth_repository.go  # Gestão tokens
│   │   └── job_repository.go   # Filas de jobs
│   ├── storage/                # Camada de armazenamento
│   │   ├── interface.go        # Interface de storage
│   │   └── minio.go           # Cliente MinIO S3
│   ├── models/                 # Modelos de dados
│   │   ├── nfse.go            # Estruturas NFS-e
│   │   └── empresa.go         # Modelos multi-empresa
│   ├── services/               # Lógica de negócio
│   │   ├── organizer.go       # Organizador inteligente
│   │   └── auth.go            # Serviço de autenticação
│   └── utils/                  # Utilitários
│       └── helpers.go         # Funções auxiliares
├── scripts/                    # Scripts de setup
│   ├── setup.sh               # Setup infraestrutura
│   └── start-service.sh       # Iniciar serviço completo
└── docs/                       # Documentação
```

## 🎯 Funcionalidades

### 📥 **Consulta de NFS-e**
- ✅ **Consulta de XML por período** - Endpoint `/xmlnfse`
- ✅ **XML compactado em Base64/ZIP** - Formato otimizado
- ✅ **Paginação automática** - Até 100 registros por página
- ✅ **Validação de parâmetros** - Datas e formatos
- ✅ **Decodificação automática** - Base64 para arquivo ZIP

### 🧠 **Organização Inteligente**
- ✅ **Detecção de duplicatas** - Por hash de conteúdo e número de NFS-e
- ✅ **Versionamento automático** - Controle de versões para atualizações
- ✅ **Organização hierárquica** - Por competência e CNPJ
- ✅ **Conversão de encoding** - ISO-8859-1 → UTF-8 automática
- ✅ **Processamento em lote** - Com relatórios detalhados

### 🗄️ **Gerenciamento de Metadados**
- ✅ **PostgreSQL** - Banco de dados robusto para metadados
- ✅ **Histórico de processamento** - Logs detalhados de operações
- ✅ **Cache de prestadores** - Informações otimizadas
- ✅ **Estatísticas em tempo real** - Dashboard de dados
- ✅ **Integridade de arquivos** - Verificação de checksums

## 🚀 Início Rápido

### 1. **Pré-requisitos**
```bash
# Go 1.21+
go version

# Docker & Docker Compose
docker --version
docker-compose --version
```

### 2. **Inicializar Sistema Completo**
```bash
# Clonar e entrar no diretório
git clone <repo>
cd zoomxml

# Copiar configurações
cp .env.example .env

# Inicializar e executar serviço completo
./scripts/start-service.sh
```

### 3. **Usar o Sistema**

#### **Serviço Principal (Recomendado)**
```bash
# Inicia API + Scheduler + Processamento automático
go run cmd/zoomxml/main.go
```

#### **Comandos CLI (Legacy)**
```bash
go run . fetch      # Buscar NFS-e da API
go run . organize   # Organizar XMLs existentes
go run . help       # Ver ajuda
```

## 🌐 API REST

### **Endpoints Disponíveis**

#### **Health Check**
```bash
GET /health
```

#### **Autenticação**
```bash
POST /api/v1/auth/login     # Login
POST /api/v1/auth/logout    # Logout
POST /api/v1/auth/refresh   # Refresh token
GET  /api/v1/auth/me        # Info do usuário
```

#### **Empresas**
```bash
GET    /api/v1/empresas     # Listar empresas
POST   /api/v1/empresas     # Criar empresa
GET    /api/v1/empresas/:id # Obter empresa
PUT    /api/v1/empresas/:id # Atualizar empresa
DELETE /api/v1/empresas/:id # Deletar empresa
```

#### **NFS-e (Protegido)**
```bash
POST /api/v1/nfse/sync      # Sincronização manual
GET  /api/v1/nfse/jobs      # Listar jobs
GET  /api/v1/nfse/stats     # Estatísticas
```

### **Exemplos de Uso**

#### **1. Criar Empresa**
```bash
curl -X POST http://localhost:8080/api/v1/empresas \
  -H "Content-Type: application/json" \
  -d '{
    "cnpj": "12345678000195",
    "razao_social": "Empresa Teste LTDA",
    "municipio": "imperatriz-ma",
    "security_key": "test123",
    "sync_interval_hours": 24,
    "auto_sync_enabled": true
  }'
```

#### **2. Fazer Login**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "cnpj": "12345678000195",
    "password": "test123"
  }'
```

#### **3. Usar Token (exemplo)**
```bash
# Salvar token da resposta do login
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Usar em requisições protegidas
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

## 🔧 Configuração

### **API NFS-e**
- **Município**: `imperatriz-ma`
- **URL**: `https://api-nfse-imperatriz-ma.prefeituramoderna.com.br/ws/services`
- **Security Key**: Configurado no código

### **PostgreSQL**
- **Host**: `localhost:5432`
- **Database**: `nfse_metadata`
- **User**: `postgres`
- **Password**: `password`

### **Adminer (Interface Web)**
- **URL**: `http://localhost:8080`
- **Sistema**: PostgreSQL
- **Servidor**: postgres
- **Usuário**: postgres
- **Senha**: password
- **Base de dados**: nfse_metadata

## 📊 Estrutura de Dados

### **Organização de Arquivos**
```
xml/
├── 2025-08/                    # Competência (YYYY-MM)
│   └── 34194865000158/         # CNPJ do Prestador
│       ├── nfse_250000057_20250808.xml
│       ├── nfse_250000058_20250808.xml
│       ├── resumo.txt          # Resumo do prestador
│       └── ...
└── processing_report_batch_xxx.txt  # Relatórios de processamento
```

### **Banco de Dados**
- **`nfse_metadata`** - Metadados completos das NFS-e
- **`processing_logs`** - Histórico de processamento
- **`prestador_cache`** - Cache de prestadores
- **`competencia_index`** - Índice por competência
- **`file_integrity`** - Integridade de arquivos

## 🔄 Fluxo de Trabalho

1. **Buscar NFS-e**: `go run . fetch`
   - Consulta API da Prefeitura Moderna
   - Salva XMLs como arquivos ZIP
   - Decodifica Base64 automaticamente

2. **Organizar XMLs**: `go run . organize`
   - Detecta duplicatas por hash SHA256
   - Organiza por competência e CNPJ
   - Armazena metadados no PostgreSQL
   - Gera relatórios detalhados

3. **Verificar Resultados**:
   - Arquivos organizados em `xml/`
   - Metadados em PostgreSQL
   - Relatórios de processamento
   - Interface web no Adminer

## 🛠️ Desenvolvimento

### **Compilar**
```bash
go build .
```

### **Executar Testes**
```bash
go test ./...
```

### **Limpar Dados**
```bash
# Remover XMLs organizados
rm -rf xml/

# Resetar banco de dados
docker-compose down -v
docker-compose up -d
```

## 📈 Monitoramento

### **Logs de Processamento**
- Logs detalhados no console
- Histórico no banco de dados
- Relatórios em `xml/processing_report_*.txt`

### **Estatísticas**
- Total de NFS-e processadas
- Duplicatas detectadas
- Performance de processamento
- Integridade de arquivos

### **Interface Web**
- Adminer: `http://localhost:8080`
- Visualização de dados
- Consultas SQL personalizadas
- Exportação de relatórios

## 🚨 Solução de Problemas

### **PostgreSQL não conecta**
```bash
# Verificar se está rodando
docker-compose ps

# Ver logs
docker-compose logs postgres

# Reiniciar
docker-compose restart postgres
```

### **Erro de permissão**
```bash
# Dar permissão ao script
chmod +x scripts/setup.sh
```

### **Arquivos ZIP não encontrados**
```bash
# Primeiro execute fetch para baixar
go run . fetch

# Depois organize
go run . organize
```

## 📞 Suporte

- **Documentação**: Pasta `docs/`
- **Logs**: Console e banco de dados
- **Issues**: GitHub Issues
- **API**: Documentação da Prefeitura Moderna

## 🎯 Próximos Passos

1. Executar `go run . fetch` para buscar NFS-e
2. Executar `go run . organize` para organizar
3. Acessar `http://localhost:8080` para ver metadados
4. Verificar relatórios em `xml/processing_report_*.txt`
