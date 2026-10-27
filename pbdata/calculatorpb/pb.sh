#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

cd "${ROOT}"
protoc \
  --go_out=. --go_opt=module=gozero-demo \
  --go-grpc_out=. --go-grpc_opt=module=gozero-demo \
  pbdata/calculatorpb/calculator.proto
