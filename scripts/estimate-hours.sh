#!/bin/bash
# Estimate hours worked based on commit timestamps
# Algorithm: commits within 2h are same session, add 0.5h bonus per session

THRESHOLD=7200  # 2 hours in seconds
BONUS=1800      # 0.5 hours bonus per session (AI-assisted era)

# Get base and head commits for branch comparison
BASE_REF="${1:-main}"
HEAD_REF="${2:-HEAD}"

# Use commit range if provided, otherwise all commits
if [ "$BASE_REF" != "all" ]; then
  COMMIT_RANGE="${BASE_REF}..${HEAD_REF}"
else
  COMMIT_RANGE="--all"
fi

git log $COMMIT_RANGE --format="%at|%an|%s" | sort -t'|' -k1 -n | awk -F'|' -v threshold="$THRESHOLD" -v bonus="$BONUS" -v range="$COMMIT_RANGE" '
{
  timestamp = $1
  author = $2
  message = $3
  
  if (prev_time[author]) {
    diff = timestamp - prev_time[author]
    if (diff <= threshold) {
      hours[author] += diff / 3600
    } else {
      hours[author] += bonus / 3600
    }
  } else {
    hours[author] += bonus / 3600
  }
  
  prev_time[author] = timestamp
  total_commits[author]++
}
END {
  total_hours = 0
  total_commits_all = 0
  
  for (a in hours) {
    total_hours += hours[a]
    total_commits_all += total_commits[a]
  }
  
  if (range != "--all") {
    print "**Scope**: Branch commits only (" range ")"
  } else {
    print "**Scope**: All repository commits"
  }
  
  print ""
  print "Based on commit timestamp analysis (2h session threshold)"
  print ""
  
  for (a in hours) {
    printf "- **%s**: %.1fh (%d commits)\n", a, hours[a], total_commits[a]
  }
  
  print ""
  printf "**Total**: %.1fh across %d commits\n", total_hours, total_commits_all
}
'
