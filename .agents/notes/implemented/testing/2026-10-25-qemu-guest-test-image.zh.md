# Agent Note: QEMU guest image for in-kernel agent tests

Status: implemented

[English](2026-10-25-qemu-guest-test-image.md) | 中文

## 问题

eBPF 探针必须打在真实客户机内核上，且不给 agent 宿主机级权限。测试装置是容器里的 QEMU。

## 决策

测试客户机是 **openEuler 24.03 LTS**（内核 6.6），选定于 2026-10-25。**Ubuntu 24.04 LTS**（内核 6.8）是文档中的保底。QEMU 以非特权方式运行：`devices： [/dev/kvm]` 加上 `group_add： [<host kvm GID>]`；用户态网络；hostfwd 做 SSH。启动预算是 10–30 s 到 SSH 就绪，由 healthcheck 卡住。BTF 在每次镜像构建时校验。

## 考虑过的替代

- **以 Ubuntu 24.04 作为主镜像**（内核 6.8，有 BTF，约 600 MB）：否决作为主镜像——项目选定 openEuler 24.03 LTS；Ubuntu 仍作保底。
- **特权容器**：否决——带 `/dev/kvm` 与宿主机 kvm GID 的非特权 QEMU 可行；已知陷阱是宿主机与镜像的 GID 不一致。
- **不启客户机、在宿主机上跑 bpf**：否决——探针打在真实客户机内核上，且无宿主机级权限。

## 后果

后续环境准备在此镜像上测量空闲资源下限。agent 测试假定 6.6 的 CO-RE 客户机，且容器能用 `/dev/kvm`。
