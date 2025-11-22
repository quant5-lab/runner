#!/bin/bash
# Validates that .config files follow the naming convention
# Config filename must match PineScript source filename (without .pine extension)
#
# Usage: ./scripts/validate-configs.sh
#   Exit 0: All configs valid
#   Exit 1: Invalid config names found

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 Config Filename Validation"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Rule: Config filename must match PineScript source filename (without .pine)"
echo "  Example: strategies/my-strategy.pine → out/my-strategy.config"
echo ""

# Find all .config files (excluding template.config)
CONFIG_FILES=$(find out -name "*.config" -type f ! -name "template.config" 2>/dev/null || true)

if [ -z "$CONFIG_FILES" ]; then
    echo -e "${YELLOW}⚠ No config files found in out/ directory${NC}"
    echo ""
    exit 0
fi

VALID_COUNT=0
INVALID_COUNT=0
ORPHAN_COUNT=0

echo "Checking config files:"
echo ""

for config_file in $CONFIG_FILES; do
    config_name=$(basename "$config_file" .config)
    
    # Search for corresponding .pine file in strategies/
    pine_file="strategies/${config_name}.pine"
    
    if [ -f "$pine_file" ]; then
        echo -e "  ${GREEN}✓${NC} ${BLUE}${config_name}.config${NC} → ${pine_file}"
        VALID_COUNT=$((VALID_COUNT + 1))
    else
        # Check if it exists in subdirectories
        found_pine=$(find strategies -name "${config_name}.pine" -type f 2>/dev/null | head -1)
        if [ -n "$found_pine" ]; then
            echo -e "  ${YELLOW}⚠${NC} ${config_name}.config → ${found_pine} ${YELLOW}(in subdirectory)${NC}"
            VALID_COUNT=$((VALID_COUNT + 1))
        else
            echo -e "  ${RED}✗${NC} ${config_name}.config ${RED}(no matching .pine file found)${NC}"
            ORPHAN_COUNT=$((ORPHAN_COUNT + 1))
        fi
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Summary:"
echo "  Valid configs:   ${VALID_COUNT}"
echo "  Orphan configs:  ${ORPHAN_COUNT}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ $ORPHAN_COUNT -gt 0 ]; then
    echo -e "${RED}✗ Validation failed: ${ORPHAN_COUNT} orphan config(s) found${NC}"
    echo ""
    echo "To fix:"
    echo "  1. Rename config to match source filename, OR"
    echo "  2. Delete orphaned config file if no longer needed"
    echo ""
    exit 1
fi

echo -e "${GREEN}✓ All config files valid${NC}"
echo ""
exit 0
