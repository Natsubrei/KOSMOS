# AGENTS.md

KOSMOS is a client-server tool for diagnosing anomalies on Linux virtual machines and hosts: agents collect low-overhead in-kernel evidence via eBPF with a small CPU and memory footprint on the target machine; a central server runs the diagnosis. Its code architecture is highly scalable and every design is changeable — the system must absorb continuous change.

Read [docs/architecture.md](docs/architecture.md) before changing a seam: it maps the packages, what runs in which process, and where each variable concern is allowed to appear.

## Conventions

- Comments describe non-obvious contracts or rationale that code cannot express; they do not restate what code already implies.
- One explanation has one home; essential contract facts may repeat locally — link the owning document for architecture and rationale.
- TODO markers by urgency: `FIXME` blocks a release, `TODO` means fix soon, `XXX` may be fixed someday, no commitment.
- Agent Notes record durable decision rationale — alternatives and consequences; mechanical changes are exempt.
- Repo-specific workflows live as skills under `.agents/skills/`.

## Development

- `scripts/check.sh` is the full gate — formatting, vet, tests, layering, Agent Note structure, bilingual pairs, Markdown links — and is what CI runs. The pre-commit hook checks only what matches the staged files; pre-push runs the tests.
- Evidence does not survive a history rewrite: after a rebase or forced push, re-run the affected checks before treating the branch as verified.
- Push rewritten branches with `--force-with-lease` and abort on remote movement; raw `--force` is forbidden.
- Report only the commands actually run; never claim a check that was not executed.
- Never bypass a failing check to make progress; fix the cause or stop and report.
- Commit messages follow Conventional Commits — `type(scope): imperative subject` (feat, fix, docs, refactor, test, chore, build, ci, perf); a `!` after the type marks a breaking change with a `BREAKING CHANGE:` footer.
- A commit's message and the prose it adds or edits must not refer to unstaged files. If the prose needs those paths, stage them in this commit or an earlier one, or do not commit.
- Development proceeds in small, iterative steps: land each step working and green before starting the next; prefer the smallest change that moves the work forward.

## Language

English: this file, code comments, commit messages. Human-facing docs — the glossary, the architecture map, Agent Notes included — are bilingual pairs, `foo.md` (English) + `foo.zh.md` (Chinese); either side may be authored, sync the counterpart minimally in the same change. The corpus is `README.md`, `docs/`, `.agents/notes/`, and `test/`: a document there must be a complete pair — `foo.zh.md` plus its `foo.i18n.yaml` record — and the deliberate exceptions are listed in `scripts/pairing-exclusions.txt`. `scripts/check.sh` runs the gates; `scripts/verify-agent-notes.sh` owns the Agent Note tree. How to write the other side is `docs/i18n/translation-rules.md`; how a domain term renders is `docs/i18n/terminology.md` (bilingual by construction, not a pair). Load both before editing a `.zh.md` file. The gates do not check terminology, punctuation, or spacing; do not add scanners for those. Chinese prose uses full-width punctuation; inline code, URLs, and identifiers keep ASCII; put a half-width space between Chinese text and a Latin word or numeral.

## Deployment and testing

- The server side deploys fully containerized.
- Agents are tested inside a QEMU Linux guest that itself runs in a container: eBPF probes run against a real guest kernel with no host-level privileges. That guest is [test/qemu](test/qemu/README.md); `test/qemu/smoke.sh` boots it and proves an agent upload reaches the server.
