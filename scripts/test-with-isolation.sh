#!/bin/sh
set -e

if command -v tcpdump >/dev/null 2>&1; then
  LOG_FILE="/tmp/network.log"
  rm -f "$LOG_FILE"
  
  tcpdump -i any -n -l port 80 or port 443 > "$LOG_FILE" 2>&1 & 
  TCPDUMP_PID=$!
  sleep 1
  
  pnpm vitest run
  TEST_EXIT=$?
  
  sleep 1
  kill $TCPDUMP_PID 2>/dev/null || true
  ./scripts/validate-network.sh "$LOG_FILE"
  exit $TEST_EXIT
else
  pnpm vitest run
fi
