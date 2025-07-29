#!/bin/bash

# Site Availability Authentication & Authorization Test Script
# This script tests all authentication and authorization features

BASE_URL="http://localhost:8080"
COOKIE_JAR="test-cookies.txt"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_test() {
    echo -e "${BLUE}🧪 TEST: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_section() {
    echo ""
    echo -e "${YELLOW}=== $1 ===${NC}"
    echo ""
}

# Function to make API calls and show results
test_api_call() {
    local method=$1
    local endpoint=$2
    local description=$3
    local expect_auth_required=${4:-false}
    local use_cookies=${5:-true}

    print_test "$description"

    local curl_args="-s -w 'HTTP Status: %{http_code}\n'"
    if [ "$use_cookies" = true ]; then
        curl_args="$curl_args -b $COOKIE_JAR"
    fi

    local response
    response=$(eval "curl $curl_args -X $method '$BASE_URL$endpoint'")
    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        echo "$response" | head -20  # Show first 20 lines
        local status_code=$(echo "$response" | grep "HTTP Status:" | awk '{print $3}')

        if [ "$expect_auth_required" = true ] && [ "$status_code" = "401" ]; then
            print_success "Correctly requires authentication"
        elif [ "$expect_auth_required" = false ] && [ "$status_code" = "200" ]; then
            print_success "API call successful"
        elif [ "$status_code" = "200" ]; then
            print_success "API call successful"
        else
            print_warning "Unexpected status code: $status_code"
        fi
    else
        print_error "API call failed"
    fi
    echo ""
}

# Function to login
login() {
    local username=$1
    local password=$2

    print_test "Logging in as $username"

    local response
    response=$(curl -s -c $COOKIE_JAR -w 'HTTP Status: %{http_code}\n' \
        -X POST \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}" \
        "$BASE_URL/auth/login")

    local status_code=$(echo "$response" | grep "HTTP Status:" | awk '{print $3}')

    if [ "$status_code" = "200" ]; then
        print_success "Login successful for $username"
        echo "$response" | grep -v "HTTP Status:"
    else
        print_error "Login failed for $username"
        echo "$response"
    fi
    echo ""
}

# Function to logout
logout() {
    print_test "Logging out"

    local response
    response=$(curl -s -b $COOKIE_JAR -c $COOKIE_JAR -w 'HTTP Status: %{http_code}\n' \
        -X POST \
        "$BASE_URL/auth/logout")

    local status_code=$(echo "$response" | grep "HTTP Status:" | awk '{print $3}')

    if [ "$status_code" = "200" ]; then
        print_success "Logout successful"
    else
        print_error "Logout failed"
    fi
    echo ""
}

# Function to check user info
check_user_info() {
    print_test "Checking user information"

    local response
    response=$(curl -s -b $COOKIE_JAR -w 'HTTP Status: %{http_code}\n' \
        "$BASE_URL/auth/user")

    local status_code=$(echo "$response" | grep "HTTP Status:" | awk '{print $3}')

    if [ "$status_code" = "200" ]; then
        print_success "User info retrieved"
        echo "$response" | grep -v "HTTP Status:" | jq . 2>/dev/null || echo "$response" | grep -v "HTTP Status:"
    else
        print_warning "Could not get user info (status: $status_code)"
    fi
    echo ""
}

# Main test execution
main() {
    print_section "Site Availability Authentication & Authorization Tests"

    # Clean up any existing cookies
    rm -f $COOKIE_JAR

    print_section "Phase 1: Authentication Configuration Check"
    test_api_call "GET" "/auth/config" "Check authentication configuration" false false

    print_section "Phase 2: Unauthenticated API Access Tests"
    test_api_call "GET" "/api/labels" "Get labels without authentication" true false
    test_api_call "GET" "/api/apps" "Get apps without authentication" true false
    test_api_call "GET" "/api/locations" "Get locations without authentication" true false

    print_section "Phase 3: Authentication Tests"

    # Test invalid login
    print_test "Testing invalid login credentials"
    curl -s -w 'HTTP Status: %{http_code}\n' \
        -X POST \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"wrongpass"}' \
        "$BASE_URL/auth/login" | grep "HTTP Status: 401" > /dev/null

    if [ $? -eq 0 ]; then
        print_success "Invalid credentials correctly rejected"
    else
        print_error "Invalid credentials should be rejected"
    fi
    echo ""

    # Test valid login
    login "admin" "testpass123"

    # Check user info after login
    check_user_info

    print_section "Phase 4: Authenticated API Access Tests (Admin User)"
    test_api_call "GET" "/api/labels" "Get all labels as admin"
    test_api_call "GET" "/api/labels?team" "Get team label values as admin"
    test_api_call "GET" "/api/labels?env" "Get env label values as admin"
    test_api_call "GET" "/api/apps" "Get all apps as admin"
    test_api_call "GET" "/api/locations" "Get all locations as admin"
    test_api_call "GET" "/api/docs" "Get documentation as admin"
    test_api_call "GET" "/api/scrape-interval" "Get scrape interval as admin"

    print_section "Phase 5: Label Filtering Tests"
    test_api_call "GET" "/api/apps?labels.team=frontend" "Filter apps by team=frontend"
    test_api_call "GET" "/api/apps?labels.env=production" "Filter apps by env=production"
    test_api_call "GET" "/api/apps?labels.region=us-east" "Filter apps by region=us-east"
    test_api_call "GET" "/api/locations?labels.env=production" "Filter locations by env=production"

    print_section "Phase 6: Logout Test"
    logout

    # Test that we can't access protected endpoints after logout
    test_api_call "GET" "/api/labels" "Get labels after logout" true

    print_section "Test Summary"
    print_success "All authentication and authorization tests completed!"
    print_warning "Note: This tested admin user (full access). Role-based tests require:"
    print_warning "  1. OIDC integration, OR"
    print_warning "  2. Modifying user session to simulate different roles"

    # Clean up
    rm -f $COOKIE_JAR
}

# Check if jq is available for JSON formatting
if ! command -v jq &> /dev/null; then
    print_warning "jq not found - JSON responses will not be formatted"
fi

# Run the tests
main
