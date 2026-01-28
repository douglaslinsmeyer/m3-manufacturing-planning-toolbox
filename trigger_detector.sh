#!/bin/bash
# Trigger joint delivery detector for PRD environment

docker compose exec -T postgres psql -U postgres -d m3_planning <<EOF
-- Clear previous joint delivery issues for this job
DELETE FROM detected_issues
WHERE job_id = 'job-1769477439249489130'
  AND detector_type = 'joint_delivery_date_mismatch';

-- Show we cleared the old issues
SELECT 'Cleared old joint delivery issues' as status;
EOF

# Now trigger the detector through the backend worker
# The worker will pick up this message and execute the detector
cat <<'JSON' | docker compose exec -T backend sh -c "
cat > /tmp/detector_msg.json
cat /tmp/detector_msg.json
echo ''
echo 'Message created - now we need a NATS publisher...'
"
{
  "jobId": "job-1769477439249489130-joint_delivery_date_mismatch",
  "parentJobId": "job-1769477439249489130",
  "detectorName": "joint_delivery_date_mismatch",
  "environment": "PRD",
  "company": "100",
  "facility": "AZ1"
}
JSON

echo ""
echo "Note: NATS CLI not available. Alternative: Run full detection job."
