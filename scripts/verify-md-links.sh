#!/usr/bin/env bash
# Markdown link gate: every relative link target in a tracked .md file must
# exist in the index. Anchors, absolute URLs, and paths starting with / are not
# checked; fenced code blocks are skipped. One line per dead link; nonzero exit.
#
# Both sides of the comparison come from the index. Listing the index while
# reading the worktree, or reading the index while testing targets on disk, lets
# an unstaged change hide a dead link; the two planes must agree.
set -euo pipefail
cd "$(dirname "$0")/.."

root=$(pwd)
declare -A present=()
declare -A dirs=()
while IFS= read -r p; do
  present["$p"]=1
  d=$(dirname "$p")
  while [[ $d != . && -n $d ]]; do
    dirs["$d"]=1
    d=$(dirname "$d")
  done
done < <(git ls-files)

status=0
for doc in $(git ls-files '*.md'); do
  # Archived notes are frozen snapshots; outbound links are not current authority.
  case "$doc" in
    .agents/notes/archived/*) continue ;;
  esac
  if ! content=$(git cat-file blob ":$doc" 2>/dev/null); then
    printf 'verify-md-links: %s is tracked but not in the index\n' "$doc"
    status=1
    continue
  fi
  dir=$(dirname "$doc")
  while IFS=$'\t' read -r file target; do
    [[ -z $target ]] && continue
    case "$target" in
      *://*|mailto:*|'#'*) continue ;;
      /*) continue ;;
    esac
    path=${target%%#*}
    [[ -z $path ]] && continue
    normalized=$(realpath -m --relative-to="$root" "$dir/$path")
    if [[ -z ${present[$normalized]:-} && -z ${dirs[$normalized]:-} ]]; then
      printf 'verify-md-links: %s: %s does not exist\n' "$file" "$target"
      status=1
    fi
  done < <(printf '%s\n' "$content" | awk -v file="$doc" '
    /^```/ { fence = !fence; next }
    fence { next }
    {
      line = $0
      while (match(line, /\]\([^)]*\)/)) {
        target = substr(line, RSTART + 2, RLENGTH - 3)
        gsub(/^ +| +$/, "", target)
        sub(/ .*$/, "", target)
        if (target != "") print file "\t" target
        line = substr(line, RSTART + RLENGTH)
      }
    }
  ')
done

if (( status == 0 )); then
  echo "verify-md-links: OK"
fi
exit "$status"
