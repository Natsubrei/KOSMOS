# Guest host for the KOSMOS test environment. QEMU runs unprivileged inside
# this container: the only host capability it gets is /dev/kvm, and the guest
# is reached over user-mode networking, so no host bridge or TAP exists.
FROM debian:bookworm-slim

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      ca-certificates \
      cloud-image-utils \
      curl \
      openssh-client \
      qemu-system-x86 \
      qemu-utils \
      socat \
      xz-utils \
 && rm -rf /var/lib/apt/lists/*

COPY cloud-init/user-data /etc/kosmos/user-data
COPY cloud-init/meta-data /etc/kosmos/meta-data
COPY images.env /etc/kosmos/images.env
COPY entrypoint.sh /usr/local/bin/entrypoint

ENTRYPOINT ["/usr/local/bin/entrypoint"]
