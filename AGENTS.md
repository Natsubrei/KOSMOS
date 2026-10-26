# AGENTS.md

KOSMOS is a client-server tool for diagnosing anomalies on Linux virtual machines and hosts: agents collect low-overhead in-kernel evidence via eBPF with a small CPU and memory footprint on the target machine; a central server runs the diagnosis. Its code architecture is highly scalable and every design is changeable — the system must absorb continuous change.

## Conventions

- Comments describe non-obvious contracts or rationale that code cannot express; they do not restate what code already implies.
- One explanation has one home; essential contract facts may repeat locally — link the owning document for architecture and rationale.
- TODO markers by urgency: `FIXME` blocks a release, `TODO` means fix soon, `XXX` may be fixed someday, no commitment.
- Agent Notes record durable decision rationale — alternatives and consequences; mechanical changes are exempt.
- Repo-specific workflows live as skills under `.agents/skills/`.

## Development

- Run `scripts/check.sh` before reporting a change complete; it stops at the
  first red gate (formatting, vet, tests, layering, bilingual pairs, links).
- Evidence does not survive a history rewrite: after a rebase or forced push, re-run the affected checks before treating the branch as verified.
- Push rewritten branches with `--force-with-lease` and abort on remote movement; raw `--force` is forbidden.
- Report only the commands actually run; never claim a check that was not executed.
- Never bypass a failing check to make progress; fix the cause or stop and report.
- Commit messages follow Conventional Commits — `type(scope): imperative subject` (feat, fix, docs, refactor, test, chore, build, ci, perf); a `!` after the type marks a breaking change with a `BREAKING CHANGE:` footer.
- A commit's message and the prose it adds or edits must not refer to unstaged files. If the prose needs those paths, stage them in this commit or an earlier one, or do not commit.
- Development proceeds in small, iterative steps: land each step working and green before starting the next; prefer the smallest change that moves the work forward.

## Language

English: this file, code comments, commit messages. Human-facing docs — the glossary and Agent Notes included — are bilingual pairs, `foo.md` (English) + `foo.zh.md` (Chinese); either side may be authored, sync the counterpart minimally in the same change. Pairs are gated by `scripts/verify-translation-pair.sh`: after editing either side, bring the other along and re-record the `<doc>.i18n.yaml` sidecar. Relative Markdown links are gated by `scripts/verify-md-links.sh`. Chinese prose uses full-width punctuation; inline code, URLs, and identifiers keep ASCII.

## Deployment and testing

- The server side deploys fully containerized.
- Agents are tested inside a QEMU Linux guest that itself runs in a container: eBPF probes run against a real guest kernel with no host-level privileges.
