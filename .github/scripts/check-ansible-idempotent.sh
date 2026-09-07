#!/usr/bin/env bash
set -euo pipefail

ansible_dir="${1:-ansible}"
log_file="$(mktemp)"

(cd "$ansible_dir" && ansible-playbook site.yml) | tee "$log_file"

if grep -qE "changed=[1-9]" "$log_file"; then
  echo "::error::Playbook is not idempotent - second run reported changed tasks"
  exit 1
fi

echo "Idempotency confirmed: second run reported no changes"
