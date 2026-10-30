//go:build linux

package cpu

// The kernel uapi headers need the architecture's asm/ include directory, whose
// path differs per distribution and architecture; ask the host compiler rather
// than pinning one machine's path here. -no-strip keeps bpf2go from calling
// llvm-strip, which is not always installed.
//
// The generated .o files are committed: the agent embeds them, so a checkout
// builds without clang. Regenerating needs clang, libbpf headers, and gcc.
//
//go:generate sh -c "go run github.com/cilium/ebpf/cmd/bpf2go -tags linux -no-strip -type task_val cpu bpf/sched.bpf.c -- -O2 -g -Wall -Werror -I/usr/include/$(gcc -print-multiarch)"
