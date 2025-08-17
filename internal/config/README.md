# 🔧 ZoomXML Configuration Module

## 📋 Visão Geral

O módulo de configuração do ZoomXML é responsável por carregar e validar todas as configurações do sistema a partir de variáveis de ambiente e arquivos `.env`.

## 🏗️ Estrutura

### Arquivos
- **`config.go`** - Estruturas de configuração e carregamento
- **`README.md`** - Esta documentação

### Estruturas de Configuração

#### 🎯 **Config Principal**
```go
type Config struct {
    App       AppConfig
    Database  DatabaseConfig
    Storage   StorageConfig
    Auth      AuthConfig
    Server    ServerConfig
    Scheduler SchedulerConfig
    Logging   LoggingConfig
    RateLimit RateLimitConfig
}
```

## 📝 Configurações Disponíveis

### 🚀 **Aplicação (AppConfig)**
| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `APP_NAME` | ZoomXML | Nome da aplicação |
| `APP_VERSION` | 1.0.0 | Versão da aplicação |
| `APP_ENV` | development | Ambiente (development/production) |
| `APP_DEBUG` | false | Modo debug |

### 🗄️ **Banco de Dados (DatabaseConfig)**
| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `DB_HOST` | localhost | Host do PostgreSQL |
| `DB_PORT` | 5432 | Porta do PostgreSQL |
| `DB_USER` | postgres | Usuário do banco |
| `DB_PASSWORD` | password | Senha do banco |
| `DB_NAME` | nfse_metadata | Nome do banco |
| `DB_SSLMODE` | disable | Modo SSL |
| `DB_MAX_OPEN_CONNS` | 25 | Máximo de conexões abertas |
| `DB_MAX_IDLE_CONNS` | 5 | Máximo de conexões idle |
| `DB_CONN_MAX_LIFETIME` | 5m | Tempo de vida das conexões |

### 💾 **Armazenamento (StorageConfig)**
| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `MINIO_ENDPOINT` | localhost:9000 | Endpoint do MinIO |
| `MINIO_ACCESS_KEY` | admin | Chave de acesso |
| `MINIO_SECRET_KEY` | password123 | Chave secreta |
| `MINIO_BUCKET` | nfse-storage | Nome do bucket |
| `MINIO_USE_SSL` | false | Usar SSL |
| `MINIO_REGION` | us-east-1 | Região |

### 🔐 **Autenticação (AuthConfig)**
| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `JWT_SECRET` | ⚠️ **OBRIGATÓRIO** | Chave secreta JWT |
| `JWT_EXPIRATION_HOURS` | 24 | Expiração do token (horas) |
| `REFRESH_TOKEN_EXPIRY` | 168h | Expiração do refresh token |
| `PASSWORD_MIN_LENGTH` | 8 | Tamanho mínimo da senha |
| `ENABLE_REFRESH_TOKENS` | true | Habilitar refresh tokens |

### 🌐 **Servidor (ServerConfig)**
| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `SERVER_HOST` | 0.0.0.0 | Host do servidor |
| `PORT` | 3000 | Porta do servidor |
| `SERVER_READ_TIMEOUT` | 30s | Timeout de leitura |
| `SERVER_WRITE_TIMEOUT` | 30s | Timeout de escrita |
| `SERVER_IDLE_TIMEOUT` | 120s | Timeout idle |
| `ENABLE_CORS` | true | Habilitar CORS |
| `ALLOWED_ORIGINS` | * | Origens permitidas |
| `ALLOWED_METHODS` | GET,POST,PUT,DELETE,OPTIONS | Métodos permitidos |
| `ALLOWED_HEADERS` | * | Headers permitidos |

### ⏰ **Agendador (SchedulerConfig)**
| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `ENABLE_AUTO_SYNC` | true | Habilitar sync automático |
| `DEFAULT_SYNC_INTERVAL` | 1h | Intervalo padrão de sync |
| `JOB_PROCESSOR_INTERVAL` | 30s | Intervalo do processador |
| `MAX_RETRIES` | 5 | Máximo de tentativas |
| `RETRY_BACKOFF_FACTOR` | 2.0 | Fator de backoff |

