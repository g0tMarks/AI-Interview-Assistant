#!/bin/bash

# Test POST /rubrics/upload (PDF/DOCX text extraction)

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

usage() {
  echo "Usage: $0 <path-to-pdf-or-docx> [teacher-id]"
  echo ""
  echo "  path-to-pdf-or-docx   Path to a PDF or DOCX file (required)"
  echo "  teacher-id            UUID of an existing teacher (optional; registers one if omitted)"
  echo ""
  echo "Example:"
  echo "  $0 ./my-rubric.pdf"
  echo "  $0 ./rubric.docx c6328c04-e891-40ab-9bec-c8df4915ba3a"
  exit 1
}

FILE_PATH="${1:-}"
TEACHER_ID_ARG="${2:-}"

if [ -z "$FILE_PATH" ]; then
  echo -e "${RED}Error: path to PDF or DOCX file is required${NC}"
  echo ""
  usage
fi

if [ ! -f "$FILE_PATH" ]; then
  echo -e "${RED}Error: file not found: $FILE_PATH${NC}"
  exit 1
fi

case "${FILE_PATH,,}" in
  *.pdf|*.docx|*.doc) ;;
  *)
    echo -e "${RED}Error: file must be .pdf, .docx, or .doc${NC}"
    exit 1
    ;;
esac

echo "=========================================="
echo "Testing rubric file upload (PDF/DOCX)"
echo "=========================================="
echo ""

HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health" || echo "000")
if [ "$HEALTH_CHECK" != "200" ]; then
  echo -e "${RED}✗ Server health check failed (got $HEALTH_CHECK)${NC}"
  echo "Start the API with: cd backend/cmd/api && go run main.go"
  exit 1
fi

if [ -z "$TEACHER_ID_ARG" ]; then
  echo -e "${YELLOW}Registering a teacher to get teacherId...${NC}"
  REG_RESP=$(curl -s -X POST "${BASE_URL}/teachers/register" \
    -H "Content-Type: application/json" \
    -d "{
      \"email\": \"rubric-upload-test-$(date +%s)@example.com\",
      \"fullName\": \"Rubric Upload Test\",
      \"password\": \"TestPassword123!\"
    }")
  TEACHER_ID=$(echo "$REG_RESP" | jq -r '.teacherId // empty')
  if [ -z "$TEACHER_ID" ]; then
    echo -e "${RED}✗ Failed to register teacher${NC}"
    echo "$REG_RESP" | jq '.' 2>/dev/null || echo "$REG_RESP"
    exit 1
  fi
  echo -e "${GREEN}Using teacherId: $TEACHER_ID${NC}"
else
  TEACHER_ID="$TEACHER_ID_ARG"
  echo "Using teacherId: $TEACHER_ID"
fi

echo -e "${YELLOW}POST /rubrics/upload (file=$(basename "$FILE_PATH"))${NC}"
RESP=$(curl -s -X POST "${BASE_URL}/rubrics/upload" \
  -F "file=@${FILE_PATH}" \
  -F "teacherId=${TEACHER_ID}" \
  -F "title=Uploaded rubric test" \
  -F "description=Test via script" \
  -w "\n%{http_code}")
HTTP_CODE=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

echo "Response:"
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
echo ""
echo "HTTP Status: $HTTP_CODE"

if [ "$HTTP_CODE" = "201" ]; then
  RAW_LEN=$(echo "$BODY" | jq -r '.rawText | length' 2>/dev/null || echo "?")
  echo ""
  echo -e "${GREEN}✓ Rubric created${NC}"
  echo "  Extracted rawText length: $RAW_LEN characters"
  if command -v jq >/dev/null 2>&1; then
    PREVIEW=$(echo "$BODY" | jq -r '.rawText | if length > 200 then .[0:200] + "..." else . end' 2>/dev/null)
    if [ -n "$PREVIEW" ] && [ "$PREVIEW" != "null" ]; then
      echo "  Preview: $PREVIEW"
    fi
  fi
  echo ""
  echo -e "${GREEN}Test completed!${NC}"
else
  echo ""
  echo -e "${RED}✗ Upload failed (HTTP $HTTP_CODE)${NC}"
  exit 1
fi
