package routes

/*
API Routes Documentation
========================

Base URL: http://localhost:3000/api/v1

## Public Routes

### Health Check
GET /health
- Description: Check service health status
- Authentication: None
- Response: Service status and component health

## Authentication Routes

### Login
POST /api/v1/auth/login
- Description: Authenticate user and get JWT token
- Authentication: None
- Body: {"cnpj": "string", "password": "string"}
- Response: JWT token and user info

### Logout
POST /api/v1/auth/logout
- Description: Invalidate current JWT token
- Authentication: Bearer Token
- Response: Success confirmation

### Refresh Token
POST /api/v1/auth/refresh
- Description: Refresh JWT token
- Authentication: Bearer Token
- Response: New JWT token

## Protected Routes (Require Authentication)

### User Info
GET /api/v1/auth/me
- Description: Get current authenticated user info
- Authentication: Bearer Token
- Response: User/empresa information

## Empresa Management Routes

### Create Empresa
POST /api/v1/empresas
- Description: Create a new empresa
- Authentication: Bearer Token
- Body: EmpresaCreateRequest
- Response: Created empresa

### List Empresas
GET /api/v1/empresas
- Description: List empresas with pagination
- Authentication: Bearer Token
- Query Params: page, per_page, status
- Response: Paginated list of empresas

### Get Empresa
GET /api/v1/empresas/{id}
- Description: Get empresa by ID
- Authentication: Bearer Token
- Response: Empresa details

### Update Empresa
PUT /api/v1/empresas/{id}
- Description: Update empresa
- Authentication: Bearer Token
- Body: EmpresaUpdateRequest
- Response: Updated empresa

### Delete Empresa
DELETE /api/v1/empresas/{id}
- Description: Delete empresa (soft delete)
- Authentication: Bearer Token
- Response: Success confirmation

## NFS-e Routes

### Manual Sync
POST /api/v1/nfse/sync
- Description: Trigger manual XML consultation for authenticated empresa
- Authentication: Bearer Token
- Response: Job created for sync

### List Jobs
GET /api/v1/nfse/jobs
- Description: List processing jobs for authenticated empresa
- Authentication: Bearer Token
- Query Params: page, per_page, status, job_type
- Response: Paginated list of jobs

### Get Statistics
GET /api/v1/nfse/stats
- Description: Get NFS-e statistics for authenticated empresa
- Authentication: Bearer Token
- Response: Statistics summary

## XML Consumption Routes (Stored Data)

### List All XMLs
GET /api/v1/nfse/xmls
- Description: List all stored XMLs for authenticated empresa
- Authentication: Bearer Token
- Query Params: page, per_page
- Response: Paginated list of XML files

### List XMLs by Competência
GET /api/v1/nfse/xmls/{competencia}
- Description: List XMLs for specific competência (YYYY-MM format)
- Authentication: Bearer Token
- Path Params: competencia (e.g., "2025-01")
- Response: List of XML files for competência

### Get XML Content
GET /api/v1/nfse/xml/{competencia}/{numero}
- Description: Get XML content for specific NFS-e
- Authentication: Bearer Token
- Path Params: competencia, numero
- Response: XML content and metadata

### Download XML File
GET /api/v1/nfse/xml/{competencia}/{numero}/download
- Description: Download XML file
- Authentication: Bearer Token
- Path Params: competencia, numero
- Response: XML file download

## Automatic Processing

The system automatically:
1. Consults XMLs from external APIs based on empresa configuration
2. Stores XMLs in MinIO S3 storage with organized structure
3. Processes jobs in background with retry logic
4. Maintains metadata in PostgreSQL database

## Authentication

All protected routes require a Bearer token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

## Error Responses

All endpoints return standardized error responses:
```json
{
  "success": false,
  "error": "Error message description"
}
```

## Success Responses

All endpoints return standardized success responses:
```json
{
  "success": true,
  "message": "Optional success message",
  "data": {}, // Response data
  "meta": {   // Optional pagination metadata
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

## Rate Limiting

- Default: 100 requests per minute per IP
- Authenticated: 1000 requests per minute per user

## File Organization

XMLs are stored in MinIO with the following structure:
```
{cnpj}/{ano}/{mes}/xml/nfse_{numero}_{data}.xml
{cnpj}/{ano}/{mes}/zip/nfse_{numero}_{data}.zip
{cnpj}/relatorios/processing_report_{batch_id}.txt
```

Example:
```
12345678000195/2025/01/xml/nfse_000001_20250108.xml
12345678000195/2025/01/zip/nfse_000001_20250108.zip
12345678000195/relatorios/processing_report_batch_001.txt
```
*/
