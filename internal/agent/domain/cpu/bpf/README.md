# Regenerating the CPU bytecode

`cpu_bpfel.o` and `cpu_bpfeb.o` are committed because the agent embeds them
with `go:embed`. A checkout therefore builds without clang, `bpftool`, or
`vmlinux.h`.

Regenerating needs clang, `bpftool`, libbpf headers, `llvm-strip`, and sysfs
BTF:

```text
go generate ./internal/agent/domain/cpu
```

That runs [gen.sh](gen.sh), which dumps `vmlinux.h` from this machine's BTF
(gitignored — clang emits CO-RE relocations by field name, so the header's
layout never reaches an agent host), compiles, and strips DWARF. DWARF is
roughly 800 KB per object and the agent embeds one of them; `.BTF` and
`.BTF.ext` carry the relocations and survive the strip.

The agent host runs none of this. It loads the embedded object against its own
`/sys/kernel/btf/vmlinux`, so clients may differ in kernel and architecture.
