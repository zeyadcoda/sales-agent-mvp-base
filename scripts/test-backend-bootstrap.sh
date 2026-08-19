#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../backend"
GOTOOLCHAIN=local go test ./...
