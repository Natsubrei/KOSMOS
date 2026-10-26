# Agent Note: QEMU guest image for in-kernel agent tests

Status: implemented

[English](2026-10-25-qemu-guest-test-image.md) | 中文

## 问题

eBPF 探针必须打在真实客户机内核上，且不给 agent 宿主机级权限。测试装置是容器里的 QEMU。

## 决策

测试客户机是 **openEuler 24.03 LTS**（内核 6.6），选定于 2026-10-25。**Ubuntu 24.04 LTS**（内核 6.8）是文档中的保底。QEMU 以非特权方式运行：`devices: [/dev/kvm]` 加上 `group_add: [<host kvm GID>]`；用户态网络；hostfwd 做 SSH。启动预算是 10–30 s 到 SSH 就绪，由 healthcheck 卡住。每次启动都检查 BTF，并写进启动事实。

## 已搭好的测试栈

环境在 `test/qemu/`，用法看它的 README。

- `compose.yaml` 跑两个服务：`qemu-guest`（Debian 容器里的 QEMU，只拿到 `/dev/kvm`）与 `server`（本仓库的 `cmd/server` 二进制，发布在宿主机回环上便于查看）。
- `images.env` 按 URL 与发布方校验和钉住镜像；`entrypoint.sh` 首次启动前校验，缓存 qcow2，每次启动都新建写时复制 overlay。
- 用 `cloud-localds` 生成的 NoCloud seed 注入 runner 的 SSH 公钥并创建 `kosmos` 用户；`up.sh` 首次运行时生成该密钥。
- agent 由 `push-agent.sh` 经 hostfwd 用 scp 投递，镜像里不带任何 agent 负载。
- 启动事实写进 `.runbin/qemu/run/facts`：镜像、到 SSH 的秒数、内核、BTF、cloud-init 标记、QEMU RSS、客户机内存。

## 启动、开销与访问路径

2026-10-26 在带 `/dev/kvm` 的 WSL2 宿主机上实测：

- openEuler：13 s 到 SSH 就绪，内核 `6.6.0-28.0.0.34.oe2403.x86_64`，`/sys/kernel/btf/vmlinux` 存在（约 5 MB）。
- Ubuntu 保底：12 s 到 SSH 就绪，内核 `6.8.0-147-generic`，BTF 存在。
- 默认 2 vCPU / 2048 MiB 下空闲：客户机 MemTotal 1713 MiB，客户机已用约 220 MiB，QEMU RSS 约 600–640 MiB，启动后容器 CPU 低于 1%。
  openEuler 的内核 cmdline 保留 `crashkernel=256M`，所以客户机看到的内存比配置的少约 340 MiB；配置 1024 MiB 时只有 707 MiB 可用，因此默认取 2048。
- 两个镜像都认 NoCloud seed：注入的公钥能认证，`runcmd` 标记存在。
- 宿主机到客户机：`ssh -p 2222 kosmos@127.0.0.1`，密钥是 `.runbin/qemu/ssh/id_ed25519`。
- 客户机到 server：agent 往 `http://10.0.2.2:8080` 发，由客户机宿主容器转发到 `server:8080`。客户机到 Compose 网络没有路由，这条路径上也没有任何宿主机特权。
- `test/qemu/smoke.sh` 端到端验证这条链路，冷启动约 20 s 完成。

## 考虑过的替代

- **以 Ubuntu 24.04 作为主镜像**（内核 6.8，有 BTF，约 600 MB）：否决作为主镜像——项目选定 openEuler 24.03 LTS；Ubuntu 仍作保底。
- **特权容器**：否决——带 `/dev/kvm` 与宿主机 kvm GID 的非特权 QEMU 可行；已知陷阱是宿主机与镜像的 GID 不一致。
- **不启客户机、在宿主机上跑 bpf**：否决——探针打在真实客户机内核上，且无宿主机级权限。
- **把 agent 烧进 seed**：否决——每次重编 agent 都要重建 seed 并重启客户机；经 hostfwd 的 scp 让投递离开启动路径。
- **用共享的 9p 或 virtiofs 目录投递**：否决——多一个客户机内核模块依赖和一个要排查的挂载点，而 sshd 本来就是就绪判定所需的。
- **把客户机接入 Compose 网络**（host 网络或网桥）：否决——用户态网络加 `socat` 转发把 `/dev/kvm` 保持为容器唯一持有的特权。
- **用 QEMU 的 `guestfwd` 代替 `socat` 转发**：否决——它要求客户机指向合成地址 `10.0.2.100`，而不是 agent 默认使用的宿主机别名 `10.0.2.2`。

## 后果

- 空闲资源下限已知：每台空闲客户机按 2 vCPU、配置 2048 MiB（客户机可见 1713 MiB）、宿主侧 QEMU RSS 约 620 MiB 预留；并行跑池之前要重测。
- agent 测试假定 6.6 的 CO-RE 客户机、有 BTF，且容器能打开 `/dev/kvm`。没有 KVM 时栈在启动阶段失败；`KOSMOS_QEMU_ACCEL=tcg` 是未经验证的退路，不是受支持的路径。
- Compose 把发布的端口送到容器网卡，所以 hostfwd 在容器内必须绑 `0.0.0.0`；只绑回环会让宿主机连接被重置。
- 绑在宿主机回环上的代理从默认 Docker 构建网络不可达；这类宿主机设 `KOSMOS_BUILD_NETWORK=host`。
