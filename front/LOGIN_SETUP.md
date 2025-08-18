# Login Setup - Backend Integrado

## ⚠️ Importante
O login agora está integrado com o backend Go real. Você precisa:

1. **Iniciar o backend Go** na porta 8000
2. **Criar usuários** usando o endpoint `/api/users` com o admin token
3. **Usar as credenciais** dos usuários criados no banco de dados

## Pré-requisitos

1. Backend Go rodando em `http://localhost:8000`
2. Banco de dados PostgreSQL configurado
3. Admin token configurado no `.env`: `ADMIN_TOKEN=admin-secret-token`

## Como Criar Usuários

Use o admin token do `.env` para criar usuários via API:

### Criar Admin User
```bash
curl -X POST http://localhost:8000/api/users \
  -H "Authorization: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin User",
    "email": "admin@zoomxml.com", 
    "password": "admin123456",
    "role": "admin"
  }'
```

### Criar Regular User
```bash
curl -X POST http://localhost:8000/api/users \
  -H "Authorization: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Regular User",
    "email": "user@zoomxml.com",
    "password": "user123456", 
    "role": "user"
  }'
```

## Usuários Sugeridos para Teste

### Admin User
- **Email:** `admin@zoomxml.com`
- **Senha:** `admin123456`
- **Role:** `admin`

### Regular User  
- **Email:** `user@zoomxml.com`
- **Senha:** `user123456`
- **Role:** `user`

## Como Testar

1. Acesse `http://localhost:3000/login`
2. Use uma das credenciais criadas
3. Após login bem-sucedido, será redirecionado para `/dashboard`
4. O token será armazenado no localStorage

## Funcionalidades

- ✅ **Autenticação Real** - Integrada com backend Go
- ✅ **Validação de Senha** - Usando bcrypt no backend
- ✅ **Token de Usuário** - Token único por usuário armazenado no banco
- ✅ **Validação de Formulário** - Client-side com mensagens de erro
- ✅ **Estados de Loading** - Indicadores visuais
- ✅ **Gerenciamento de Sessão** - Token storage no localStorage
- ✅ **Design Responsivo** - shadcn/ui components

## Arquitetura

```
Frontend (Next.js) → API Route (/api/auth/login) → Backend Go (/api/auth/login) → PostgreSQL
```

## Troubleshooting

### Backend não está rodando
- Verifique se o backend Go está rodando na porta 8000
- Teste: `curl http://localhost:8000/api/users -H "Authorization: admin-secret-token"`

### Usuário não existe
- Crie o usuário usando o comando curl acima
- Verifique se o email está correto

### Senha incorreta
- Verifique se a senha está correta
- Lembre-se que as senhas são case-sensitive

### Token inválido
- O token é gerado automaticamente quando o usuário é criado
- Cada usuário tem um token único armazenado no banco
