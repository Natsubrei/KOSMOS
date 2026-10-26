#!/usr/bin/env bash
# Markdown link gate: every relative link target in a tracked .md file must
# exist. Anchors, absolute URLs, and paths starting with / are not checked;
# fenced code blocks are skipped. One line per dead link; nonzero exit.
set -euo pipefail
cd "$(dirname "$0")/.."

status=0
for doc in $(git ls-files '*.md'); do
  dir=$(dirname "$doc")
  # Strip fenced code blocks, then emit "file<TAB>target" per relative link.
  while IFS=$'\t' read -r file target; do
    [ -z "$target" ] && continue
    case "$target" in
      *://*|mailto:*|'#'*) continue ;;
      /*) continue ;;
    esac
    path=${target%%#*}
    [ -z "$path" ] && continue
    if [ ! -e "$dir/$path" ]; then
      printf 'verify-md-links: %s: %s does not exist\n' "$file" "$target"
      status=1
    fi
  done < <(awk '
    /^```/ { fence = !fence; next }
    fence { next }
    {
      line = $0
      while (match(line, /\]\([^)]*\)/)) {
        target = substr(line, RSTART + 2, RLENGTH - 3)
        gsub(/^ +| +$/, "", target)
        sub(/ .*$/, "", target)
        if (target != "") print FILENAME "\t" target
        line = substr(line, RSTART + RLENGTH)
      }
    }
  ' "$doc")
done

if (( status == 0 )); then
  echo "verify-md-links: OK"
fi
exit "$status"
