---
labels: [wayfinder:task]
assignee:
---

# 0009 · Provision the QEMU test environment

## Question

Build the containerized test environment per the [QEMU guest image](../.agents/notes/implemented/testing/2026-10-25-qemu-guest-test-image.md) note: a Docker Compose service running a QEMU Linux guest (**openEuler 24.03 LTS, kernel 6.6** — Ubuntu 24.04 as fallback) with cloud-init access, usable as the agent's test target without host privileges. Done when an agent binary can be dropped into the guest and reach the server. Record what was built, image/kernel choices, boot time, and the access path future tickets depend on.
