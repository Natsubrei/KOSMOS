#!/usr/bin/env bash
# End-to-end proof: boot the guest, drop the agent in, run one CPU hog with a
# known PID, and wait until the server reports a Conclusion about that exact
# PID. A report about some other task does not pass.
set -euo pipefail
cd "$(dirname "$0")"
. ./env.sh

./up.sh
./push-agent.sh

# The guest survives the run, so the hog is cleaned up on every exit path.
cleanup_hog() {
  [[ -n ${hog_pid:-} ]] && guest_ssh "kill $hog_pid" >/dev/null 2>&1 || true
}
trap cleanup_hog EXIT

echo "smoke: starting a CPU hog in the guest"
# The PID is the proof's anchor: the report must localize to this task. `yes`
# sits at 100% CPU from its first tick, so one 10s window exceeds the 1s bound.
hog_pid=$(guest_ssh 'nohup yes >/dev/null 2>&1 & echo $!' | tr -dc '0-9')
if [[ -z $hog_pid ]]; then
  echo "smoke: could not read the hog PID from the guest" >&2
  exit 1
fi
echo "smoke: hog PID $hog_pid"

echo "smoke: waiting for a report naming PID $hog_pid at http://127.0.0.1:$server_port/reports"
deadline=$((SECONDS + 90))
while (( SECONDS < deadline )); do
  body=$(curl -fsS -H 'X-User-Id: smoke' "http://127.0.0.1:$server_port/reports" || true)
  if printf '%s' "$body" | grep -q "\"pid\":$hog_pid,"; then
    echo "smoke: OK — the report localizes to PID $hog_pid"
    # Only the matching Report: dumping every Report buries the proof and costs
    # far more context than the run is worth.
    printf '%s' "$body" | python3 -c '
import json, sys
reports = json.load(sys.stdin)
for r in reports:
    if r.get("conclusion", {}).get("subject", {}).get("pid") == int(sys.argv[1]):
        print(json.dumps(r, indent=2))
' "$hog_pid"
    cat "$state/run/facts"
    exit 0
  fi
  sleep 2
done

echo "smoke: no report for PID $hog_pid within 90s; guest agent log, records, and reports follow" >&2
guest_ssh 'cat /var/log/kosmos-agent.log' >&2 || true
curl -fsS -H 'X-User-Id: smoke' "http://127.0.0.1:$server_port/records" >&2 || true
echo >&2
curl -fsS -H 'X-User-Id: smoke' "http://127.0.0.1:$server_port/reports" >&2 || true
exit 1
