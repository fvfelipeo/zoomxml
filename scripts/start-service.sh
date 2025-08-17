#!/bin/bash

# ZoomXML Service Startup Script
echo "🚀 Starting ZoomXML Multi-Enterprise NFS-e Service"
echo "=================================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Start infrastructure services
echo "📦 Starting infrastructure services..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 15

# Check PostgreSQL
echo "🔍 Checking PostgreSQL..."
max_attempts=30
attempt=1
while [ $attempt -le $max_attempts ]; do
    if docker-compose exec -T postgres pg_isready -U postgres -d nfse_metadata > /dev/null 2>&1; then
        echo "✅ PostgreSQL is ready!"
        break
    fi
    
    if [ $attempt -eq $max_attempts ]; then
        echo "❌ PostgreSQL failed to start after $max_attempts attempts"
        docker-compose logs postgres
        exit 1
    fi
    
    echo "⏳ Attempt $attempt/$max_attempts - PostgreSQL not ready yet..."
    sleep 2
    ((attempt++))
done

# Check MinIO
echo "🔍 Checking MinIO..."
max_attempts=30
attempt=1
while [ $attempt -le $max_attempts ]; do
    if curl -f http://localhost:9000/minio/health/live > /dev/null 2>&1; then
        echo "✅ MinIO is ready!"
        break
    fi
    
    if [ $attempt -eq $max_attempts ]; then
        echo "❌ MinIO failed to start after $max_attempts attempts"
        docker-compose logs minio
        exit 1
    fi
    
    echo "⏳ Attempt $attempt/$max_attempts - MinIO not ready yet..."
    sleep 2
    ((attempt++))
done

# Initialize MinIO bucket
echo "🪣 Initializing MinIO bucket..."
docker-compose exec -T minio mc alias set local http://localhost:9000 admin password123 > /dev/null 2>&1
docker-compose exec -T minio mc mb local/nfse-storage > /dev/null 2>&1 || echo "Bucket already exists"

# Build the service
echo "🔨 Building ZoomXML service..."
go build -o zoomxml-service cmd/zoomxml/main.go

if [ $? -ne 0 ]; then
    echo "❌ Failed to build service"
    exit 1
fi

echo "✅ Service built successfully"

# Set environment variables
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=nfse_metadata
export DB_SSLMODE=disable
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=admin
export MINIO_SECRET_KEY=password123
export MINIO_BUCKET=nfse-storage
export JWT_SECRET=your-secret-key-change-in-production
export PORT=8080

echo ""
echo "🎯 Infrastructure Ready!"
echo "======================="
echo ""
echo "📊 Database (PostgreSQL):"
echo "  Host: localhost:5432"
echo "  Database: nfse_metadata"
echo "  User: postgres"
echo "  Password: password"
echo ""
echo "🗄️ Storage (MinIO S3):"
echo "  API: http://localhost:9000"
echo "  Console: http://localhost:9001"
echo "  Access Key: admin"
echo "  Secret Key: password123"
echo "  Bucket: nfse-storage"
echo ""
echo "🌐 Database Admin (Adminer):"
echo "  URL: http://localhost:8080"
echo ""
echo "🚀 Starting ZoomXML Service..."
echo "=============================="

# Start the service
./zoomxml-service
