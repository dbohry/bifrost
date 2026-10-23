#!/bin/sh
# Runs every browser check against bin/bifrost, each on its own empty data directory.
# Build first with `make build`. Needs Chromium: run `npm install` in this directory once.
set -eu
cd "$(dirname "$0")/.."
port=18100
for check in flows backup views keyboard status a11y; do
  data=$(mktemp -d)
  BIFROST_DATA_DIR="$data" BIFROST_ADDR="127.0.0.1:$port" bin/bifrost serve >"$data/server.log" 2>&1 &
  pid=$!
  sleep 1
  status=0
  BASE="http://127.0.0.1:$port" node "e2e/$check.js" || status=$?
  kill "$pid"
  [ "$status" -eq 0 ] || exit "$status"
  port=$((port + 1))
done
