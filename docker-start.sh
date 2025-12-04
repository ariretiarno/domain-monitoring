#!/bin/bash
# Docker startup script for Domain Expiration Monitor

set -e

echo "🚀 Starting Domain Expiration Monitor with Docker..."
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "⚠️  No .env file found. Creating from .env.example..."
    cp .env.example .env
    echo "✅ Created .env file. Please edit it with your configuration."
    echo ""
    echo "Important: Set these variables in .env:"
    echo "  - MYSQL_ROOT_PASSWORD"
    echo "  - MYSQL_PASSWORD"
    echo "  - GOOGLE_CHAT_WEBHOOK (optional)"
    echo ""
    read -p "Press Enter to continue or Ctrl+C to exit and edit .env..."
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed. Please install it and try again."
    exit 1
fi

echo "📦 Building and starting services..."
docker-compose up -d --build

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 5

# Check if services are running
if docker-compose ps | grep -q "Up"; then
    echo ""
    echo "✅ Services started successfully!"
    echo ""
    echo "📊 Service Status:"
    docker-compose ps
    echo ""
    echo "🌐 Application URL: http://localhost:8080"
    echo ""
    echo "📝 Useful commands:"
    echo "  View logs:        docker-compose logs -f"
    echo "  Stop services:    docker-compose stop"
    echo "  Restart:          docker-compose restart"
    echo "  Remove all:       docker-compose down -v"
    echo ""
    echo "📖 For more information, see DOCKER_DEPLOYMENT.md"
else
    echo ""
    echo "❌ Services failed to start. Check logs:"
    echo "  docker-compose logs"
    exit 1
fi
