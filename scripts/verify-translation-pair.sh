#!/usr/bin/env bash
# Bilingual-pair gate:
# every heading section of an English document must have a counterpart on the
# Chinese side, and both must hash identically to the record in the sidecar
# (<doc>.i18n.yaml). After editing either side of a pair, bring the other
# along, then re-record:
#   scripts/verify-translation-pair.sh --write docs/glossary.md
#   scripts/verify-translation-pair.sh --cached docs/glossary.md
# With no document arguments, every pair that has a sidecar is verified.
# --cached reads both sides and the record from the index, so an unstaged edit
# cannot pass a commit.
# Sections pair positionally; keys come from the English headings, so
# translated Chinese headings do not break the record. Translation quality,
# terminology, and punctuation are review's job; this script does not scan them.
set -euo pipefail
cd "$(dirname "$0")/.."

write=0
cached=0
docs=()
for arg in "$@"; do
  case "$arg" in
    --write) write=1 ;;
    --cached) cached=1 ;;
    --*) printf 'verify-translation-pair: unknown option %s\n' "$arg" >&2; exit 2 ;;
    *) docs+=("$arg") ;;
  esac
done

if (( cached && write )); then
  printf 'verify-translation-pair: --cached and --write are exclusive\n' >&2
  exit 2
fi
if (( cached )); then
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
fi

if (( write && ${#docs[@]} == 0 )); then
  printf 'verify-translation-pair: --write needs explicit document paths\n' >&2
  exit 2
fi
if (( ! write && ${#docs[@]} == 0 )); then
  while IFS= read -r sidecar; do
    if git check-ignore -q -- "${sidecar#./}"; then
      continue
    fi
    docs+=("${sidecar%.i18n.yaml}.md")
  done < <(find . -name '*.i18n.yaml' -not -path './.git/*' | sort)
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

exit "$status"
