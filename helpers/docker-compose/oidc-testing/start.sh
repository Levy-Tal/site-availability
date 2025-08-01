#!/bin/bash

# OIDC Testing Environment Startup Script
set -e

echo "🚀 Starting Site Availability OIDC Testing Environment"
echo "=================================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "docker-compose.yaml" ]; then
    echo "❌ docker-compose.yaml not found. Please run this script from the oidc-testing directory."
    exit 1
fi

echo "🏗️  Building and starting services..."
docker compose up --build -d

echo "⏳ Waiting for services to be ready..."
echo "   This may take 2-3 minutes..."

# Wait for Keycloak to be healthy
echo "   Waiting for Keycloak..."
timeout=180
counter=0
while [ $counter -lt $timeout ]; do
    if curl -s http://localhost:8090/health/ready > /dev/null 2>&1; then
        echo "   ✅ Keycloak is ready"
        break
    fi
    sleep 2
    counter=$((counter + 2))
    if [ $((counter % 20)) -eq 0 ]; then
        echo "   Still waiting for Keycloak... ($counter/${timeout}s)"
    fi
done

if [ $counter -ge $timeout ]; then
    echo "   ❌ Keycloak failed to start within $timeout seconds"
    echo "   Check logs with: docker compose logs keycloak"
    exit 1
fi

# Wait for Site Availability to be healthy
echo "   Waiting for Site Availability..."
counter=0
timeout=60
while [ $counter -lt $timeout ]; do
    if curl -s http://localhost:8080/healthz > /dev/null 2>&1; then
        echo "   ✅ Site Availability is ready"
        break
    fi
    sleep 2
    counter=$((counter + 2))
done

if [ $counter -ge $timeout ]; then
    echo "   ❌ Site Availability failed to start within $timeout seconds"
    echo "   Check logs with: docker compose logs site-availability"
    exit 1
fi

echo ""
echo "🎉 OIDC Testing Environment is ready!"
echo "=================================="
echo ""
echo "🌐 Access URLs:"
echo "   Site Availability:  http://localhost:8080"
echo "   Keycloak Admin:     http://localhost:8090 (admin/admin123)"
echo ""
echo "👥 Test Users:"
echo "   admin           admin123      Full access (devops group)"
echo "   frontend-user   frontend123   Frontend/backend access (developers group)"
echo "   backend-user    backend123    Frontend/backend access (developers group)"
echo "   qa-user         qa123         QA staging access (qa-team group)"
echo "   eu-user         eu123         EU regional access (eu-team group)"
echo ""
echo "🧪 Testing Steps:"
echo "   1. Open http://localhost:8080"
echo "   2. Click 'Login with Keycloak SSO'"
echo "   3. Try different users to see role-based filtering"
echo "   4. Test local admin fallback: admin/admin123"
echo ""
echo "📊 Monitor with:"
echo "   docker compose logs -f          # All logs"
echo "   docker compose ps               # Service status"
echo "   docker compose down             # Stop everything"
echo ""
echo "📖 See README.md for detailed testing scenarios and troubleshooting"
