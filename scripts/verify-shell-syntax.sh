#!/usr/bin/env bash
# Syntax-check every shell script git tracks, including the hooks, which carry
# no .sh suffix. These scripts gate everything else, so a typo in one of them
# would fail open rather than red.
set -euo pipefail
cd "$(dirname "$0")/.."

status=0
while IFS= read -r script; do
  if ! out=$(bash -n "$script" 2>&1); then
    printf 'verify-shell-syntax: %s\n%s\n' "$script" "$out" >&2
    status=1
  fi
done < <(git ls-files '*.sh' 'scripts/git-hooks/*')

if (( status == 0 )); then
  echo "verify-shell-syntax: OK (every tracked shell script parses)"
fi
exit "$status"
