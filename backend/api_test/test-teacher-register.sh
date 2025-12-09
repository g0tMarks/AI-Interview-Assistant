#!/bin/bash

# Test POST /teachers/register endpoint
# Based on PRD: 2. PRD-API-Endpoints-CreateTeacherAccount.md

BASE_URL="http://localhost:8080"
ENDPOINT="${BASE_URL}/teachers/register"

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
    else
        echo -e "${RED}✗ FAILED (expected $expected_status, got $HTTP_CODE)${NC}"
        ((FAILED++))
    fi
    
    echo ""
    echo "---"
    echo ""
}

echo "=========================================="
echo "Testing POST /teachers/register endpoint"
echo "=========================================="
echo ""

# Pre-flight check: Verify server is running and endpoint exists
echo "Pre-flight check: Verifying server and endpoint..."
HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health")
if [ "$HEALTH_CHECK" != "200" ]; then
    echo -e "${RED}✗ Server health check failed (got $HEALTH_CHECK)${NC}"
    echo "Please ensure the API server is running on ${BASE_URL}"
    exit 1
fi

# Check if endpoint exists by making a test request
ENDPOINT_CHECK=$(curl -X POST "${ENDPOINT}" \
    -H "Content-Type: application/json" \
    -d '{"email":"test@test.com","fullName":"Test","password":"Test123!"}' \
    -s -o /dev/null -w "%{http_code}")

if [ "$ENDPOINT_CHECK" == "404" ]; then
    echo -e "${RED}✗ Endpoint not found (404)${NC}"
    echo "The /teachers/register endpoint is not available."
    echo "Please ensure:"
    echo "  1. The server has been rebuilt with the latest code"
    echo "  2. The server has been restarted"
    echo "  3. The route is properly registered in router.go"
    exit 1
fi

echo -e "${GREEN}✓ Server is running and endpoint is accessible${NC}"
echo ""

# Success Cases
echo "=== SUCCESS CASES ==="
echo ""

# Test 1: Valid request with all required fields
run_test \
    "1. Valid request" \
    "201" \
    '{
        "email": "teacher1@example.com",
        "fullName": "John Doe",
        "password": "SecurePass123!"
    }' \
    "Should return 201 with teacher data (excluding passwordHash)"

# Test 2: Valid request with different email
run_test \
    "2. Valid request (different email)" \
    "201" \
    '{
        "email": "teacher2@example.com",
        "fullName": "Jane Smith",
        "password": "MyP@ssw0rd"
    }' \
    "Should return 201 with teacher data"

# Validation Error Cases (400 Bad Request)
echo "=== VALIDATION ERROR CASES (400 Bad Request) ==="
echo ""

# Test 3: Missing email
run_test \
    "3. Missing email" \
    "400" \
    '{
        "fullName": "John Doe",
        "password": "SecurePass123!"
    }' \
    "Should return 400 - email is required"

# Test 4: Missing fullName
run_test \
    "4. Missing fullName" \
    "400" \
    '{
        "email": "test@example.com",
        "password": "SecurePass123!"
    }' \
    "Should return 400 - fullName is required"

# Test 5: Missing password
run_test \
    "5. Missing password" \
    "400" \
    '{
        "email": "test@example.com",
        "fullName": "John Doe"
    }' \
    "Should return 400 - password is required"

# Test 6: Invalid email format
run_test \
    "6. Invalid email format" \
    "400" \
    '{
        "email": "not-an-email",
        "fullName": "John Doe",
        "password": "SecurePass123!"
    }' \
    "Should return 400 - invalid email format"

# Test 7: Empty email
run_test \
    "7. Empty email" \
    "400" \
    '{
        "email": "",
        "fullName": "John Doe",
        "password": "SecurePass123!"
    }' \
    "Should return 400 - email cannot be empty"

# Test 8: Empty fullName (after trimming)
run_test \
    "8. Empty fullName (whitespace only)" \
    "400" \
    '{
        "email": "test@example.com",
        "fullName": "   ",
        "password": "SecurePass123!"
    }' \
    "Should return 400 - fullName cannot be empty"

# Test 9: Password too short
run_test \
    "9. Password too short (< 8 characters)" \
    "400" \
    '{
        "email": "test@example.com",
        "fullName": "John Doe",
        "password": "Short1!"
    }' \
    "Should return 400 - password must be at least 8 characters"

# Test 10: Password without uppercase
run_test \
    "10. Password without uppercase letter" \
    "400" \
    '{
        "email": "test@example.com",
        "fullName": "John Doe",
        "password": "lowercase123!"
    }' \
    "Should return 400 - password must contain uppercase letter"

# Test 11: Password without lowercase
run_test \
    "11. Password without lowercase letter" \
    "400" \
    '{
        "email": "test@example.com",
        "fullName": "John Doe",
        "password": "UPPERCASE123!"
    }' \
    "Should return 400 - password must contain lowercase letter"

# Test 12: Password without number
run_test \
    "12. Password without number" \
    "400" \
    '{
        "email": "test@example.com",
        "fullName": "John Doe",
        "password": "NoNumber!"
    }' \
    "Should return 400 - password must contain number"

# Test 13: Password without special character
run_test \
    "13. Password without special character" \
    "400" \
    '{
        "email": "test@example.com",
        "fullName": "John Doe",
        "password": "NoSpecial123"
    }' \
    "Should return 400 - password must contain special character"

# Test 14: Invalid JSON body
run_test \
    "14. Invalid JSON body" \
    "400" \
    '{"email": "test@example.com", "fullName": "John Doe", "password": "SecurePass123!" invalid}' \
    "Should return 400 - invalid JSON"

# Conflict Error Cases (409 Conflict)
echo "=== CONFLICT ERROR CASES (409 Conflict) ==="
echo ""

# Test 15: Email already registered (use email from Test 1)
run_test \
    "15. Email already registered" \
    "409" \
    '{
        "email": "teacher1@example.com",
        "fullName": "Different Name",
        "password": "DifferentPass123!"
    }' \
    "Should return 409 - email already exists"

# Test 16: Duplicate email (different case - if applicable)
run_test \
    "16. Duplicate email (case variation)" \
    "409" \
    '{
        "email": "TEACHER1@EXAMPLE.COM",
        "fullName": "Another Name",
        "password": "AnotherPass123!"
    }' \
    "Should return 409 - email already exists (case-insensitive check)"

# Edge Cases
echo "=== EDGE CASES ==="
echo ""

# Test 17: Valid email with plus sign
run_test \
    "17. Valid email with plus sign" \
    "201" \
    '{
        "email": "teacher+test@example.com",
        "fullName": "Test User",
        "password": "ValidPass123!"
    }' \
    "Should return 201 - plus sign in email is valid"

# Test 18: Valid email with subdomain
run_test \
    "18. Valid email with subdomain" \
    "201" \
    '{
        "email": "teacher@mail.example.com",
        "fullName": "Subdomain User",
        "password": "ValidPass123!"
    }' \
    "Should return 201 - subdomain in email is valid"

# Test 19: Full name with spaces
run_test \
    "19. Full name with multiple words" \
    "201" \
    '{
        "email": "multiname@example.com",
        "fullName": "John Michael Smith",
        "password": "ValidPass123!"
    }' \
    "Should return 201 - multi-word names are valid"

# Test 20: Password with all required character types
run_test \
    "20. Complex password" \
    "201" \
    '{
        "email": "complexpass@example.com",
        "fullName": "Complex User",
        "password": "A1b2C3d4E5f6G7h8!@#"
    }' \
    "Should return 201 - complex password is valid"

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo "Total: $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi

