#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
./scripts/golangci.sh run ./...
python3 - <<'PYTHON'
from pathlib import Path
for path in Path('.').rglob('*'):
    if not path.is_file() or '.git' in path.parts:
        continue
    if len(path.read_bytes().splitlines()) > 1000:
        raise SystemExit(f'File exceeds 1,000 lines: {path}')
PYTHON
go test ./...
go vet ./...
go run ./cmd/openapi -check
