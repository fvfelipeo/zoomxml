# Sistema Inteligente de NFS-e

Sistema completo para consulta, organização e gerenciamento de NFS-e com detecção inteligente de duplicatas e armazenamento de metadados em PostgreSQL.

## 🎯 Funcionalidades Principais

### 📥 **Consulta de NFS-e**
- ✅ **Consulta de XML por período** - Endpoint `/xmInfse`
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

## 🔧 Configuração do Sistema

### Pré-requisitos
- **Go 1.21+**
- **Docker & Docker Compose**
- **PostgreSQL** (via Docker)

### Configuração da API
- **Município**: `imperatriz-ma`
- **URL**: `https://api-nfse-imperatriz-ma.prefeituramoderna.com.br/ws/services`
- **Security Key**: `69415f14b56ccabe8cc5ec8cf5d5a2d2dc2ac66f0bb9859484dd5f8ce7ae2d2a`

## 🚀 Como Executar

1. Certifique-se de ter o Go instalado (versão 1.21 ou superior)
2. Execute o comando:

```bash
go run .
```

## 📋 Estrutura do Código

### Principais Estruturas

- **`NFSeClient`**: Cliente HTTP otimizado para NFS-e
- **`ConsultaXMLRequest`**: Parâmetros de consulta por período
- **`NFSeXMLResponse`**: Resposta estruturada da API
- **`NFSeXMLItem`**: Item individual de NFS-e

### Parâmetros de Consulta

| Parâmetro | Tipo | Obrigatório | Descrição |
|-----------|------|-------------|-----------|
| `DataInicial` | string | ✅ | Data inicial (YYYY-MM-DD) |
| `DataFinal` | string | ✅ | Data final (YYYY-MM-DD) |
| `NumeroInicial` | string | ❌ | Número inicial da NFS-e |
| `NumeroFinal` | string | ❌ | Número final da NFS-e |
| `Competencia` | string | ❌ | Competência (YYYY MM) |
| `Pagina` | int | ❌ | Número da página (padrão: 1) |

## 📊 Exemplo de Resultado

```
🚀 Cliente NFS-e - Consulta de XML por Período
===================================================
🏛️  Município: imperatriz-ma
🔗 URL: https://api-nfse-imperatriz-ma.prefeituramoderna.com.br/ws/services

🔍 Verificando conectividade...
✅ Conectividade OK - Último RPS: 0

📅 Consultando NFS-e de 01 a 17 de Agosto de 2025...
✅ Consulta realizada com sucesso!
📄 Total de registros: 6
📑 Página atual: 1 de 1
📋 Registros por página: 100

📋 NFS-e encontradas:
  1. NFS-e: 250000062 | Data: 2025-08-12 10:54:30 | Competência: 202508
      📦 XML: 1836 caracteres
  2. NFS-e: 250000061 | Data: 2025-08-08 15:53:11 | Competência: 202508
      📦 XML: 1856 caracteres
  ...

💾 Salvando 6 XMLs compactados...
✅ XML da NFS-e 250000062 salvo em: nfse_250000062_20250812.zip
✅ XML da NFS-e 250000061 salvo em: nfse_250000061_20250808.zip
...

🎯 Consulta concluída!
```

## 📁 Arquivos Gerados

O sistema gera dois tipos de arquivos:

### 1. Arquivos ZIP (temporários)
- **Formato**: `nfse_{numero}_{data}.zip`
- **Conteúdo**: XML da NFS-e compactado em Base64
- **Exemplo**: `nfse_250000062_20250812.zip`

### 2. Estrutura Organizada (pasta `xml/`)
```
xml/
├── 2025-08/                    # Competência (YYYY-MM)
│   └── 34194865000158/         # CNPJ do Prestador
│       ├── nfse_250000057_20250808.xml
│       ├── nfse_250000058_20250808.xml
│       ├── nfse_250000059_20250808.xml
│       ├── nfse_250000060_20250808.xml
│       ├── nfse_250000061_20250808.xml
│       ├── nfse_250000062_20250812.xml
│       └── resumo.txt          # Resumo do prestador
└── indice_geral.txt           # Índice de todas as NFS-e
```

## 🗂️ Organizador de XMLs

Para organizar os XMLs por competência e CNPJ:

```bash
go run organizer.go organize_nfse.go
```

**Funcionalidades do Organizador:**
- ✅ Descompacta arquivos ZIP automaticamente
- ✅ Converte encoding ISO-8859-1 para UTF-8
- ✅ Organiza por competência (mês/ano)
- ✅ Agrupa por CNPJ do prestador
- ✅ Gera resumos e índices
- ✅ Extrai informações principais das NFS-e

## Personalização

### Configurando o Município

Você pode configurar o município de duas formas:

1. **Modificando a constante** no início do arquivo `main.go`:
```go
const (
    BaseURL = "https://api-nfse-seumunicipio-uf.prefeituramoderna.com.br/ws/services"
)
```

2. **Usando a função específica** na função main:
```go
client := NewNFSeClientWithMunicipio("seumunicipio-uf")
```

### Exemplos de URLs de Município:
- São Paulo-SP: `saopaulo-sp`
- Rio de Janeiro-RJ: `riodejaneiro-rj`
- Belo Horizonte-MG: `belohorizonte-mg`

## Observações

- Os dados de exemplo são fictícios e podem retornar erros de validação
- Para testes reais, substitua pelos dados válidos do seu município
- Consulte o manual da NFS-e para detalhes sobre validações e formatos específicos
