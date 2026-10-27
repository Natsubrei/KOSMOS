#!/usr/bin/env bash
# Bilingual-pair gate:
# every document inside the pairing scope must be a complete pair — `foo.md`,
# `foo.zh.md`, and its `<doc>.i18n.yaml` record — unless the exclusion manifest
# names it. A pair's heading sections must pair positionally and both sides
# must hash identically to the record. Relative links inside a pair must point
# at the sibling of their own locale. After editing either side, bring the
# other along, then re-record:
#   scripts/verify-translation-pair.sh --write docs/glossary.md
#   scripts/verify-translation-pair.sh --cached docs/glossary.md
# With no document arguments, every documented pair in scope is verified;
# `--list` reports each in-scope document's state instead. `--staged` checks the
# in-scope documents among the staged paths, reading all three files from the
# index — that is what the pre-commit hook runs, so the scope rule lives only
# here.
# --cached reads both sides and the record from the index, so an unstaged edit
# cannot pass a commit.
# Sections pair positionally; keys come from the English headings, so
# translated Chinese headings do not break the record. Translation quality,
# terminology, and punctuation are review's job; this script does not scan them.
set -euo pipefail
cd "$(dirname "$0")/.."

manifest=scripts/pairing-exclusions.txt

write=0
cached=0
list=0
staged_only=0
docs=()
for arg in "$@"; do
  case "$arg" in
    --write) write=1 ;;
    --cached) cached=1 ;;
    --list) list=1 ;;
    --staged) staged_only=1; cached=1 ;;
    --*) printf 'verify-translation-pair: unknown option %s\n' "$arg" >&2; exit 2 ;;
    *) docs+=("$arg") ;;
  esac
done

# Normalize foo.zh.md / foo.i18n.yaml / a bare stem to the English anchor.
for i in "${!docs[@]}"; do
  doc=${docs[i]%.i18n.yaml}
  doc=${doc%.zh.md}
  doc=${doc%.md}
  docs[i]=$doc.md
done

if (( cached && write )); then
  printf 'verify-translation-pair: --cached and --write are exclusive\n' >&2
  exit 2
fi
if (( cached )); then
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
fi

