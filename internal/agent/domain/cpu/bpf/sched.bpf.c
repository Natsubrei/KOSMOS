// SPDX-License-Identifier: GPL-2.0 OR BSD-3-Clause
// In-kernel on-CPU aggregation for the CPU collection domain.
// sched_switch updates a hash keyed by tid; userspace drains the map.
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

struct switch_ctx {
	unsigned long long pad;
	char prev_comm[16];
	int prev_pid;
	int prev_prio;
	long prev_state;
	char next_comm[16];
	int next_pid;
	int next_prio;
};

struct task_val {
	__u64 on_cpu_ns;
};

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 16384);
	__type(key, __u32);
	__type(value, struct task_val);
} runtime_ns SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(max_entries, 1);
	__type(key, __u32);
	__type(value, __u64);
} last_ts SEC(".maps");

SEC("tracepoint/sched/sched_switch")
int on_sched_switch(struct switch_ctx *ctx)
{
	__u64 now = bpf_ktime_get_ns();
	__u32 zero = 0;
	__u64 *last = bpf_map_lookup_elem(&last_ts, &zero);
	if (!last)
		return 0;
	__u64 prev_ts = *last;
	*last = now;
	if (prev_ts == 0 || now <= prev_ts)
		return 0;

	__s32 pid = ctx->prev_pid;
	if (pid <= 0)
		return 0;

	__u32 key = (__u32)pid;
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
