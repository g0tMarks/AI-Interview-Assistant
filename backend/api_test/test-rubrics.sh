#!/bin/bash

# Test POST /rubrics endpoint

BASE_URL="http://localhost:8080"

echo "Testing POST /rubrics endpoint..."
echo ""

# Generate a test UUID for teacherId
TEACHER_ID="c6328c04-e891-40ab-9bec-c8df4915ba3a"

# Test 1: Valid request
echo "Test 1: Valid request"
RESPONSE=$(curl -X POST "${BASE_URL}/rubrics" \
  -H "Content-Type: application/json" \
  -d "{
    \"teacherId\": \"${TEACHER_ID}\",
    \"title\": \"Math Assessment Rubric\",
    \"description\": \"A rubric for assessing math skills\",
    \"rawText\": \"This is the raw text content of the rubric\"
  }" \
  -w "\n%{http_code}" \
  -s)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
echo "HTTP Status: $HTTP_CODE"

echo ""
echo "---"
echo ""

# Test 2: Missing title (should return 400)
echo "Test 2: Missing title (should return 400)"
RESPONSE=$(curl -X POST "${BASE_URL}/rubrics" \
  -H "Content-Type: application/json" \
  -d "{
    \"teacherId\": \"${TEACHER_ID}\",
    \"rawText\": \"Some text\"
  }" \
  -w "\n%{http_code}" \
  -s)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
echo "$BODY"
echo "HTTP Status: $HTTP_CODE"

echo ""
echo "---"
echo ""

# Test 3: Missing teacherId (should return 400)
echo "Test 3: Missing teacherId (should return 400)"
RESPONSE=$(curl -X POST "${BASE_URL}/rubrics" \
  -H "Content-Type: application/json" \
  -d "{
    \"title\": \"Test Rubric\",
    \"rawText\": \"Some text\"
  }" \
  -w "\n%{http_code}" \
  -s)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
echo "$BODY"
echo "HTTP Status: $HTTP_CODE"

echo ""
echo "---"
echo ""

# Test 4: Missing rawText (should return 400)
echo "Test 4: Missing rawText (should return 400)"
RESPONSE=$(curl -X POST "${BASE_URL}/rubrics" \
  -H "Content-Type: application/json" \
  -d "{
    \"teacherId\": \"${TEACHER_ID}\",
    \"title\": \"Test Rubric\"
  }" \
  -w "\n%{http_code}" \
  -s)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
echo "$BODY"
echo "HTTP Status: $HTTP_CODE"

echo ""
echo "Done!"


