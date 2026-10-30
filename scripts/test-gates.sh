#!/usr/bin/env bash
# Negative controls for the gates themselves. A gate that always exits 0, or
# that exits nonzero for the wrong reason, guards nothing — and the gates here
# guard everything else. Each case breaks one thing in a throwaway clone,
# requires the owning gate to turn red, restores it, and requires green.
#
# The clone is HEAD, so this checks the committed gates rather than whatever is
# in the working tree. It runs in CI; run it locally after touching scripts/.
set -uo pipefail
cd "$(dirname "$0")/.."

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# --no-hardlinks: mktemp may land on a different filesystem than the repo, where
# the local-clone optimization fails.
git clone -q --local --no-hardlinks . "$tmp/clone" || {
  echo "test-gates: cannot clone the repo" >&2
  exit 1
}

# The clone supplies HEAD's tree and a real index; the gates under test come
# from the working tree, so a gate being edited is tested before it is
# committed. In CI the two are the same. Committing that copy inside the clone
# gives every case a baseline to reset to.
cp -a scripts/. "$tmp/clone/scripts/"
git -C "$tmp/clone" add -A
git -C "$tmp/clone" -c user.name=test-gates -c user.email=test-gates@invalid \
  commit -q --allow-empty -m "gates under test" || {
  echo "test-gates: cannot snapshot the gates under test" >&2
  exit 1
}
base=$(git -C "$tmp/clone" rev-parse HEAD)

status=0
fails=0

# check <label> <expect> <gate...>: expect is red or green. A red gate must
# say something; a green one must print its OK line, so a gate that silently
# does nothing cannot pass as green.
check() {
  local label=$1 expect=$2
  shift 2
  local out rc
  out=$(cd "$tmp/clone" && "$@" 2>&1)
  rc=$?
  if [[ $expect == red ]]; then
    if (( rc == 0 )); then
      printf 'test-gates: %s: expected a red gate, got exit 0\n%s\n' "$label" "$out" >&2
      status=1
      fails=$((fails + 1))
    elif [[ -z ${out//[[:space:]]/} ]]; then
      printf 'test-gates: %s: red without a message\n' "$label" >&2
      status=1
      fails=$((fails + 1))
    fi
  elif (( rc != 0 )); then
    printf 'test-gates: %s: expected green, got exit %d:\n%s\n' "$label" "$rc" "$out" >&2
    status=1
    fails=$((fails + 1))
  elif ! grep -q 'OK' <<< "$out"; then
    printf 'test-gates: %s: green without an OK line\n%s\n' "$label" "$out" >&2
    status=1
    fails=$((fails + 1))
  fi
}

reset_clone() {
  git -C "$tmp/clone" reset -q --hard "$base"
  git -C "$tmp/clone" clean -qfd
}

# 1. Markdown links: a dead relative link is red. The gate reads the index, so
#    the break has to be staged too.
reset_clone
printf '\n[gone](no-such-file.md)\n' >> "$tmp/clone/docs/architecture.md"
git -C "$tmp/clone" add docs/architecture.md
check "md-links dead link" red bash scripts/verify-md-links.sh
reset_clone
check "md-links restored" green bash scripts/verify-md-links.sh

# 2. Pairing: an untracked counterpart must not fill a hole, and an English
#    edit without re-recording is drift.
reset_clone
git -C "$tmp/clone" rm -q --cached docs/glossary.zh.md
check "pairing missing counterpart in index" red bash scripts/verify-translation-pair.sh
reset_clone
printf '\nAn unstaged sentence.\n' >> "$tmp/clone/docs/glossary.md"
git -C "$tmp/clone" add docs/glossary.md
check "pairing EN drift" red bash scripts/verify-translation-pair.sh --staged
reset_clone
check "pairing restored" green bash scripts/verify-translation-pair.sh

# 3. Notes: wrong status, an archived note without its Archived line, and a
#    note outside the closed lifecycle tree.
reset_clone
sed -i '3s/^Status: implemented$/Status: proposed/' \
  "$tmp/clone/.agents/notes/implemented/architecture/2026-10-26-report-evidence-model.md"
check "notes wrong status" red bash scripts/verify-agent-notes.sh
reset_clone
mkdir -p "$tmp/clone/.agents/notes/archived/architecture"
cp "$tmp/clone/.agents/notes/implemented/architecture/2026-10-26-report-evidence-model.md" \
  "$tmp/clone/.agents/notes/archived/architecture/2026-10-30-probe.md"
check "notes archived without Archived line" red bash scripts/verify-agent-notes.sh
reset_clone
mkdir -p "$tmp/clone/.agents/notes/draft/architecture"
cp "$tmp/clone/.agents/notes/implemented/architecture/2026-10-26-report-evidence-model.md" \
  "$tmp/clone/.agents/notes/draft/architecture/2026-10-30-probe.md"
check "notes unknown lifecycle" red bash scripts/verify-agent-notes.sh
reset_clone
check "notes restored" green bash scripts/verify-agent-notes.sh

# 4. Layering: a domain-free package importing a domain contract.
reset_clone
sed -i 's|^import (|import (\n\t_ "github.com/natsubrei/kosmos/contracts/cpu"|' \
  "$tmp/clone/internal/server/httpapi/httpapi.go"
check "layering contracts import" red bash scripts/check-layering.sh
reset_clone
check "layering restored" green bash scripts/check-layering.sh

# 5. Shell syntax, including the hooks that carry no .sh suffix.
reset_clone
printf 'if true; then\n' >> "$tmp/clone/scripts/git-hooks/pre-commit"
check "shell syntax broken hook" red bash scripts/verify-shell-syntax.sh
reset_clone
check "shell syntax restored" green bash scripts/verify-shell-syntax.sh

# 6. Index build: a build input that was never staged.
reset_clone
git -C "$tmp/clone" rm -q --cached go.sum
check "index build missing go.sum" red bash scripts/verify-index-build.sh
reset_clone
check "index build restored" green bash scripts/verify-index-build.sh

if ((status == 0)); then
  echo "test-gates: OK (every gate turns red on its own breakage and green again)"
fi
exit "$status"
