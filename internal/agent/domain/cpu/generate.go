//go:build linux

package cpu

// The embedded bytecode is regenerated with `go generate`, which runs
// bpf/gen.sh: it dumps vmlinux.h from this machine's BTF, compiles CO-RE C, and
// strips DWARF so the committed objects stay small. A checkout needs neither
// clang nor bpftool — the .o files are committed and embedded.
//
//go:generate sh bpf/gen.sh
