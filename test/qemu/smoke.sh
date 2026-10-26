#!/usr/bin/env bash
# End-to-end proof of the test environment: boot the guest, drop the agent in,
# and wait until the server lists records from it.
set -euo pipefail
cd "$(dirname "$0")"
. ./env.sh

./up.sh
./push-agent.sh

echo "smoke: waiting for records at http://127.0.0.1:$server_port/records"
deadline=$((SECONDS + 60))
while (( SECONDS < deadline )); do
  body=$(curl -fsS -H 'X-User-Id: smoke' "http://127.0.0.1:$server_port/records" || true)
  # awk keeps this exit-0 on zero matches; grep would trip pipefail.
  count=$(printf '%s' "$body" | awk '{ n += gsub(/"domain":/, "") } END { print n+0 }')
  if (( count > 0 )); then
    echo "smoke: OK — server holds $count record(s) uploaded from the guest"
    cat "$state/run/facts"
    exit 0
  fi
  sleep 1
done

echo "smoke: no records arrived within 60s; guest agent log follows" >&2
guest_ssh 'cat /var/log/kosmos-agent.log' >&2 || true
exit 1
