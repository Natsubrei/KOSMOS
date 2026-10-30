#!/usr/bin/env bash
# Build the tree as a commit would contain it: what git tracks, not what the
# working directory happens to hold. A file the build needs but nobody staged
# (embedded bytecode, go.sum) breaks a fresh clone, while every other gate in
# this repository is green — they all read the working tree.
#
# The plane is the index, so this also runs from the pre-commit hook.
set -euo pipefail
cd "$(dirname "$0")/.."

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# checkout-index writes index content, so an unstaged edit cannot mask a
# missing file.
git checkout-index -a --prefix="$tmp/"

if ! out=$(cd "$tmp" && go build ./... 2>&1); then
  printf 'verify-index-build: the committed tree does not build:\n%s\n' "$out" >&2
  printf 'verify-index-build: stage every file the build reads (git add).\n' >&2
  exit 1
fi

echo "verify-index-build: OK (the index builds on its own)"
