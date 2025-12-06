#!/bin/bash
# Verify system dependencies (Go, build tools)

set -e

COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_BLUE='\033[0;34m'
COLOR_RESET='\033[0m'

echo_info() { echo -e "${COLOR_BLUE}ℹ ${COLOR_RESET}$1"; }
echo_success() { echo -e "${COLOR_GREEN}✓${COLOR_RESET} $1"; }
echo_error() { echo -e "${COLOR_RED}✗${COLOR_RESET} $1"; }

MISSING_DEPS=0

check_command() {
    local cmd=$1
    
    if command -v "$cmd" &> /dev/null; then
        echo_success "$cmd"
        return 0
    else
        echo_error "$cmd NOT FOUND"
        MISSING_DEPS=$((MISSING_DEPS + 1))
        return 1
    fi
}

main() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Dependency Check"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    check_command go
    check_command gofmt
    check_command make
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if [ $MISSING_DEPS -eq 0 ]; then
        echo_success "Ready"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        exit 0
    else
        echo_error "Missing $MISSING_DEPS required tools"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo_info "Install: make install"
        echo ""
        exit 1
    fi
}

main "$@"
        exit 1
    fi
}

main "$@"
