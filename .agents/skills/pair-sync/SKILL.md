---
name: pair-sync
description: Edit a KOSMOS bilingual doc pair (foo.md + foo.zh.md, e.g. docs/glossary or an Agent Note) or re-record its .i18n.yaml hash sidecar. Use when touching any .zh.md file, adding a heading to a paired doc, or when verify-translation-pair.sh reports drift.
---

# Pair sync

Both sides of a pair and its sidecar move in the same change;
`scripts/verify-translation-pair.sh` ends every pairing step green.

## Steps

1. Load [docs/i18n/translation-rules.md](../../../docs/i18n/translation-rules.md)
   and [docs/i18n/terminology.md](../../../docs/i18n/terminology.md). Every listed
   term follows its row and its 不要译作 prohibitions. A new term gets a row in
   the same change; do not invent a rendering inline. The pairing gate does not
   check terminology, punctuation, or spacing; do not add scanners for those.
   A green pairing hash does not prove translation quality.
2. Edit both sides in one change. Sections pair positionally (same heading
   count, same order); keys come from the English headings, so the Chinese
   side may translate its headings freely. Follow translation-rules: full-width
   punctuation in Chinese prose; ASCII inside backticks; one half-width space
   between Chinese text and a Latin word or numeral.
3. Re-record the sidecar: `bash scripts/verify-translation-pair.sh --write <doc>`.
4. Corpus green: `bash scripts/verify-translation-pair.sh` (no arguments
   verifies every pair that has a sidecar). A red line means step 2 missed a
   side — fix the pair, then re-record.

## Reference

- Do not run a whole-line punctuation converter over `.zh.md`; it rewrites
  identifiers inside backticks.
- **Drift meanings**: EN drift = English side changed since recording; ZH
  drift = Chinese side changed; stale record key = an English heading was
  renamed. All resolve the same way: bring both sides to the intended
  content, then re-record.
