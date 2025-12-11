#!/bin/bash

# Test POST /interview-templates endpoint
# Based on PRD: 4. PRD-POST-interview-templates.md

BASE_URL="http://localhost:8080"
ENDPOINT="${BASE_URL}/interview-templates"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter for test results
PASSED=0
FAILED=0

# Test teacher ID and rubric ID (should exist)
TEACHER_ID="c6328c04-e891-40ab-9bec-c8df4915ba3a"

# Function to run a test
run_test() {
    local test_name="$1"
    local expected_status="$2"
    local json_data="$3"
    local description="$4"
    
    echo -e "${YELLOW}Test: ${test_name}${NC}"
    if [ -n "$description" ]; then
        echo "Description: $description"
    fi
    echo ""
    
    RESPONSE=$(curl -X POST "${ENDPOINT}" \
        -H "Content-Type: application/json" \
        -d "${json_data}" \
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
        
        # Extract interviewPlanId for use in subsequent tests
        if [ "$HTTP_CODE" == "201" ] && command -v jq &> /dev/null; then
            INTERVIEW_PLAN_ID=$(echo "$BODY" | jq -r '.interviewPlanId // empty')
            if [ -n "$INTERVIEW_PLAN_ID" ] && [ "$INTERVIEW_PLAN_ID" != "null" ]; then
                echo "$INTERVIEW_PLAN_ID" > /tmp/interview_plan_id.txt
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

# First, create a rubric to use for testing
echo "Creating a test rubric..."
RUBRIC_RESPONSE=$(curl -X POST "${BASE_URL}/rubrics" \
    -H "Content-Type: application/json" \
    -d "{
        \"teacherId\": \"${TEACHER_ID}\",
        \"title\": \"Test Rubric for Interview Templates\",
        \"description\": \"A test rubric\",
        \"rawText\": \"Test rubric content\"
    }" \
    -s)

if command -v jq &> /dev/null; then
    RUBRIC_ID=$(echo "$RUBRIC_RESPONSE" | jq -r '.rubricId // empty')
else
    # Fallback: extract UUID manually
    RUBRIC_ID=$(echo "$RUBRIC_RESPONSE" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | head -n1)
fi

if [ -z "$RUBRIC_ID" ] || [ "$RUBRIC_ID" == "null" ]; then
    echo -e "${RED}Failed to create test rubric. Exiting.${NC}"
    exit 1
fi

echo "Using rubric ID: $RUBRIC_ID"
echo ""

echo "=========================================="
echo "Testing POST /interview-templates endpoint"
echo "=========================================="
echo ""

# Test 1: Valid request with all fields
run_test \
    "Valid request with all fields" \
    "201" \
    "{
        \"rubricId\": \"${RUBRIC_ID}\",
        \"title\": \"Year 7 Algebra Diagnostic Interview\",
        \"instructions\": \"Conduct a diagnostic interview focusing on algebraic thinking.\",
        \"config\": {\"maxQuestions\": 10, \"timeLimit\": 30},
        \"status\": \"draft\",
        \"curriculumSubject\": \"Mathematics\",
        \"curriculumLevelBand\": \"7-8\"
    }" \
    "Should return 201 with interview template data"

# Test 2: Valid request with minimal fields (defaults)
run_test \
    "Valid request with minimal fields" \
    "201" \
    "{
        \"rubricId\": \"${RUBRIC_ID}\",
        \"title\": \"Minimal Interview Template\"
    }" \
    "Should return 201 with defaults (status=draft, config={})"

# Test 3: Missing rubricId
run_test \
    "Missing rubricId" \
    "400" \
    "{
        \"title\": \"Test Template\"
    }" \
    "Should return 400 Bad Request"

# Test 4: Missing title
run_test \
    "Missing title" \
    "400" \
    "{
        \"rubricId\": \"${RUBRIC_ID}\"
    }" \
    "Should return 400 Bad Request"

# Test 5: Empty title
run_test \
    "Empty title" \
    "400" \
    "{
        \"rubricId\": \"${RUBRIC_ID}\",
        \"title\": \"   \"
    }" \
    "Should return 400 Bad Request"

# Test 6: Invalid UUID format for rubricId
run_test \
    "Invalid UUID format for rubricId" \
    "400" \
    "{
        \"rubricId\": \"invalid-uuid\",
        \"title\": \"Test Template\"
    }" \
    "Should return 400 Bad Request"

# Test 7: Non-existent rubricId
run_test \
    "Non-existent rubricId" \
    "404" \
    "{
        \"rubricId\": \"00000000-0000-0000-0000-000000000000\",
        \"title\": \"Test Template\"
    }" \
    "Should return 404 Not Found"

# Test 8: Invalid status enum
run_test \
    "Invalid status enum" \
    "400" \
    "{
        \"rubricId\": \"${RUBRIC_ID}\",
        \"title\": \"Test Template\",
        \"status\": \"invalid_status\"
    }" \
    "Should return 400 Bad Request"

# Test 9: Valid status values
for status in "draft" "in_progress" "completed"; do
    run_test \
        "Valid status: ${status}" \
        "201" \
        "{
            \"rubricId\": \"${RUBRIC_ID}\",
            \"title\": \"Test Template ${status}\",
            \"status\": \"${status}\"
        }" \
        "Should return 201 with status=${status}"
done

# Test 10: Invalid config JSON
run_test \
    "Invalid config JSON" \
    "400" \
    "{
        \"rubricId\": \"${RUBRIC_ID}\",
        \"title\": \"Test Template\",
        \"config\": \"invalid json\"
    }" \
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

