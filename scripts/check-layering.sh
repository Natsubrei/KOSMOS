#!/usr/bin/env bash
# Layering gate for compile-time domain bundles: contracts are pure types (stdlib only), domain
# bundles never import sibling domains, and packages under internal/agent except domain/
# import no domain — domain knowledge reaches
# the process only by registration in cmd/. Test files are included via
# go list -deps -test. Violations print one line each and exit nonzero.
set -euo pipefail
cd "$(dirname "$0")/.."

module="$(go list -m)"
status=0

# Package dependencies without go-list test scaffolding entries
# ("<pkg> [<pkg>.test]" and "<pkg>.test" lines).
deps() {
  go list -deps -test "$1" | grep -v ' \[' | grep -v '\.test$'
}

# Rule 1: contracts/<domain> imports nothing from this module.
for p in $(go list ./contracts/...); do
  while IFS= read -r dep; do
    printf 'layering: %s imports %s — contracts are pure types\n' "$p" "$dep"
    status=1
  done < <(deps "$p" | grep -F "$module/" | grep -vxF "$p")
done

# Rule 2: a domain bundle imports no sibling domain.
for p in $(go list ./internal/agent/domain/...); do
  while IFS= read -r dep; do
    printf 'layering: %s imports sibling domain %s\n' "$p" "$dep"
    status=1
  done < <(deps "$p" | grep -F "$module/internal/agent/domain/" | grep -vxF "$p")
done

# Rule 3: packages under internal/agent except domain/ import no domain.
for p in $(go list ./internal/agent/...); do
  case "$p" in
    "$module"/internal/agent/domain/*) continue ;;
  esac
  while IFS= read -r dep; do
    printf 'layering: %s imports domain %s — register in cmd/ instead\n' "$p" "$dep"
    status=1
  done < <(deps "$p" | grep -F "$module/internal/agent/domain/")
done

if (( status == 0 )); then
  echo "layering: OK (contracts pure, no sibling-domain imports, domain-free packages import no domain)"
fi
exit "$status"
