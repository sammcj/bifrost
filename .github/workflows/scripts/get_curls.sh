#!/bin/bash
set -uo pipefail

# Bifrost HTTP Transport - GET API Endpoints
# This script tests all GET endpoints in parallel and reports their status

# Base URL (update as needed)
BASE_URL="${BASE_URL:-http://localhost:8080}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo "Bifrost GET API Endpoints - Status Check"
echo "========================================"
echo "Base URL: $BASE_URL"
echo ""

# All endpoints to test
ENDPOINTS=(
  "/health"
  "/api/session/is-auth-enabled"
  "/api/plugins"
  "/api/plugins/telemetry"
  "/api/mcp/clients"
  "/api/logs?limit=10&offset=0&sort_by=timestamp&order=desc"
  "/api/logs/dropped"
  "/api/logs/filterdata"
  "/api/providers"
  "/api/providers/openai"
  "/api/keys"
  "/api/governance/virtual-keys"
  "/api/governance/virtual-keys/vk-123"
  "/api/governance/teams"
  "/api/governance/teams/team-123"
  "/api/governance/customers"
  "/api/governance/customers/cust-123"
  "/api/config"
  "/api/config?from_db=true"
  "/api/version"
  "/v1/models"
)

TOTAL_TESTS=${#ENDPOINTS[@]}

# Create a temporary file to store results
RESULTS_FILE=$(mktemp)

# Function to test a single endpoint and write result to file
test_endpoint() {
  local path=$1
  local status=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL$path" -H "Content-Type: application/json")
  
  if [ "$status" -ge 200 ] && [ "$status" -lt 300 ]; then
    echo "PASS:$path:$status" >> "$RESULTS_FILE"
  else
    echo "FAIL:$path:$status" >> "$RESULTS_FILE"
  fi
}

# Export function and variables for parallel execution
export -f test_endpoint
export BASE_URL
export RESULTS_FILE

# Run all endpoint tests in parallel (max 10 concurrent)
printf '%s\n' "${ENDPOINTS[@]}" | xargs -P 10 -I {} bash -c 'test_endpoint "$@"' _ {}

# Process results
FAILED_TESTS=0
while IFS=: read -r result path status; do
  if [ "$result" = "PASS" ]; then
    echo -e "GET $path - ${GREEN}✓ SUCCESS${NC} ($status)"
  else
    echo -e "GET $path - ${RED}✗ FAILURE${NC} ($status)"
    FAILED_TESTS=$((FAILED_TESTS + 1))
  fi
done < "$RESULTS_FILE"

# Clean up
rm -f "$RESULTS_FILE"

echo ""
echo -e "${YELLOW}Note: WebSocket endpoint (/ws) requires a WebSocket client${NC}"
echo ""
echo "========================================"
echo "Test Summary:"
echo "  Total tests: $TOTAL_TESTS"
echo "  Passed: $((TOTAL_TESTS - FAILED_TESTS))"
echo "  Failed: $FAILED_TESTS"
echo "========================================"

echo "The aim of the script is to make sure bifrost server is not crashing"

# Exit with failure if any tests failed
if [ $FAILED_TESTS -gt 0 ]; then
  exit 1
fi
