#!/bin/bash

TOKEN="eyJraWQiOiJrZzoyMDNkZjNlMy01YTdkLTRmN2MtYjhkZC1lMGUxNWExYmMxYTAiLCJhbGciOiJSUzI1NiJ9.eyJUZW5hbnQiOiJYSzNKUlQ4Q0pDQUY5R1dZX1RSTiIsIklkZW50aXR5MiI6IjRkYTMwYjJmLWUwYzYtNDcxZi1iMmIwLTZiZmE2MDgzOTViYyIsInNjb3BlIjoib3BlbmlkIHByb2ZpbGUiLCJJRlNBdXRoZW50aWNhdGlvbk1vZGUiOiJPTlBSRU1JU0VfSURFTlRJVElFUyIsIkVuZm9yY2VTY29wZXNGb3JDbGllbnQiOiIwIiwiZ3JhbnRfaWQiOiIxYjk0ZWRlZC0yY2FkLTQxYjEtYTkzMS1mMmMwZDQ5MGI3NGIiLCJJbmZvclNUU0lzc3VlZFR5cGUiOiJBUyIsImNsaWVudF9pZCI6IlhLM0pSVDhDSkNBRjlHV1lfVFJOfjRCMG4xSUhMenFvaGFLOHlsOThPbzJKVmhXSEFSWnVZUkpsVHJWZVBUM2siLCJqdGkiOiIyZDE0ZTc2OC0wNDViLTQ3ODEtYjM2OC1jYzBjOTkyYTM1OWYiLCJpYXQiOjE3NjkwMjM5MzcsIm5iZiI6MTc2OTAyMzkzNywiZXhwIjoxNzY5MDMxMTM3LCJpc3MiOiJodHRwczovL21pbmdsZS1zc28uaW5mb3JjbG91ZHN1aXRlLmNvbTo0NDMiLCJhdWQiOiJodHRwczovL21pbmdsZS1pb25hcGkuaW5mb3JjbG91ZHN1aXRlLmNvbSJ9.Af3FYwqPTCS5rc_vdeeWKRmZrIqePt5XRzYZw98RKW9DIMp4cXUuBDCBkUZpecv7HPN-NAYlh-asCDv8qCQRHwrjnvHjT_Os6Qt5d-4cAaboxEcjBAkdwE0VG84KP9gID9r2wkQrYEorV-OOuCR6xeaY965_pbtF2mX4_Ex-9vxKl0dRy3NmkOccdIVGBL4zEKmt688xejbshXpwvrgcl_5uVk6GslvELoQE7JzgO101qP8kkQTrtCM7X6JcaaA4kagoDIjjarmeT8zEFi2VKad4kbLsRqvv2S6bp2sx9SV01MByylGRUpbA49BylWimTmByUV30mXsPjagFfxug5w"

BASE_URL="https://mingle-ionapi.inforcloudsuite.com/XK3JRT8CJCAF9GWY_TRN/DATAFABRIC/compass/v2"

echo "=== Testing MMOPLP + MPREAL LEFT JOIN ==="
echo ""

# Test query with LEFT JOIN
QUERY='SELECT mop.PLPN, mop.PLPS, mop.FACI, mpreal.DRDN as linked_co_number, mpreal.DRDL as linked_co_line FROM MMOPLP mop LEFT JOIN MPREAL mpreal ON mpreal.AOCA = '"'"'5'"'"' AND CAST(mpreal.ARDN AS BIGINT) = mop.PLPN AND mpreal.DOCA = '"'"'3'"'"' AND mpreal.deleted = '"'"'false'"'"' WHERE mop.deleted = '"'"'false'"'"' AND mop.PSTS IN ('"'"'10'"'"', '"'"'20'"'"') LIMIT 3'

echo "Query: $QUERY"
echo ""

# Submit
echo "1. Submitting..."
SUBMIT=$(curl -s -X POST "${BASE_URL}/jobs/?records=10" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: text/plain" \
  -d "${QUERY}")

echo "$SUBMIT" | jq '.'
QID=$(echo "$SUBMIT" | jq -r '.queryId')
echo ""

# Poll status
echo "2. Polling status..."
for i in {1..10}; do
  sleep 2
  STATUS=$(curl -s "${BASE_URL}/jobs/${QID}/status/?timeout=0" \
    -H "Authorization: Bearer ${TOKEN}")

  STAT=$(echo "$STATUS" | jq -r '.status')
  echo "  Attempt $i: $STAT"

  if [ "$STAT" = "FINISHED" ]; then
    break
  fi
done

echo ""
echo "3. Fetching results..."
RESULTS=$(curl -s "${BASE_URL}/jobs/${QID}/result/?offset=0&limit=10" \
  -H "Authorization: Bearer ${TOKEN}")

echo "$RESULTS" | jq '.'
echo ""
echo "=== Field Names ==="
echo "$RESULTS" | jq -r '.[0] | keys[]' 2>/dev/null
