# 测试客户机

一台用于 agent 测试的非特权 QEMU 客户机——openEuler 24.03 LTS，内核 6.6——跑在带 `/dev/kvm` 与用户态网络的容器里，外加 agent 上传的 server。镜像选择的理由与代价：

[QEMU 客户机 Agent Note](../../.agents/notes/implemented/testing/2026-10-25-qemu-guest-test-image.zh.md)

## 前提

- 装有 Compose v2 的 Docker，且宿主机 `/dev/kvm` 允许容器使用。
- 宿主机有 Go：`up.sh` 会交叉编译 guest 用的 agent 与 server。

## 用法

```sh
test/qemu/up.sh          # 构建、启动、等 SSH 就绪；打印启动事实
test/qemu/push-agent.sh  # 用 scp 把 agent 放进 guest 并启动
test/qemu/ssh.sh         # 进 guest 的 shell
test/qemu/smoke.sh       # up + push-agent + 等 server 收到 Record
test/qemu/down.sh        # 停止
```

状态——镜像缓存、runner SSH 密钥、每次启动的 overlay、串口日志、启动事实——都在 `.runbin/qemu/`，已 gitignore。

## 访问路径

- 宿主机到 guest：`ssh -p 2222 kosmos@127.0.0.1`；Compose 把端口发布在宿主机回环上，runner 密钥是 `.runbin/qemu/ssh/id_ed25519`。
- guest 到 server：agent 往 `http://10.0.2.2:8080` 发，这是 QEMU 的宿主机别名，由 guest 宿主容器转发到 Compose 的 `server` 服务。这条路径上没有任何宿主机特权。
- guest 到互联网：QEMU 用户态 NAT。

## 可调项

| 变量 | 默认值 | 含义 |
|---|---|---|
| `KOSMOS_GUEST_IMAGE` | `openEuler` | 设为 `ubuntu` 启动 24.04 保底镜像 |
| `KOSMOS_GUEST_CPUS` | 2 | guest 的 vCPU 数 |
| `KOSMOS_GUEST_MEMORY_MB` | 2048 | guest 内存 |
| `KOSMOS_SSH_PORT` | 2222 | 发布到宿主机的 SSH 端口 |
| `KOSMOS_SERVER_PORT` | 8080 | 发布到宿主机的 server 端口 |
| `KOSMOS_BUILD_NETWORK` | `default` | 构建镜像时需要访问宿主机回环上的代理就设为 `host` |

镜像在 [images.env](images.env) 里按 URL 与校验和钉住；首次启动前校验和，之后 qcow2 留在缓存里。