# in_scope <path>: the documentation corpus that must be bilingual. Everything
# else — skills, agent instructions, the temporary issue tracker — is outside.
in_scope() {
  case "$1" in
    README.md | docs/*.md | .agents/notes/*.md | test/*.md) return 0 ;;
  esac
  return 1
}

# pairing_excluded <path>: the manifest holds every deliberate exception, so an
# unpaired document is a gate failure by default rather than a silent pass.
# Patterns are bash globs over the whole path; `*` also crosses `/`.
pairing_excluded() {
  local path=$1 pattern
  [[ -f $manifest ]] || return 1
  while IFS= read -r pattern; do
    pattern=${pattern%%#*}
    pattern=${pattern//[[:space:]]/}
    [[ -z $pattern ]] && continue
    # shellcheck disable=SC2254 # the manifest holds patterns, not literals
    case "$path" in
      $pattern) return 0 ;;
    esac
  done < "$manifest"
  return 1
}

# pair_state <doc>: complete, missing, or orphan-zh.
pair_state() {
  local doc=$1
  [[ -f $doc && -f ${doc%.md}.zh.md && -f ${doc%.md}.i18n.yaml ]] && {
    echo complete
    return
  }
  echo missing
}

if ((list)); then
  printf '%-12s %s\n' STATE DOCUMENT
  while IFS= read -r doc; do
    printf '%-12s %s\n' "$(pair_state "$doc")" "$doc"
  done < <(
    git ls-files '*.md' | while IFS= read -r path; do
      case "$path" in
        *.zh.md) continue ;;
      esac
      in_scope "$path" || continue
      pairing_excluded "$path" && continue
      printf '%s\n' "$path"
    done | sort
  )
  exit 0
fi

if ((write && ${#docs[@]} == 0)); then
  printf 'verify-translation-pair: --write needs explicit document paths\n' >&2
  exit 2
fi
if ((staged_only && ${#docs[@]} > 0)); then
  printf 'verify-translation-pair: --staged reads the index; do not pass paths\n' >&2
  exit 2
fi
if ((staged_only)); then
  while IFS= read -r path; do
    case "$path" in
      *.zh.md) continue ;;
    esac
    in_scope "$path" || continue
    pairing_excluded "$path" && continue
    docs+=("$path")
  done < <(git diff --cached --name-only --diff-filter=ACMR | sort)
  if ((${#docs[@]} == 0)); then
    exit 0
  fi
fi
if ((!write && ${#docs[@]} == 0)); then
  # Scope enumeration, not a sidecar search: a document that was never
  # recorded is still required to have its counterpart.
  while IFS= read -r path; do
    docs+=("$path")
  done < <(
    git ls-files '*.md' | while IFS= read -r path; do
      case "$path" in
        *.zh.md) continue ;;
      esac
      in_scope "$path" || continue
      pairing_excluded "$path" && continue
      printf '%s\n' "$path"
    done | sort
  )
fi

# Print "SECTION <idx> <key>" markers followed by that section's normalized
# lines (heading included, fenced code excluded, trailing whitespace
# stripped, blanks dropped). Keys come from the English filename and
# headings, lowercased and slugged.
extract_sections() {
  awk -v root="/$(basename "${1%.md}" | tr '[:upper:]' '[:lower:]')" '
    function slug(s) { s = tolower(s); gsub(/[^a-z0-9]+/, "-", s); gsub(/^-+|-+$/, "", s); return s }
    function emit(    i) {
      printf "SECTION %d %s\n", idx, key
      for (i = 1; i <= n; i++) print lines[i]
    }
    BEGIN { idx = 0; key = root; n = 0; infence = 0 }
    {
      if (index($0, "```") == 1) { infence = !infence; next }
      line = $0
      if (!infence && index($0, "## ") == 1) {
        emit()
        idx++; key = root "/" slug(substr($0, 4)); n = 0
      }
      sub(/[ \t]+$/, "", line)
      if (line != "") lines[++n] = line
    }
    END { emit() }
  ' "$1"
}

hash_lines() {
  local -n lines_ref="$1"
  if (( ${#lines_ref[@]} == 0 )); then
    printf '' | sha256sum | cut -c1-16
  else
    printf '%s\n' "${lines_ref[@]}" | sha256sum | cut -c1-16
  fi
}

# sections_of <file> <hashes-array> <keys-array>: index-ordered hashes/keys.
sections_of() {
  local -n out_h="$2" out_k="$3"
  local line idx key cur_idx=-1
  local -a buf=()
  out_h=(); out_k=()
  while IFS= read -r line; do
    if [[ $line == SECTION\ * ]]; then
      if (( cur_idx >= 0 )); then out_h[cur_idx]=$(hash_lines buf); fi
      read -r _ idx key <<< "$line"
      cur_idx=$idx
      buf=()
      out_k[idx]=$key
    else
      buf+=("$line")
    fi
  done < <(extract_sections "$1")
  if (( cur_idx >= 0 )); then out_h[cur_idx]=$(hash_lines buf); fi
}

status=0

# A Chinese side without its English anchor is invisible to a scope walk over
# `*.md`, so it gets its own pass.
while IFS= read -r zh_path; do
  [[ -f ${zh_path%.zh.md}.md ]] && continue
  printf 'verify-translation-pair: %s: Chinese side without an English anchor\n' "$zh_path" >&2
  status=1
done < <(git ls-files '*.zh.md' | while IFS= read -r path; do
  in_scope "$path" || continue
  pairing_excluded "$path" && continue
  printf '%s\n' "$path"
done | sort)

# links_of <file>: `line<TAB>target` for every relative Markdown link outside
# fenced code. The language switcher is the one line allowed to cross locales.
links_of() {
  awk '
    /^```/ { fence = !fence; next }
    fence { next }
    /^(English \| \[中文\]\([^)]*\)|\[English\]\([^)]*\) \| 中文)$/ { next }
    {
      line = $0
      while (match(line, /\]\([^)]*\)/)) {
        target = substr(line, RSTART + 2, RLENGTH - 3)
        gsub(/^ +| +$/, "", target)
        sub(/ .*$/, "", target)
        if (target != "") print NR "\t" target
        line = substr(line, RSTART + RLENGTH)
      }
    }
  ' "$1"
}

# check_locale_links <en> <zh>: a Chinese side links the Chinese sibling when
# the pair has one, and an English side never links the Chinese one.
check_locale_links() {
  local en=$1 zh=$2 src dir line target path want
  for src in "$en" "$zh"; do
    dir=$(dirname "$src")
    while IFS=$'\t' read -r line target; do
      [[ -z $target ]] && continue
      case "$target" in
        *://* | mailto:* | '#'* | /*) continue ;;
      esac
      path=${target%%#*}
      [[ -z $path ]] && continue
      case "$src" in
        *.zh.md)
          case "$path" in
            *.zh.md) continue ;;
            *.md) ;;
            *) continue ;;
          esac
          [[ -e "$dir/${path%.md}.zh.md" ]] || continue
          want=${path%.md}.zh.md
          ;;
        *)
          case "$path" in
            *.zh.md) ;;
            *) continue ;;
          esac
          want=${path%.zh.md}.md
          ;;
      esac
      printf 'verify-translation-pair: %s:%s: link target %s uses the wrong locale; expected %s\n' \
        "$src" "$line" "$target" "$want" >&2
      status=1
    done < <(links_of "$src")
  done
}

# index_copy <path>: copy the staged blob for path into $tmpdir under its own
# basename (section keys derive from the filename). Returns 1 when the index
# has no such path.
index_copy() {
  local path=$1
  git cat-file -e ":$path" 2>/dev/null || return 1
  git show ":$path" > "$tmpdir/$(basename "$path")"
}

for doc in "${docs[@]}"; do
  en="$doc"
  zh="${doc%.md}.zh.md"
  record="${doc%.md}.i18n.yaml"

  # Link locale is resolved against real paths, so it reads the worktree copy
  # even under --cached; a staged tree that disagrees with the worktree is the
  # commit's own problem, not this check's.
  if [[ -f $en && -f $zh ]]; then
    check_locale_links "$en" "$zh"
  fi

  missing=0
  if (( cached )); then
    for f in "$en" "$zh"; do
      if ! index_copy "$f"; then
        printf 'verify-translation-pair: %s: %s is not staged\n' "$doc" "$f" >&2
        missing=1
      fi
    done
    index_copy "$record" || true
    en="$tmpdir/$(basename "$en")"
    zh="$tmpdir/$(basename "$zh")"
    record="$tmpdir/$(basename "$record")"
  else
    for f in "$en" "$zh"; do
      if [[ ! -f $f ]]; then
        printf 'verify-translation-pair: %s: missing %s\n' "$doc" "$f" >&2
        missing=1
      fi
    done
  fi
  if (( missing )); then
    status=1
    continue
  fi

  declare -a en_k en_h zh_h
  sections_of "$en" en_h en_k
  sections_of "$zh" zh_h zh_k

  if (( ${#en_k[@]} != ${#zh_h[@]} )); then
    printf 'verify-translation-pair: %s: section count mismatch (en %d, zh %d) — structure drifted; sync the pair first\n' "$doc" "${#en_k[@]}" "${#zh_h[@]}" >&2
    status=1
    continue
  fi

  if (( write )); then
    {
      printf '# Bilingual-pair consistency record for %s: per heading section, a hash of\n' "$en"
      printf '# its English and Chinese blocks outside code blocks. Sections pair positionally;\n'
      printf '# keys come from the English headings. After editing either side, bring the other\n'
      printf '# along and re-record with:\n#   scripts/verify-translation-pair.sh --write %s\n' "$en"
      for (( i = 0; i < ${#en_k[@]}; i++ )); do
        printf '%s:\n  en: %s\n  zh: %s\n' "${en_k[i]}" "${en_h[i]}" "${zh_h[i]}"
      done
    } > "$record"
    printf 'verify-translation-pair: recorded %d sections -> %s\n' "${#en_k[@]}" "$record"
    continue
  fi

  if [[ ! -f $record ]]; then
    printf 'verify-translation-pair: %s: no record at %s; run --write after confirming the pair\n' "$doc" "$record" >&2
    status=1
    continue
  fi

  unset rec_en rec_zh
  declare -A rec_en rec_zh
  key=""
  while IFS= read -r line; do
    case "$line" in
      '#'*) ;;
      /*:) key="${line%:}" ;;
      '  en: '*) rec_en[$key]=${line#'  en: '} ;;
      '  zh: '*) rec_zh[$key]=${line#'  zh: '} ;;
    esac
  done < "$record"

  doc_status=0
  for (( i = 0; i < ${#en_k[@]}; i++ )); do
    key=${en_k[i]}
    if [[ -z ${rec_en[$key]+x} ]]; then
      printf 'verify-translation-pair: %s: section %s missing from record (heading changed?) — sync the pair, then --write\n' "$doc" "$key" >&2
      doc_status=1
    elif [[ ${en_h[i]} != ${rec_en[$key]} ]]; then
      printf 'verify-translation-pair: %s: section %s EN drift — sync the pair, then --write\n' "$doc" "$key" >&2
      doc_status=1
    elif [[ ${zh_h[i]} != ${rec_zh[$key]:-} ]]; then
      printf 'verify-translation-pair: %s: section %s ZH drift — bring %s along, then --write\n' "$doc" "$key" "$zh" >&2
      doc_status=1
    fi
  done
  for key in "${!rec_en[@]}"; do
    found=0
    for (( i = 0; i < ${#en_k[@]}; i++ )); do
      if [[ ${en_k[i]} == "$key" ]]; then found=1; break; fi
    done
    if (( found == 0 )); then
      printf 'verify-translation-pair: %s: stale record key %s — sync the pair, then --write\n' "$doc" "$key" >&2
      doc_status=1
    fi
  done

  if (( doc_status == 0 )); then
    printf 'verify-translation-pair: %s OK (%d sections)\n' "$doc" "${#en_k[@]}"
  fi
  status=$((status + doc_status))
done

exit $((status != 0))
