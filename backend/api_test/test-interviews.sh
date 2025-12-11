#!/bin/bash

# Test POST /interviews and GET /interviews/{id} endpoints
# Based on PRD: 5. PRD-POST-interviews.md and 6. PRD-GET-interviews-id.md

BASE_URL="http://localhost:8080"
POST_ENDPOINT="${BASE_URL}/interviews"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter for test results
PASSED=0
FAILED=0

# Test teacher ID (should exist)
TEACHER_ID="c6328c04-e891-40ab-9bec-c8df4915ba3a"

# Function to run a test
run_test() {
    local test_name="$1"
    local expected_status="$2"
    local method="$3"
    local url="$4"
    local json_data="$5"
    local description="$6"
    
    echo -e "${YELLOW}Test: ${test_name}${NC}"
    if [ -n "$description" ]; then
        echo "Description: $description"
    fi
    echo ""
    
    if [ -n "$json_data" ]; then
        RESPONSE=$(curl -X "${method}" "${url}" \
            -H "Content-Type: application/json" \
            -d "${json_data}" \
            -w "\n%{http_code}" \
            -s)
    else
        RESPONSE=$(curl -X "${method}" "${url}" \
            -H "Content-Type: application/json" \
            -w "\n%{http_code}" \
            -s)
    fi
    
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
        
        # Extract interviewId for use in GET tests
        if [ "$HTTP_CODE" == "201" ] && [ "$method" == "POST" ] && command -v jq &> /dev/null; then
            INTERVIEW_ID=$(echo "$BODY" | jq -r '.interviewId // empty')
            if [ -n "$INTERVIEW_ID" ] && [ "$INTERVIEW_ID" != "null" ]; then
                echo "$INTERVIEW_ID" > /tmp/interview_id.txt
            fi
        fi
    else
        echo -e "${RED}✗ FAILED (Expected $expected_status, got $HTTP_CODE)${NC}"
        ((FAILED++))
    fi
    
    echo ""
    echo "---"
    echo ""
}

# Setup: Create a rubric and interview template
echo "Setting up test data..."
echo ""

# Create a rubric
echo "Creating test rubric..."
RUBRIC_RESPONSE=$(curl -X POST "${BASE_URL}/rubrics" \
    -H "Content-Type: application/json" \
    -d "{
        \"teacherId\": \"${TEACHER_ID}\",
        \"title\": \"Test Rubric for Interviews\",
        \"description\": \"A test rubric\",
        \"rawText\": \"Test rubric content\"
    }" \
    -s)

if command -v jq &> /dev/null; then
    RUBRIC_ID=$(echo "$RUBRIC_RESPONSE" | jq -r '.rubricId // empty')
else
    RUBRIC_ID=$(echo "$RUBRIC_RESPONSE" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | head -n1)
fi

if [ -z "$RUBRIC_ID" ] || [ "$RUBRIC_ID" == "null" ]; then
    echo -e "${RED}Failed to create test rubric. Exiting.${NC}"
    exit 1
fi

echo "Created rubric ID: $RUBRIC_ID"

# Create an interview template
echo "Creating test interview template..."
TEMPLATE_RESPONSE=$(curl -X POST "${BASE_URL}/interview-templates" \
    -H "Content-Type: application/json" \
    -d "{
        \"rubricId\": \"${RUBRIC_ID}\",
        \"title\": \"Test Interview Template\",
        \"status\": \"draft\"
    }" \
    -s)

if command -v jq &> /dev/null; then
    INTERVIEW_PLAN_ID=$(echo "$TEMPLATE_RESPONSE" | jq -r '.interviewPlanId // empty')
else
    INTERVIEW_PLAN_ID=$(echo "$TEMPLATE_RESPONSE" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | head -n1)
fi

if [ -z "$INTERVIEW_PLAN_ID" ] || [ "$INTERVIEW_PLAN_ID" == "null" ]; then
    echo -e "${RED}Failed to create test interview template. Exiting.${NC}"
    exit 1
fi

echo "Created interview plan ID: $INTERVIEW_PLAN_ID"
echo ""

echo "=========================================="
echo "Testing POST /interviews endpoint"
echo "=========================================="
echo ""

