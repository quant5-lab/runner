#!/bin/bash
# Helper script to create a new visualization config with correct naming
#
# Usage: ./scripts/create-config.sh STRATEGY_FILE
#   Example: ./scripts/create-config.sh strategies/my-strategy.pine
#
# This script:
# 1. Validates the strategy file exists
# 2. Creates config with correct filename (source filename without .pine)
# 3. Runs strategy to get indicator names
# 4. Generates config template with actual indicator names

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

STRATEGY_FILE="${1:-}"

if [ -z "$STRATEGY_FILE" ]; then
    echo "Usage: $0 STRATEGY_FILE"
    echo ""
    echo "Example:"
    echo "  $0 strategies/my-strategy.pine"
    echo ""
    exit 1
fi

if [ ! -f "$STRATEGY_FILE" ]; then
    echo -e "${RED}Error: Strategy file not found: ${STRATEGY_FILE}${NC}"
    exit 1
fi

# Extract strategy name from filename (without .pine extension)
STRATEGY_NAME=$(basename "$STRATEGY_FILE" .pine)
CONFIG_FILE="out/${STRATEGY_NAME}.config"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📝 Creating Visualization Config"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Strategy file: ${BLUE}${STRATEGY_FILE}${NC}"
echo "Config file:   ${GREEN}${CONFIG_FILE}${NC}"
echo ""

if [ -f "$CONFIG_FILE" ]; then
    echo -e "${YELLOW}⚠ Config file already exists: ${CONFIG_FILE}${NC}"
    read -p "Overwrite? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled"
        exit 0
    fi
fi

# Check if strategy has been run to get indicator names
DATA_FILE="out/chart-data.json"
NEEDS_RUN=false

if [ ! -f "$DATA_FILE" ]; then
    NEEDS_RUN=true
else
    # Check if data file is from this strategy
    CURRENT_STRATEGY=$(jq -r '.metadata.strategy // empty' "$DATA_FILE" 2>/dev/null || echo "")
    if [ "$CURRENT_STRATEGY" != "$STRATEGY_NAME" ]; then
        NEEDS_RUN=true
    fi
fi

if [ "$NEEDS_RUN" = true ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🚀 Running strategy to extract indicator names..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "This may take a moment. You can also run manually:"
    echo "  make fetch-strategy SYMBOL=BTCUSDT TIMEFRAME=1h BARS=100 STRATEGY=${STRATEGY_FILE}"
    echo ""
    
    # Try to run with minimal data for speed
    make fetch-strategy SYMBOL=BTCUSDT TIMEFRAME=1h BARS=100 STRATEGY="$STRATEGY_FILE" > /tmp/strategy-run.log 2>&1 || {
        echo -e "${RED}Failed to run strategy. See /tmp/strategy-run.log for details${NC}"
        echo ""
        echo "Creating empty config template instead..."
        cat > "$CONFIG_FILE" << 'EOF'
{
  "indicators": {
    "Indicator Name 1": "main",
    "Indicator Name 2": "indicator"
  }
}
EOF
        echo -e "${YELLOW}⚠ Config created with placeholder names${NC}"
        echo "  Edit ${CONFIG_FILE} and replace with actual indicator names"
        echo ""
        exit 0
    }
fi

# Extract indicator names from chart-data.json
INDICATORS=$(jq -r '.indicators | keys[]' "$DATA_FILE" 2>/dev/null || echo "")

if [ -z "$INDICATORS" ]; then
    echo -e "${YELLOW}⚠ No indicators found in chart-data.json${NC}"
    echo ""
    echo "Creating empty config template..."
    cat > "$CONFIG_FILE" << 'EOF'
{
  "indicators": {
    "Indicator Name 1": "main",
    "Indicator Name 2": "indicator"
  }
}
EOF
else
    echo "Found indicators:"
    echo "$INDICATORS" | while read -r ind; do
        echo "  - ${ind}"
    done
    echo ""
    
    # Generate config with actual indicator names
    echo "{" > "$CONFIG_FILE"
    echo '  "indicators": {' >> "$CONFIG_FILE"
    
    FIRST=true
    echo "$INDICATORS" | while read -r ind; do
        if [ "$FIRST" = true ]; then
            echo "    \"${ind}\": \"main\"" >> "$CONFIG_FILE"
            FIRST=false
        else
            echo "    ,\"${ind}\": \"main\"" >> "$CONFIG_FILE"
        fi
    done
    
    echo '  }' >> "$CONFIG_FILE"
    echo '}' >> "$CONFIG_FILE"
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ Config created: ${CONFIG_FILE}${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Next steps:"
echo "  1. Edit config: ${CONFIG_FILE}"
echo "  2. Customize pane assignments (\"main\" or \"indicator\")"
echo "  3. Add styling (color, style, lineWidth) if needed"
echo "  4. Test: make serve && open http://localhost:8000"
echo ""
echo "Example full styling:"
echo '  "My Indicator": {'
echo '    "pane": "indicator",'
echo '    "style": "histogram",'
echo '    "color": "rgba(128, 128, 128, 0.3)"'
echo '  }'
echo ""
