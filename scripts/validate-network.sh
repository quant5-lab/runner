#!/bin/sh
# Check if tcpdump captured packets on ports 80/443

LOGFILE="$1"

if [ ! -f "$LOGFILE" ]; then
  echo "✅ No network log - no activity"
  exit 0
fi

# Extract packet count from tcpdump summary
CAPTURED=$(grep "packets captured" "$LOGFILE" | awk '{print $1}')

if [ -z "$CAPTURED" ] || [ "$CAPTURED" -eq 0 ]; then
  echo "✅ No network activity detected (0 packets)"
  exit 0
fi

echo "❌ NETWORK ACTIVITY: $CAPTURED packets on ports 80/443"
cat "$LOGFILE"
exit 1
