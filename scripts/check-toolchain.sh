#!/usr/bin/env bash
set -euo pipefail

fail=0
check_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "MISSING: $cmd"
    fail=1
  fi
}

check_cmd go
check_cmd node
check_cmd git
check_cmd pnpm
check_cmd docker

if command -v go >/dev/null 2>&1; then
  actual=$(go version | awk '{print $3}' | sed 's/^go//')
  [[ "$actual" == "1.26.6" ]] || { echo "WRONG Go: $actual (need 1.26.6)"; fail=1; }
fi
if command -v node >/dev/null 2>&1; then
  actual=$(node --version | sed 's/^v//')
  [[ "$actual" == "24.19.0" ]] || { echo "WRONG Node: $actual (need 24.19.0)"; fail=1; }
fi
if command -v pnpm >/dev/null 2>&1; then
  actual=$(pnpm --version)
  [[ "$actual" == "10.34.5" ]] || { echo "WRONG pnpm: $actual (need 10.34.5)"; fail=1; }
fi

exit "$fail"
