#!/bin/sh
# Regenerate the embedded CPU bytecode. Run through `go generate`:
#   go generate ./internal/agent/domain/cpu
#
# vmlinux.h is dumped from this machine's BTF and gitignored: clang emits CO-RE
# relocations by field name, so the header's layout never reaches an agent host.
#
# bpf2go is asked not to strip, because that step wants a bare `llvm-strip` on
# PATH; this script strips afterwards instead, with whichever versioned binary
# exists. Stripping DWARF is not cosmetic: -g is required for CO-RE and leaves
# roughly 800 KB of debug sections in a 30 KB object, which the agent embeds.
# .BTF and .BTF.ext carry the relocations and must survive the strip.
set -eu

cd "$(dirname "$0")/.."

strip=$(command -v llvm-strip || command -v llvm-strip-21 || command -v llvm-strip-20 || true)
if [ -z "$strip" ]; then
  echo "gen.sh: llvm-strip not found; the objects would keep ~800 KB of DWARF" >&2
  exit 1
fi

bpftool btf dump file /sys/kernel/btf/vmlinux format c > bpf/vmlinux.h

go run github.com/cilium/ebpf/cmd/bpf2go -tags linux -no-strip \
  -type task_key -type task_val cpu bpf/sched.bpf.c -- \
  -O2 -g -Wall -Werror -I bpf -I /usr/include

for obj in cpu_bpfel.o cpu_bpfeb.o; do
  "$strip" --strip-debug --keep-section .BTF --keep-section .BTF.ext "$obj"
done
