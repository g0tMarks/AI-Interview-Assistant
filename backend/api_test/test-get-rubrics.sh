#!/bin/bash

# Test GET /rubrics endpoint
# Based on PRD: 3. PRD-GET-rubrics.md

BASE_URL="http://localhost:8080"
ENDPOINT="${BASE_URL}/rubrics"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter for test results
PASSED=0
FAILED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local expected_status="$2"
    local query_params="$3"
    local description="$4"
    
    echo -e "${YELLOW}Test: ${test_name}${NC}"
    if [ -n "$description" ]; then
        echo "Description: $description"
    fi
    echo ""
    
    RESPONSE=$(curl -X GET "${ENDPOINT}${query_params}" \
        -H "Content-Type: application/json" \
        -w "\n%{http_code}" \
        -s)
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    # Pretty print JSON if jq is available
    if command -v jq &> /dev/null; then
        echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    else
        echo "$BODY"
    fi
    
    echo "HTTP Status: $HTTP_CODE"
    
    if [ "$HTTP_CODE" == "$expected_status" ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED (Expected $expected_status, got $HTTP_CODE)${NC}"
        ((FAILED++))
    fi
    
    echo ""
    echo "---"
    echo ""
}

# Test teacher ID (should exist from create-test-teacher.sql)
TEACHER_ID="c6328c04-e891-40ab-9bec-c8df4915ba3a"


echo "=========================================="
echo "Testing GET /rubrics endpoint"
echo "=========================================="
echo ""

# Test 1: Valid request with existing teacherId
run_test \
    "Valid request with existing teacherId" \
    "200" \
    "?teacherId=${TEACHER_ID}" \
    "Should return 200 with array of rubrics"

# Test 2: Valid request with non-existent teacherId
run_test \
    "Valid request with non-existent teacherId" \
    "200" \
    "?teacherId=00000000-0000-0000-0000-000000000000" \
    "Should return 200 with empty array"

# Test 3: Missing teacherId query parameter
run_test \
    "Missing teacherId query parameter" \
    "400" \
    "" \
    "Should return 400 Bad Request"

# Test 4: Invalid UUID format
run_test \
    "Invalid UUID format" \
    "400" \
    "?teacherId=invalid-uuid" \
    "Should return 400 Bad Request"

# Test 5: Empty teacherId query parameter
run_test \
    "Empty teacherId query parameter" \
    "400" \
    "?teacherId=" \
    "Should return 400 Bad Request"

echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${GREEN}Passed: ${PASSED}${NC}"
echo -e "${RED}Failed: ${FAILED}${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi

