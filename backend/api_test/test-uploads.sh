#!/bin/bash

# Test POST /uploads and GET /uploads/{key}

set -euo pipefail

BASE_URL="http://localhost:8080"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "Testing uploads"
echo "=========================================="
echo ""

HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health")
if [ "$HEALTH_CHECK" != "200" ]; then
  echo -e "${RED}✗ Server health check failed (got $HEALTH_CHECK)${NC}"
  echo "Please ensure the API server is running on ${BASE_URL}"
  exit 1
fi

TMP_DIR=$(mktemp -d)
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

TEST_FILE="${TMP_DIR}/hello.txt"
echo "hello uploads" > "$TEST_FILE"

echo -e "${YELLOW}1) POST /uploads${NC}"
UPLOAD_RESP=$(curl -s -X POST "${BASE_URL}/uploads" \
  -F "file=@${TEST_FILE};type=text/plain" \
  -w "\n%{http_code}")
HTTP_CODE=$(echo "$UPLOAD_RESP" | tail -n1)
BODY=$(echo "$UPLOAD_RESP" | sed '$d')
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
echo "HTTP Status: $HTTP_CODE"

if [ "$HTTP_CODE" != "201" ]; then
  echo -e "${RED}✗ Expected 201${NC}"
  exit 1
fi

KEY=$(echo "$BODY" | jq -r '.key' 2>/dev/null || true)
if [ -z "${KEY}" ] || [ "${KEY}" == "null" ]; then
  echo -e "${RED}✗ Missing key in response${NC}"
  exit 1
fi

echo ""
echo -e "${YELLOW}2) GET /uploads/${KEY}${NC}"
DOWNLOADED="${TMP_DIR}/downloaded.txt"
DL_CODE=$(curl -s -o "${DOWNLOADED}" -w "%{http_code}" "${BASE_URL}/uploads/${KEY}")
echo "HTTP Status: $DL_CODE"

if [ "$DL_CODE" != "200" ]; then
  echo -e "${RED}✗ Expected 200${NC}"
  exit 1
fi

if diff -q "$TEST_FILE" "$DOWNLOADED" >/dev/null; then
  echo -e "${GREEN}✓ Download matches upload${NC}"
else
  echo -e "${RED}✗ Download does not match upload${NC}"
  exit 1
fi

echo ""
echo -e "${GREEN}All upload tests passed!${NC}"