### 📊 **Logging (LoggingConfig)**
| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `LOG_LEVEL` | info | Nível de log |
| `LOG_FORMAT` | json | Formato do log |
| `LOG_OUTPUT` | stdout | Saída do log |
| `LOG_ENABLE_FILE` | false | Habilitar arquivo |
| `LOG_FILE_PATH` | logs/zoomxml.log | Caminho do arquivo |
| `LOG_MAX_SIZE` | 100 | Tamanho máximo (MB) |
| `LOG_MAX_BACKUPS` | 3 | Máximo de backups |
| `LOG_MAX_AGE` | 28 | Idade máxima (dias) |

### 🚦 **Rate Limiting (RateLimitConfig)**
| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `ENABLE_RATE_LIMIT` | true | Habilitar rate limiting |
| `PUBLIC_RPM` | 100 | Requests/min públicos |
| `AUTHENTICATED_RPM` | 1000 | Requests/min autenticados |
| `HEAVY_OPERATIONS_RPM` | 10 | Requests/min operações pesadas |
| `DOWNLOAD_RPM` | 50 | Requests/min downloads |

## 🚀 Como Usar

### 1. **Carregamento Básico**
```go
import "github.com/zoomxml/internal/config"

// Carregar configuração
cfg, err := config.LoadConfig()
if err != nil {
    log.Fatal("Failed to load config:", err)
}

// Usar configuração
fmt.Printf("Server running on port %d\n", cfg.Server.Port)
```

### 2. **Arquivo .env**
```bash
# Copiar exemplo
cp .env.example .env

# Editar configurações
nano .env
```

### 3. **Variáveis de Ambiente**
```bash
# Definir diretamente
export DB_HOST=production-db.example.com
export JWT_SECRET=super-secret-key-for-production

# Executar aplicação
go run cmd/zoomxml/main.go
```

## ✅ Validação

O sistema valida automaticamente:

### **Campos Obrigatórios:**
- ✅ Database host, user, dbname
- ✅ Storage endpoint, access_key, secret_key, bucket
- ✅ JWT secret (não pode ser o padrão)

### **Validações de Range:**
- ✅ Portas entre 1-65535
- ✅ JWT expiration >= 1 hora
- ✅ Password min length >= 4

### **Exemplo de Erro:**
```
configuration validation failed: JWT secret must be set and changed from default
```

## 🔧 Funções Utilitárias

### **Tipos Suportados:**
- `getEnv(key, default)` - String
- `getEnvInt(key, default)` - Integer
- `getEnvBool(key, default)` - Boolean
- `getEnvFloat(key, default)` - Float64
- `getEnvDuration(key, default)` - Duration
- `getEnvStringSlice(key, default)` - []string

### **Exemplo de Uso:**
```go
// String com fallback
host := getEnv("DB_HOST", "localhost")

// Integer com validação
port := getEnvInt("DB_PORT", 5432)

// Boolean
debug := getEnvBool("APP_DEBUG", false)

// Duration
timeout := getEnvDuration("TIMEOUT", 30*time.Second)

// String slice (separado por vírgula)
origins := getEnvStringSlice("ALLOWED_ORIGINS", []string{"*"})
```

## 🛡️ Segurança

### **⚠️ Configurações Críticas:**
1. **JWT_SECRET** - DEVE ser alterado em produção
2. **DB_PASSWORD** - Usar senhas fortes
3. **MINIO_SECRET_KEY** - Proteger chaves de acesso
4. **APP_ENV=production** - Desabilita debug

### **🔒 Boas Práticas:**
- ✅ Usar arquivo `.env` para desenvolvimento
- ✅ Usar variáveis de ambiente em produção
- ✅ Não commitar arquivos `.env` com dados reais
- ✅ Rotacionar chaves regularmente
- ✅ Usar SSL em produção (`DB_SSLMODE=require`, `MINIO_USE_SSL=true`)

## 📁 Estrutura de Arquivos

```
internal/config/
├── config.go          # Configuração principal
└── README.md          # Esta documentação

.env.example           # Exemplo de configuração
.env                   # Configuração local (não commitado)
```

## 🎯 Benefícios

1. **🔧 Flexibilidade** - Configuração via env vars ou .env
2. **✅ Validação** - Validação automática de configurações
3. **🛡️ Segurança** - Validações de segurança integradas
4. **📝 Documentação** - Todas as opções documentadas
5. **🚀 Performance** - Carregamento único na inicialização
6. **🔄 Compatibilidade** - Funciona em dev, test e prod

Este módulo garante que o ZoomXML seja **configurável**, **seguro** e **fácil de deployar**! 🎉
