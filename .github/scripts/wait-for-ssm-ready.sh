#!/usr/bin/env bash
set -euo pipefail

project_tag="${PROJECT_TAG:-zero-credential-pipeline}"
timeout_seconds="${TIMEOUT_SECONDS:-300}"
poll_interval_seconds="${POLL_INTERVAL_SECONDS:-10}"

elapsed=0
while [ "$elapsed" -lt "$timeout_seconds" ]; do
  status=$(aws ssm describe-instance-information \
    --filters "Key=tag:Project,Values=${project_tag}" \
    --query 'InstanceInformationList[0].PingStatus' \
    --output text)

  if [ "$status" = "Online" ]; then
    echo "Instance registered with SSM after ${elapsed}s"
    exit 0
  fi

  sleep "$poll_interval_seconds"
  elapsed=$((elapsed + poll_interval_seconds))
done

echo "::error::Instance did not register with SSM within ${timeout_seconds}s"
exit 1