# Test 1: Valid request with all fields
run_test \
    "Valid request with all fields" \
    "201" \
    "POST" \
    "${POST_ENDPOINT}" \
    "{
        \"interviewPlanId\": \"${INTERVIEW_PLAN_ID}\",
        \"teacherId\": \"${TEACHER_ID}\",
        \"simulated\": true,
        \"studentName\": \"Test Student\",
        \"status\": \"in_progress\"
    }" \
    "Should return 201 with interview data"

# Test 2: Valid request with minimal fields (defaults)
run_test \
    "Valid request with minimal fields" \
    "201" \
    "POST" \
    "${POST_ENDPOINT}" \
    "{
        \"interviewPlanId\": \"${INTERVIEW_PLAN_ID}\"
    }" \
    "Should return 201 with defaults (simulated=true, status=in_progress)"

# Test 3: Missing interviewPlanId
run_test \
    "Missing interviewPlanId" \
    "400" \
    "POST" \
    "${POST_ENDPOINT}" \
    "{
        \"teacherId\": \"${TEACHER_ID}\"
    }" \
    "Should return 400 Bad Request"

# Test 4: Invalid UUID format for interviewPlanId
run_test \
    "Invalid UUID format for interviewPlanId" \
    "400" \
    "POST" \
    "${POST_ENDPOINT}" \
    "{
        \"interviewPlanId\": \"invalid-uuid\"
    }" \
    "Should return 400 Bad Request"

# Test 5: Non-existent interviewPlanId
run_test \
    "Non-existent interviewPlanId" \
    "404" \
    "POST" \
    "${POST_ENDPOINT}" \
    "{
        \"interviewPlanId\": \"00000000-0000-0000-0000-000000000000\"
    }" \
    "Should return 404 Not Found"

# Test 6: Invalid status enum
run_test \
    "Invalid status enum" \
    "400" \
    "POST" \
    "${POST_ENDPOINT}" \
    "{
        \"interviewPlanId\": \"${INTERVIEW_PLAN_ID}\",
        \"status\": \"invalid_status\"
    }" \
    "Should return 400 Bad Request"

# Test 7: Valid status values
for status in "draft" "in_progress" "completed"; do
    run_test \
        "Valid status: ${status}" \
        "201" \
        "POST" \
        "${POST_ENDPOINT}" \
        "{
            \"interviewPlanId\": \"${INTERVIEW_PLAN_ID}\",
            \"status\": \"${status}\"
        }" \
        "Should return 201 with status=${status}"
done

# Test 8: Simulated false
run_test \
    "Simulated false" \
    "201" \
    "POST" \
    "${POST_ENDPOINT}" \
    "{
        \"interviewPlanId\": \"${INTERVIEW_PLAN_ID}\",
        \"simulated\": false
    }" \
    "Should return 201 with simulated=false"

# Get the interview ID from the last successful creation
if [ -f /tmp/interview_id.txt ]; then
    INTERVIEW_ID=$(cat /tmp/interview_id.txt)
    echo "Using interview ID: $INTERVIEW_ID"
    echo ""
    
    echo "=========================================="
    echo "Testing GET /interviews/{id} endpoint"
    echo "=========================================="
    echo ""
    
    # Test 9: Valid GET request
    run_test \
        "Valid GET request" \
        "200" \
        "GET" \
        "${BASE_URL}/interviews/${INTERVIEW_ID}" \
        "" \
        "Should return 200 with interview data"
    
    # Test 10: Non-existent interviewId
    run_test \
        "Non-existent interviewId" \
        "404" \
        "GET" \
        "${BASE_URL}/interviews/00000000-0000-0000-0000-000000000000" \
        "" \
        "Should return 404 Not Found"
    
    # Test 11: Invalid UUID format
    run_test \
        "Invalid UUID format" \
        "400" \
        "GET" \
        "${BASE_URL}/interviews/invalid-uuid" \
        "" \
        "Should return 400 Bad Request"
    
    # Test 12: Missing ID in path
    run_test \
        "Missing ID in path" \
        "404" \
        "GET" \
        "${BASE_URL}/interviews/" \
        "" \
        "Should return 404 (route not found)"
else
    echo -e "${YELLOW}Warning: No interview ID available for GET tests${NC}"
fi

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

