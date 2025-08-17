#!/bin/bash

# Setup script for NFS-e Multi-Enterprise System
echo "🚀 Setting up NFS-e Multi-Enterprise System"
echo "==========================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose not found. Please install docker-compose."
    exit 1
fi

echo "📦 Starting services with Docker Compose..."
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
        echo "📋 Checking logs..."
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
        echo "📋 Checking logs..."
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

echo ""
echo "🎯 System Setup Complete!"
echo "========================"
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
echo "  System: PostgreSQL"
echo "  Server: postgres"
echo "  Username: postgres"
echo "  Password: password"
echo "  Database: nfse_metadata"
echo ""
echo "🔧 Available Commands:"
echo "  go run cmd/api/main.go           # Start API server"
echo "  go run . fetch                   # Fetch NFS-e from API"
echo "  go run . organize                # Organize existing XMLs"
echo ""
echo "📚 API Documentation:"
echo "  Swagger UI: http://localhost:8080/swagger/"
echo ""
echo "🛑 To stop the system:"
echo "  docker-compose down"
echo ""
echo "🔄 To reset all data:"
echo "  docker-compose down -v"
echo ""
