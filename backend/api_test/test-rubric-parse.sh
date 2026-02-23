#!/bin/bash

# Test POST /rubrics/{id}/parse (LLM one-shot parse → criteria + question plan, validation, store)

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

usage() {
  echo "Usage: $0 <rubric-id>"
  echo ""
  echo "  rubric-id   UUID of an existing rubric that has rawText (e.g. from POST /rubrics/upload or POST /rubrics)"
  echo ""
  echo "Requires OPENAI_API_KEY or ANTHROPIC_API_KEY for the LLM parse."
  echo ""
  echo "Example:"
  echo "  1. Upload a rubric: ./test-rubric-upload.sh ./my-rubric.pdf"
  echo "  2. Parse it:       $0 <rubricId-from-step-1>"
  exit 1
}

RUBRIC_ID="${1:-}"

if [ -z "$RUBRIC_ID" ]; then
  echo -e "${RED}Error: rubric-id is required${NC}"
  echo ""
  usage
fi

echo "=========================================="
echo "Testing POST /rubrics/{id}/parse"
echo "=========================================="
echo "Rubric ID: $RUBRIC_ID"
echo ""

HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health" || echo "000")
if [ "$HEALTH_CHECK" != "200" ]; then
  echo -e "${RED}✗ Server health check failed (got $HEALTH_CHECK)${NC}"
  echo "Start the API with: cd backend/cmd/api && go run main.go"
  exit 1
fi

echo -e "${YELLOW}POST /rubrics/${RUBRIC_ID}/parse${NC}"
RESP=$(curl -s -X POST "${BASE_URL}/rubrics/${RUBRIC_ID}/parse" -w "\n%{http_code}")
HTTP_CODE=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

echo "Response:"
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
echo ""
echo "HTTP Status: $HTTP_CODE"

if [ "$HTTP_CODE" = "200" ]; then
  CRITERIA_COUNT=$(echo "$BODY" | jq -r '.criteriaCount // empty' 2>/dev/null)
  QUESTION_COUNT=$(echo "$BODY" | jq -r '.questionCount // empty' 2>/dev/null)
  PLAN_ID=$(echo "$BODY" | jq -r '.interviewPlanId // empty' 2>/dev/null)
  echo ""
  echo -e "${GREEN}✓ Parse completed${NC}"
  echo "  Criteria created: $CRITERIA_COUNT"
  echo "  Questions created: $QUESTION_COUNT"
  echo "  Interview plan ID: $PLAN_ID"
  echo ""
  echo -e "${GREEN}Test completed!${NC}"
else
  echo ""
  echo -e "${RED}✗ Parse failed (HTTP $HTTP_CODE)${NC}"
  echo "  If 503: set OPENAI_API_KEY or ANTHROPIC_API_KEY"
  echo "  If 422: LLM returned invalid or empty criteria/questions; check rubric rawText"
  exit 1
fi
