# Test guest

One unprivileged QEMU guest for agent tests — openEuler 24.03 LTS, kernel 6.6 —
in a container with `/dev/kvm` and user-mode networking, plus the server the
agent uploads to. Why this image and what it costs is the
[QEMU guest Agent Note](../../.agents/notes/implemented/testing/2026-10-25-qemu-guest-test-image.md).

## Prerequisites

- Docker with Compose v2 and a host `/dev/kvm` the container may use.
- Go on the host: `up.sh` cross-compiles the agent and server for the guest.

## Use

```sh
test/qemu/up.sh          # build, boot, wait for SSH; prints the boot facts
test/qemu/push-agent.sh  # scp the agent into the guest and start it
test/qemu/ssh.sh         # shell in the guest
test/qemu/smoke.sh       # up + push-agent + wait for records on the server
test/qemu/down.sh        # stop
```

State — image cache, runner SSH key, per-boot overlay, serial log, boot facts —
lives in `.runbin/qemu/` and is gitignored.

## Access path

- Host to guest: `ssh -p 2222 kosmos@127.0.0.1`; compose publishes the port on
  the host loopback, and the runner key is `.runbin/qemu/ssh/id_ed25519`.
- Guest to server: the agent posts to `http://10.0.2.2:8080`, the QEMU host
  alias, which the guest-host container relays to the compose `server` service.
  Nothing on that path has host privileges.
- Guest to the internet: QEMU user-mode NAT.

## Knobs

| Variable | Default | Meaning |
|---|---|---|
| `KOSMOS_GUEST_IMAGE` | `openEuler` | `ubuntu` boots the 24.04 fallback |
| `KOSMOS_GUEST_CPUS` | 2 | guest vCPUs |
| `KOSMOS_GUEST_MEMORY_MB` | 2048 | guest memory |
| `KOSMOS_SSH_PORT` | 2222 | published SSH port |
| `KOSMOS_SERVER_PORT` | 8080 | published server port |
| `KOSMOS_BUILD_NETWORK` | `default` | `host` when a proxy on the host loopback must be reachable during the image build |

Images are pinned by URL and checksum in [images.env](images.env); the checksum
is verified before the first boot, and the qcow2 is cached afterwards.
