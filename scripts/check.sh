#!/usr/bin/env bash
# The single local gate: formatting, vet, tests, layering, bilingual pairs,
# Markdown links. Stops at the first red gate with a nonzero exit.
set -euo pipefail
cd "$(dirname "$0")/.."

unformatted=$(gofmt -l cmd contracts internal)
if [[ -n $unformatted ]]; then
  printf 'check: not gofmt-clean:\n%s\n' "$unformatted" >&2
  exit 1
fi

go vet ./...
go test -race ./...
bash scripts/check-layering.sh
bash scripts/verify-translation-pair.sh
bash scripts/verify-md-links.sh

echo "check: OK"
