#!/bin/bash
# Initialize project after Go installation

set -e

COLOR_GREEN='\033[0;32m'
COLOR_BLUE='\033[0;34m'
COLOR_RESET='\033[0m'

echo_info() { echo -e "${COLOR_BLUE}ℹ ${COLOR_RESET}$1"; }
echo_success() { echo -e "${COLOR_GREEN}✓${COLOR_RESET} $1"; }

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Project Setup"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo_info "Downloading Go modules..."
cd golang-port
go mod download
cd ..
echo_success "Modules downloaded"

echo_info "Creating directories..."
mkdir -p out golang-port/build golang-port/coverage
echo_success "Directories created"

echo_info "Building pine-gen..."
cd golang-port
go build -o build/pine-gen ./cmd/pine-gen
cd ..
echo_success "pine-gen built"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo_success "Ready"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo_info "Next: make test"
echo ""
echo ""
