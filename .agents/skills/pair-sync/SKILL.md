---
name: pair-sync
description: Edit a KOSMOS bilingual doc pair (foo.md + foo.zh.md, e.g. docs/glossary or an Agent Note) or re-record its .i18n.yaml hash sidecar. Use when touching any .zh.md file, adding a heading to a paired doc, or when verify-translation-pair.sh reports drift.
---

# Pair sync

Both sides of a pair and its sidecar move in the same change;
`scripts/verify-translation-pair.sh` ends every pairing step green.

## Steps

1. Edit both sides in one change. Sections pair positionally (same heading
   count, same order); keys come from the English headings, so the Chinese
   side may translate its headings freely.
2. Convert the Chinese side's punctuation at the byte level and verify counts:

   ```bash
   perl -CSD -pi -e 'if (/\p{Han}/ && !/^\[/) { tr/\x{3a}\x{2c}\x{28}\x{29}\x{3b}/\x{ff1a}\x{ff0c}\x{ff08}\x{ff09}\x{ff1b}/ }' <doc>.zh.md
   grep -c $'\xef\xbc\x9a' <doc>.zh.md   # full-width colons; > 0 on prose
   grep -c ':' <doc>.zh.md               # 0 on pure-prose pairs
   ```

3. Re-record the sidecar: `bash scripts/verify-translation-pair.sh --write <doc>`.
4. Corpus green: `bash scripts/verify-translation-pair.sh` (no arguments
   verifies every pair that has a sidecar). A red line means step 1 or 2
   missed a side — fix the pair, then re-record.

## Reference

- **Full-width trap**: full-width punctuation typed into tool parameters can
  land on disk as half-width. Always set Chinese prose punctuation through the
  byte conversion and check the counts; rendered output proves nothing. CJK
  characters and `。、——` arrive intact.
- Chinese prose uses full-width `：（）；`. Markdown link lines (starting with
  `[`), inline code, URLs, and identifiers keep ASCII.
- **Drift meanings**: EN drift = English side changed since recording; ZH
  drift = Chinese side changed; stale record key = an English heading was
  renamed. All resolve the same way: bring both sides to the intended
  content, then re-record.
