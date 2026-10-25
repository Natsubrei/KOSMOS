#!/usr/bin/env bash
# Bilingual-pair gate for the glossary, ported from dsh's verify-translation-pairing
# method: every heading section of the English document must have a counterpart
# section on the Chinese side, and both must hash identically to the record in the
# sidecar (<doc>.i18n.yaml). After editing either side of a pair, bring the other
# along, then re-record:
#   scripts/verify-glossary-pair.sh --write docs/glossary.md
# Sections pair positionally; keys come from the English headings, so translated
# Chinese headings do not break the record. Translation quality is review's job.
set -euo pipefail

write=0
doc=""
for arg in "$@"; do
  case "$arg" in
    --write) write=1 ;;
    --*) printf 'verify-glossary-pair: unknown option %s\n' "$arg" >&2; exit 2 ;;
    *) doc="$arg" ;;
  esac
done
doc="${doc:-docs/glossary.md}"
en="$doc"
zh="${doc%.md}.zh.md"
record="${doc%.md}.i18n.yaml"

for f in "$en" "$zh"; do
  if [[ ! -f $f ]]; then
    printf 'verify-glossary-pair: missing %s\n' "$f" >&2
    exit 2
  fi
done

rootkey="/$(basename "${doc%.md}" | tr '[:upper:]' '[:lower:]')"

# Print "SECTION <idx> <key>" markers followed by that section's normalized lines
# (heading included, fenced code excluded, trailing whitespace stripped, blanks dropped).
extract_sections() {
  awk -v root="$rootkey" '
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

# sections_of <file> <hashes-array> <keys-array>: index-ordered section hashes/keys.
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

declare -a en_k en_h zh_h
sections_of "$en" en_h en_k
sections_of "$zh" zh_h zh_k

if (( ${#en_k[@]} != ${#zh_h[@]} )); then
  printf 'verify-glossary-pair: section count mismatch (en %d, zh %d) — structure drifted; sync the pair first\n' "${#en_k[@]}" "${#zh_h[@]}" >&2
  exit 1
fi

if (( write )); then
  {
    printf '# Bilingual-pair consistency record for %s: per heading section, a hash of\n' "$en"
    printf '# its English and Chinese blocks outside code blocks. Sections pair positionally;\n'
    printf '# keys come from the English headings. After editing either side, bring the other\n'
    printf '# along and re-record with:\n#   scripts/verify-glossary-pair.sh --write %s\n' "$en"
    for (( i = 0; i < ${#en_k[@]}; i++ )); do
      printf '%s:\n  en: %s\n  zh: %s\n' "${en_k[i]}" "${en_h[i]}" "${zh_h[i]}"
    done
  } > "$record"
  printf 'verify-glossary-pair: recorded %d sections -> %s\n' "${#en_k[@]}" "$record"
  exit 0
fi

if [[ ! -f $record ]]; then
  printf 'verify-glossary-pair: no record at %s; run --write after confirming the pair\n' "$record" >&2
  exit 1
fi

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

status=0
for (( i = 0; i < ${#en_k[@]}; i++ )); do
  key=${en_k[i]}
  if [[ -z ${rec_en[$key]+x} ]]; then
    printf 'verify-glossary-pair: section %s missing from record (heading changed?) — sync the pair, then --write\n' "$key" >&2
    status=1
  elif [[ ${en_h[i]} != ${rec_en[$key]} ]]; then
    printf 'verify-glossary-pair: section %s EN drift — sync the pair, then --write\n' "$key" >&2
    status=1
  elif [[ ${zh_h[i]} != ${rec_zh[$key]:-} ]]; then
    printf 'verify-glossary-pair: section %s ZH drift — bring %s along, then --write\n' "$key" "$zh" >&2
    status=1
  fi
done
for key in "${!rec_en[@]}"; do
  found=0
  for (( i = 0; i < ${#en_k[@]}; i++ )); do
    if [[ ${en_k[i]} == "$key" ]]; then found=1; break; fi
  done
  if (( found == 0 )); then
    printf 'verify-glossary-pair: stale record key %s — sync the pair, then --write\n' "$key" >&2
    status=1
  fi
done

if (( status == 0 )); then
  printf 'verify-glossary-pair: %s OK (%d sections)\n' "$en" "${#en_k[@]}"
fi
exit "$status"
