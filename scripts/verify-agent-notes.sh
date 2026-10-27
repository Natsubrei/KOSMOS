#!/usr/bin/env bash
# Agent Note gate: the lifecycle tree is closed, every English note sits at
# {lifecycle}/{class}/YYYY-MM-DD-slug.md, and its header and section skeleton
# match the lifecycle it lives in. The in-file rules are owned by
# .agents/notes/README.md. Chinese counterparts and their cross-links are the
# pairing gate's job; fenced code is not document structure.
set -euo pipefail
cd "$(dirname "$0")/.."

root=.agents/notes
lifecycles='proposed implemented rejected'
classes='architecture bug-fix feature process simplification testing'
# Non-note Markdown allowed to sit directly at a lifecycle root.
allowlist='AGENTS.md'

status=0
err() {
  printf 'verify-agent-notes: %s\n' "$1" >&2
  status=1
}

in_set() {
  case " $2 " in
    *" $1 "*) return 0 ;;
  esac
  return 1
}

# h2s <file>: the `## ` headings outside fenced code, in order.
h2s() {
  awk '
    /^```/ { fence = !fence; next }
    fence { next }
    /^## / { print }
  ' "$1"
}

has_section() {
  grep -qxF -- "$2" <<< "$1"
}

# banned_implemented <heading>: proposal-era spec-speak that misleads in a
# shipped note.
banned_implemented() {
  case "$1" in
    '## Proposal'|'## Proposal '*) return 0 ;;
    '## Plan'|'## Plan '*) return 0 ;;
    '## Migration plan'|'## Migration plan '*) return 0 ;;
    '## Acceptance criteria'|'## Acceptance criteria '*) return 0 ;;
  esac
  return 1
}

check_format() {
  local rel=$1 lifecycle=$2
  local file="$root/$rel"
  local -a lines=()
  mapfile -t lines < "$file"

  case "${lines[0]:-}" in
    '# Agent Note: '*) ;;
    *) err "format: $rel — line 1 must be '# Agent Note: <title>'" ;;
  esac
  [[ ${lines[0]:-} == '# Agent Note: ' ]] && err "format: $rel — line 1 has an empty title"
  [[ ${lines[1]:-} != '' ]] && err "format: $rel — line 2 must be blank"

  case "${lines[2]:-}" in
    "Status: $lifecycle") ;;
    *) err "format: $rel — line 3 must be 'Status: $lifecycle'" ;;
  esac
  if [[ $lifecycle == rejected ]]; then
    [[ ${lines[2]:-} == 'Status: rejected — '* ]] ||
      err "format: $rel — a rejected note's status carries its reason: 'Status: rejected — <why>'"
  fi
  [[ ${lines[3]:-} != '' ]] && err "format: $rel — line 4 must be blank"

  local status_count
  status_count=$(awk '
    /^```/ { fence = !fence; next }
    fence { next }
    /^Status:/ { count++ }
    END { print count + 0 }
  ' "$file")
  [[ $status_count -ne 1 ]] && err "format: $rel — the file must carry exactly one Status: line (found $status_count)"

  local headings first
  headings=$(h2s "$file")
  first=$(head -n1 <<< "$headings")
  [[ $first != '## Problem' ]] &&
    err "format: $rel — the first section must be '## Problem' (got '${first:-<none>}')"

  local required=''
  case "$lifecycle" in
    proposed) required='## Proposal|## Acceptance criteria|## Risks' ;;
    implemented) required='## Decision|## Consequences' ;;
    rejected) required='## Proposal' ;;
  esac
  local heading
  while IFS= read -r heading; do
    [[ -z $heading ]] && continue
    has_section "$headings" "$heading" ||
      err "format: $rel — a $lifecycle note requires the '$heading' section"
  done < <(tr '|' '\n' <<< "$required")

  if [[ $lifecycle == implemented ]]; then
    while IFS= read -r heading; do
      [[ -z $heading ]] && continue
      if banned_implemented "$heading"; then
        err "format: $rel — '$heading' is proposal-era; implemented notes record shipped reality"
      fi
    done <<< "$headings"
  fi
}

while IFS= read -r file; do
  rel=${file#"$root"/}
  base=$(basename "$rel")
  # A Chinese counterpart is the same note; the pairing gate owns it.
  case "$base" in
    *.zh.md) continue ;;
  esac
  # Top-level notes README and its sidecars are not notes.
  [[ $rel == "$base" ]] && continue

  IFS='/' read -r -a segs <<< "$rel"
  depth=${#segs[@]}

  if [[ $depth -eq 2 ]]; then
    if in_set "${segs[0]}" "$lifecycles" && in_set "${segs[1]}" "$allowlist"; then
      continue
    fi
    err "structure: $rel — expected {lifecycle}/{class}/YYYY-MM-DD-slug.md"
    continue
  fi
  if [[ $depth -ne 3 ]]; then
    err "structure: $rel — expected {lifecycle}/{class}/YYYY-MM-DD-slug.md (got depth $depth)"
    continue
  fi

  lifecycle=${segs[0]}
  class=${segs[1]}
  if ! in_set "$lifecycle" "$lifecycles"; then
    err "structure: $rel — '$lifecycle' is not a lifecycle (allowed: $lifecycles)"
    continue
  fi
  if ! in_set "$class" "$classes"; then
    err "structure: $rel — '$class' is not a class (allowed: $classes)"
    continue
  fi
  if [[ ! $base =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}-.+\.md$ ]]; then
    err "structure: $rel — filename must be YYYY-MM-DD-slug.md"
    continue
  fi

  check_format "$rel" "$lifecycle"

  # The machine-checked header tokens stay English on the Chinese side; every
  # other part of that file belongs to the pairing gate.
  zh_file="$root/${rel%.md}.zh.md"
  if [[ -f $zh_file ]]; then
    mapfile -t zh_lines < "$zh_file"
    case "${zh_lines[0]:-}" in
      '# Agent Note: '*) ;;
      *) err "format: ${rel%.md}.zh.md — line 1 must be '# Agent Note: <title>'" ;;
    esac
    case "${zh_lines[2]:-}" in
      "Status: $lifecycle") ;;
      *) err "format: ${rel%.md}.zh.md — line 3 must be 'Status: $lifecycle'" ;;
    esac
  fi
done < <(find "$root" -name '*.md' | sort)

if ((status == 0)); then
  echo "verify-agent-notes: OK (closed tree, headers and sections match their lifecycle)"
fi
exit "$status"
