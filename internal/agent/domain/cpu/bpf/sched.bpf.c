// SPDX-License-Identifier: GPL-2.0 OR BSD-3-Clause
// In-kernel on-CPU aggregation for the CPU collection domain.
// tp_btf/sched_switch accounts prev's delta into a hash keyed by
// {tid, tgid, start_time}; userspace drains the map and does not consult /proc.
//
// vmlinux.h is dumped at generate time from the build machine's BTF; clang
// emits CO-RE relocations by field name. The agent host never sees that header.
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

struct task_key {
	__u32 tid;
	__u32 tgid;
	__u64 start_time;
};

struct task_val {
	__u64 on_cpu_ns;
};

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 16384);
	__type(key, struct task_key);
	__type(value, struct task_val);
} runtime_ns SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(max_entries, 1);
	__type(key, __u32);
	__type(value, __u64);
} last_ts SEC(".maps");

SEC("tp_btf/sched_switch")
int BPF_PROG(on_sched_switch, bool preempt, struct task_struct *prev, struct task_struct *next)
{
	(void)preempt;
	(void)next;

	__u64 now = bpf_ktime_get_ns();
	__u32 zero = 0;
	__u64 *last = bpf_map_lookup_elem(&last_ts, &zero);
	if (!last)
		return 0;
	__u64 prev_ts = *last;
	*last = now;
	if (prev_ts == 0 || now <= prev_ts)
		return 0;

	__s32 tid = BPF_CORE_READ(prev, pid);
	if (tid <= 0)
		return 0;

	struct task_key key = {
		.tid = (__u32)tid,
		.tgid = (__u32)BPF_CORE_READ(prev, tgid),
		.start_time = BPF_CORE_READ(prev, start_time),
	};
	__u64 delta = now - prev_ts;
	struct task_val *v = bpf_map_lookup_elem(&runtime_ns, &key);
	if (v) {
		v->on_cpu_ns += delta;
		return 0;
	}
	struct task_val nv = {.on_cpu_ns = delta};
	bpf_map_update_elem(&runtime_ns, &key, &nv, BPF_ANY);
	return 0;
}

char LICENSE[] SEC("license") = "Dual BSD/GPL";
