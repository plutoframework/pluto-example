#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
./scripts/golangci.sh run ./cmd/... ./internal/...
python3 scripts/check-lines.py
go test ./cmd/... ./internal/...
go vet ./cmd/... ./internal/...
go run ./cmd/openapi -check

./scripts/frontend-check.sh
