#!/bin/bash
# Test Compass Data Fabric API directly with curl

# YOU NEED TO SET YOUR ACCESS TOKEN HERE
# Get it from your browser's dev tools -> Application -> Cookies -> m3-session
# Or from the backend logs when you login
ACCESS_TOKEN="YOUR_ACCESS_TOKEN_HERE"

BASE_URL="https://mingle-ionapi.inforcloudsuite.com/XK3JRT8CJCAF9GWY_TRN/DATAFABRIC/compass/v2"

# Simple test query
QUERY="SELECT mop.PLPN, mop.PLPS, mop.FACI, mop.ITNO, mop.PSTS, mop.WHST, mpreal.DRDN as linked_co_number, mpreal.DRDL as linked_co_line FROM MMOPLP mop LEFT JOIN MPREAL mpreal ON mpreal.AOCA = '5' AND CAST(mpreal.ARDN AS BIGINT) = mop.PLPN AND mpreal.DOCA = '3' AND mpreal.deleted = 'false' WHERE mop.deleted = 'false' AND mop.PSTS IN ('10', '20') LIMIT 5"

echo "=== Step 1: Submit Query ==="
SUBMIT_RESPONSE=$(curl -s -X POST \
  "${BASE_URL}/jobs/?records=5" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: text/plain" \
  -H "Accept: application/json" \
  --data-raw "${QUERY}")

echo "$SUBMIT_RESPONSE" | jq '.'

QUERY_ID=$(echo "$SUBMIT_RESPONSE" | jq -r '.queryId')
echo ""
echo "Query ID: $QUERY_ID"

echo ""
echo "=== Step 2: Wait 3 seconds ==="
sleep 3

echo ""
echo "=== Step 3: Check Status ==="
STATUS_RESPONSE=$(curl -s -X GET \
  "${BASE_URL}/jobs/${QUERY_ID}/status/?timeout=0" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Accept: application/json")

echo "$STATUS_RESPONSE" | jq '.'

STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.status')

if [ "$STATUS" = "FINISHED" ] || [ "$STATUS" = "COMPLETED" ]; then
  echo ""
  echo "=== Step 4: Fetch Results ==="
  RESULTS=$(curl -s -X GET \
    "${BASE_URL}/jobs/${QUERY_ID}/result/?offset=0&limit=10" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "Accept: application/json")

  echo "$RESULTS" | jq '.'

  echo ""
  echo "=== Field Names ==="
  echo "$RESULTS" | jq -r '.[0] | keys[]' 2>/dev/null
else
  echo "Query not finished yet. Status: $STATUS"
  echo "Run: curl -s '${BASE_URL}/jobs/${QUERY_ID}/result/?offset=0&limit=10' -H 'Authorization: Bearer ${ACCESS_TOKEN}' | jq '.'"
fi
