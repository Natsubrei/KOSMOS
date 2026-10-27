# Translation rules

English | [中文](translation-rules.zh.md)

How to produce the other side of a documentation pair. Either side may be
authored; that side is the source for the change. [terminology.md](terminology.md)
says how each term renders; [glossary.md](../glossary.md) says what the term
means. These rules bind humans and agents. The pairing gate does not enforce
them.

## Faithfulness

The counterpart says what the authored side says: no added behavior, warnings,
or examples, and no dropped ones. If the pair disagrees on substance, fix the
wrong side, then bring the other along in the same change. Translate meaning,
not word by word; terse stays terse.

## Voice

English is concise professional developer prose. Chinese is institutional
technical Chinese. Do not transplant metaphors. Give sentences an explicit
actor when the target language would otherwise hide it. Address the reader as
你, not 您.

## Structure

Sections pair positionally (same heading count, same order). Heading text may
be translated. Fenced code blocks stay byte-identical, including comments.
Inline code, URLs, and identifiers stay verbatim ASCII — never translated or
reformatted. Relative links keep the same target; the Chinese side of a pair
points at `.zh.md` when that counterpart exists. The pairing gate checks
section hashes and relative links. List shape, table shape, and heading depth
are review's job; do not add scanners for them.

## Terminology

Load [terminology.md](terminology.md) before editing a `.zh.md` file. Every
listed term follows its row and its 不要译作 prohibitions. Do not invent a
rendering inline; a decided term gets a row in the same change. Unlisted
technical terms stay English unless an established Chinese rendering can be
cited.

## Typography

These rules govern the Chinese side. English follows ordinary Markdown.

- One half-width space between Chinese text and a Latin word, and between
  Chinese text and a numeral: `每个 agent 采集 3 类 Record`. No space next to
  a full-width punctuation mark.
- Full-width punctuation in Chinese prose: `，。：；？！（）「」`. Half-width
  punctuation stays inside code spans (`` `Subscribe(cpu)` ``, `` `{id, start, end}` ``),
  inside complete English sentences quoted as-is, and in numbers (`3.5`, `1,024`).
- Parallel items in Chinese prose use 顿号 (`、`), not commas.
- No full-width digits or full-width Latin letters.
- Do not run a whole-line punctuation converter over `.zh.md`; it rewrites
  identifiers inside backticks.

## Quality

A pair is done when a reader of either file alone gets the same facts, caveats,
and tone, and nothing extra. `scripts/verify-translation-pair.sh` checks
hashes and structure. A green pairing hash does not prove translation quality,
terminology, punctuation, or spacing. Do not add scanners for those.
