#!/bin/bash

# Test POST /classes/{id}/roster/upload

set -euo pipefail

BASE_URL="http://localhost:8080"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "Testing bulk roster upload"
echo "=========================================="
echo ""

# Check if server is running
HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health" || echo "000")
if [ "$HEALTH_CHECK" != "200" ]; then
  echo -e "${RED}✗ Server health check failed (got $HEALTH_CHECK)${NC}"
  echo "Please ensure the API server is running on ${BASE_URL}"
  echo "Start it with: cd backend/cmd/api && go run main.go"
  exit 1
fi

# Get class ID from command line or prompt
CLASS_ID="${1:-}"

if [ -z "$CLASS_ID" ]; then
  echo -e "${YELLOW}No class ID provided. Listing classes...${NC}"
  CLASSES_RESP=$(curl -s "${BASE_URL}/classes")
  echo "$CLASSES_RESP" | jq '.' 2>/dev/null || echo "$CLASSES_RESP"
  echo ""
  echo -e "${YELLOW}Please provide a class ID as the first argument:${NC}"
  echo "Usage: $0 <class-id> [path-to-file.xlsx]"
  exit 1
fi

# Get file path from command line or use default
FILE_PATH="${2:-}"

if [ -z "$FILE_PATH" ]; then
  echo -e "${YELLOW}No file path provided.${NC}"
  echo "Usage: $0 <class-id> <path-to-file.xlsx>"
  echo ""
  echo "Example:"
  echo "  $0 123e4567-e89b-12d3-a456-426614174000 ./students.xlsx"
  exit 1
fi

if [ ! -f "$FILE_PATH" ]; then
  echo -e "${RED}✗ File not found: $FILE_PATH${NC}"
  exit 1
fi

if [[ ! "$FILE_PATH" =~ \.(xlsx|XLSX)$ ]]; then
  echo -e "${RED}✗ File must be a .xlsx file${NC}"
  exit 1
fi

echo -e "${YELLOW}Uploading roster for class: $CLASS_ID${NC}"
echo -e "${YELLOW}File: $FILE_PATH${NC}"
echo ""

UPLOAD_RESP=$(curl -s -X POST "${BASE_URL}/classes/${CLASS_ID}/roster/upload" \
  -F "file=@${FILE_PATH}" \
  -w "\n%{http_code}")

HTTP_CODE=$(echo "$UPLOAD_RESP" | tail -n1)
BODY=$(echo "$UPLOAD_RESP" | sed '$d')

echo "Response:"
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
echo ""
echo "HTTP Status: $HTTP_CODE"

if [ "$HTTP_CODE" != "200" ]; then
  echo -e "${RED}✗ Expected 200${NC}"
  exit 1
fi

CREATED=$(echo "$BODY" | jq -r '.createdCount' 2>/dev/null || echo "0")
ADDED=$(echo "$BODY" | jq -r '.addedToRosterCount' 2>/dev/null || echo "0")
SKIPPED=$(echo "$BODY" | jq -r '.skippedCount' 2>/dev/null || echo "0")
ERRORS=$(echo "$BODY" | jq -r '.errorCount' 2>/dev/null || echo "0")

echo ""
echo -e "${GREEN}✓ Upload successful!${NC}"
echo "  Created: $CREATED students"
echo "  Added to roster: $ADDED students"
echo "  Skipped (already in roster): $SKIPPED students"
echo "  Errors: $ERRORS rows"

if [ "$ERRORS" -gt 0 ]; then
  echo ""
  echo -e "${YELLOW}Errors:${NC}"
  echo "$BODY" | jq -r '.errors[]?' 2>/dev/null || true
fi

echo ""
echo -e "${GREEN}Test completed!${NC}"
