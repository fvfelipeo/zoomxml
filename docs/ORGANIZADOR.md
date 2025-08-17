# 🗂️ Organizador de NFS-e

Ferramenta para descompactar e organizar XMLs de NFS-e por competência (mês/ano) e CNPJ do prestador.

## 🎯 Funcionalidades

- ✅ **Descompactação automática** de arquivos ZIP
- ✅ **Conversão de encoding** ISO-8859-1 → UTF-8
- ✅ **Organização hierárquica** por competência e CNPJ
- ✅ **Extração de dados** principais das NFS-e
- ✅ **Geração de resumos** por prestador
- ✅ **Índice geral** de todas as NFS-e processadas

## 🚀 Como Usar

```bash
go run organizer.go organize_nfse.go
```

## 📁 Estrutura de Saída

```
xml/
├── 2025-08/                           # Competência (YYYY-MM)
│   └── 34194865000158/                # CNPJ do Prestador
│       ├── nfse_250000057_20250808.xml
│       ├── nfse_250000058_20250808.xml
│       ├── nfse_250000059_20250808.xml
│       ├── nfse_250000060_20250808.xml
│       ├── nfse_250000061_20250808.xml
│       ├── nfse_250000062_20250812.xml
│       └── resumo.txt                 # Resumo do prestador
└── indice_geral.txt                  # Índice de todas as NFS-e
```

## 📋 Dados Extraídos

Para cada NFS-e, o sistema extrai:

### Informações Básicas
- **Número da NFS-e**
- **Data de emissão**
- **Competência** (mês/ano)
- **Código de verificação**

### Prestador de Serviços
- **CNPJ**
- **Razão Social**
- **Nome Fantasia**
- **Endereço completo**
- **Inscrição Municipal**

### Tomador de Serviços
- **CNPJ/CPF**
- **Razão Social**
- **Endereço completo**

### Valores
- **Valor dos Serviços**
- **Valor do ISS**
- **Base de Cálculo**
- **Alíquota**
- **Valor Líquido**

## 📄 Arquivos Gerados

### 1. XMLs Individuais
- **Localização**: `xml/{competencia}/{cnpj}/nfse_{numero}_{data}.xml`
- **Formato**: UTF-8 (convertido de ISO-8859-1)
- **Conteúdo**: XML completo da NFS-e

### 2. Resumo do Prestador
- **Arquivo**: `resumo.txt` (em cada pasta de CNPJ)
- **Conteúdo**: Informações básicas do prestador

```
RESUMO - PRESTADOR DE SERVIÇOS
================================

CNPJ: 34194865000158
Razão Social: S. E. L. DE SOUZA SUARES VEICULOS
Competência: 2025-08

Última atualização: 17/08/2025 13:05:22
```

### 3. Índice Geral
- **Arquivo**: `indice_geral.txt` (na raiz da pasta xml)
- **Conteúdo**: Estrutura completa e estatísticas

## 🔧 Processamento Técnico

### Conversão de Encoding
```go
// Detecta e converte ISO-8859-1 para UTF-8
if strings.Contains(dataStr, "ISO-8859-1") {
    decoder := charmap.ISO8859_1.NewDecoder()
    utf8Data, _, err := transform.Bytes(decoder, data)
    // Substitui declaração de encoding
    utf8Str = strings.ReplaceAll(utf8Str, "ISO-8859-1", "UTF-8")
}
```

### Extração de Dados
```go
// Parse do XML para estruturas Go
var response ConsultarNotaResponse
xml.Unmarshal(xmlContentUTF8, &response)

// Extração de informações principais
nfse := response.ListaNfse.ComplNfse.Nfse.InfNfse
```

### Formatação de Competência
```go
// Converte "12/08/2025 00:00:00" para "2025-08"
t, err := time.Parse("02/01/2006 15:04:05", competencia)
return t.Format("2006-01")
```

## ⚠️ Requisitos

- **Go 1.21+**
- **Dependência**: `golang.org/x/text` (para conversão de encoding)
- **Arquivos ZIP**: Devem seguir o padrão `nfse_*.zip`

## 🐛 Tratamento de Erros

O sistema trata os seguintes cenários:

1. **Arquivos ZIP corrompidos**
2. **XMLs com encoding inválido**
3. **Estrutura XML malformada**
4. **Permissões de arquivo**
5. **Espaço em disco insuficiente**

## 📊 Exemplo de Execução

```
🗂️  Organizador de NFS-e por Competência e CNPJ
===================================================
📁 Encontrados 7 arquivos ZIP

📦 Processando: nfse_250000057_20250808.zip
✅ NFS-e 250000057 processada
📦 Processando: nfse_250000058_20250808.zip
✅ NFS-e 250000058 processada
...

📊 Total de NFS-e processadas: 7

📄 Salvo: xml/2025-08/34194865000158/nfse_250000057_20250808.xml
📄 Salvo: xml/2025-08/34194865000158/nfse_250000058_20250808.xml
...

🎯 Organização concluída!
```

## 🔄 Reprocessamento

Para reprocessar arquivos:
1. Delete a pasta `xml/`
2. Execute novamente o organizador
3. Os arquivos serão reorganizados do zero

## 📞 Suporte

Para dúvidas sobre o organizador:
- Verifique os logs de erro
- Confirme que os arquivos ZIP estão íntegros
- Verifique permissões de escrita no diretório
