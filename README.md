# ZoomXML - Sistema de Gerenciamento Multi-Empresarial

Sistema de gerenciamento de documentos fiscais (NFS-e) com suporte a múltiplas empresas, autenticação baseada em papéis e armazenamento seguro.

## 🚀 Características

- **Autenticação e Autorização**: JWT + Token de Admin para operações sensíveis
- **Multi-tenancy**: Suporte a múltiplas empresas com controle de acesso
- **Empresas Restritas**: Sistema de membros para empresas privadas
- **Credenciais Seguras**: Armazenamento criptografado de credenciais externas
- **Auditoria**: Log completo de todas as operações
- **Storage**: Integração com MinIO/S3 para armazenamento de documentos
- **Banco de Dados**: PostgreSQL com migrações automáticas usando Bun ORM

## 🏗️ Arquitetura

```
internal/
├── api/                 # Camada HTTP/API
│   ├── handlers/        # Handlers das rotas
│   ├── middleware/      # Middlewares de autenticação
│   └── routes/          # Configuração de rotas
├── auth/                # Lógica de autenticação (JWT, passwords)
├── database/            # Conexão, migrações e seeders
├── models/              # Modelos do banco de dados (Bun ORM)
└── storage/             # Serviços de armazenamento (MinIO)
```

## 🛠️ Desenvolvimento Local

### Pré-requisitos

- Go 1.23+
- Docker e Docker Compose

### Configuração

1. **Clone o repositório**
```bash
git clone <repo-url>
cd zoomxml
```

2. **Configure as variáveis de ambiente**
```bash
cp .env.example .env
# Edite o .env conforme necessário
```

3. **Inicie os serviços de desenvolvimento**
```bash
docker-compose -f docker-compose.dev.yml up -d
```

Isso iniciará:
- **PostgreSQL** (porta 5432)
- **MinIO** (porta 9000 - API, 9001 - Console)
- **DBGate** (porta 8080 - Interface de banco)
- **Redis** (porta 6379 - Cache)

4. **Execute a aplicação**
```bash
go run ./cmd/zoomxml
```

A aplicação estará disponível em `http://localhost:8000`

### Serviços Disponíveis

- **Aplicação**: http://localhost:8000
- **Health Check**: http://localhost:8000/health
- **Swagger/OpenAPI**: http://localhost:8000/swagger/
- **DBGate (DB Admin)**: http://localhost:8080
- **MinIO Console**: http://localhost:9001 (admin/password123)

## 📚 API Endpoints

### 📖 Documentação da API

A documentação completa da API está disponível via **Swagger/OpenAPI** em:
**http://localhost:8000/swagger/**

### Usuários (Admin Token Required)

```http
POST   /api/users              # Criar usuário
GET    /api/users              # Listar usuários
GET    /api/users/:id          # Obter usuário
PATCH  /api/users/:id          # Editar usuário
DELETE /api/users/:id          # Remover usuário
```

**Header necessário**: `Authorization: Bearer <ADMIN_TOKEN>`

### Empresas

```http
POST   /api/companies          # Criar empresa (autenticação necessária)
GET    /api/companies          # Listar empresas (regras de visibilidade)
GET    /api/companies/:id      # Obter empresa (regras de visibilidade)
PATCH  /api/companies/:id      # Atualizar empresa (autenticação necessária)
DELETE /api/companies/:id      # Deletar empresa (apenas admin)
```

## 🔐 Sistema de Autenticação

### 1. Token de Admin
- Definido na variável `ADMIN_TOKEN` do `.env`
- Necessário para todas as operações de usuários
- Usado no header: `Authorization: Bearer <ADMIN_TOKEN>`

### 2. JWT de Usuários
- Gerado após login de usuários
- Usado para operações normais da aplicação
- Contém informações do usuário (ID, email, role)

### 3. Papéis de Usuário
- **admin**: Acesso total ao sistema
- **user**: Acesso limitado conforme regras de negócio

## 🏢 Sistema de Empresas

### Visibilidade
- **Empresas Públicas**: Visíveis para todos os usuários
- **Empresas Restritas**: Apenas membros e admins podem acessar
- **Admins**: Sempre veem todas as empresas

### Membros
- Empresas restritas podem ter membros específicos
- Tabela `company_members` gerencia os vínculos
- Apenas usuários vinculados podem acessar empresas restritas

## 🗄️ Banco de Dados

### Modelos Principais

- **users**: Usuários do sistema
- **companies**: Empresas cadastradas
- **company_members**: Vínculos usuário ↔ empresa (empresas restritas)
- **company_credentials**: Credenciais externas das empresas
- **documents**: Documentos fiscais
- **audit_logs**: Logs de auditoria

### Migrações
- Migrações automáticas usando Bun ORM
- Executadas na inicialização da aplicação
- Seeders automáticos em desenvolvimento

## 📦 Dados Iniciais (Desenvolvimento)

O sistema cria automaticamente:
- **Admin padrão**: admin@zoomxml.com / admin123
- **Empresa exemplo**: Empresa Exemplo LTDA

## 🔧 Configuração

Principais variáveis do `.env`:

```env
# Aplicação
APP_ENV=development
PORT=3000

# Banco de Dados
DB_HOST=localhost
DB_PORT=5432
DB_NAME=nfse_metadata

# Autenticação
JWT_SECRET=your-secret-key
ADMIN_TOKEN=admin-secret-token

# Storage
MINIO_ENDPOINT=localhost:9000
MINIO_BUCKET=nfse-storage
```

## 📖 Documentação Swagger

A API possui documentação automática gerada via Swagger/OpenAPI.

### Acessar Documentação
- **URL**: http://localhost:8000/swagger/
- **Formato JSON**: http://localhost:8000/swagger/doc.json

### Regenerar Documentação
```bash
# Instalar swag (se não estiver instalado)
go install github.com/swaggo/swag/cmd/swag@latest

# Gerar documentação
swag init -g cmd/zoomxml/main.go -o docs
```

## 🧪 Testes

```bash
# Executar testes
go test ./...

# Executar com coverage
go test -cover ./...
```

## 📝 Logs e Monitoramento

- **Health Check**: `/health`
- **Logs estruturados**: JSON em produção
- **Auditoria**: Todas as operações são logadas
- **Métricas**: Prontas para integração com Prometheus

## 🚀 Deploy

### Docker Compose (Produção)
```bash
docker-compose up -d
```

### Variáveis Importantes para Produção
- Alterar `JWT_SECRET`
- Alterar `ADMIN_TOKEN`
- Configurar `APP_ENV=production`
- Configurar SSL para MinIO
- Configurar backup do PostgreSQL

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para detalhes.
