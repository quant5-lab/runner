#!/bin/bash
# Dependency installation to ~/.local

set -e

COLOR_GREEN='\033[0;32m'
COLOR_BLUE='\033[0;34m'
COLOR_YELLOW='\033[1;33m'
COLOR_RESET='\033[0m'

echo_info() { echo -e "${COLOR_BLUE}ℹ ${COLOR_RESET}$1"; }
echo_success() { echo -e "${COLOR_GREEN}✓${COLOR_RESET} $1"; }
echo_warn() { echo -e "${COLOR_YELLOW}⚠${COLOR_RESET} $1"; }

GO_VERSION="1.23.4"
GO_MIN_VERSION="1.21"
LOCAL_DIR="$HOME/.local"
GO_ROOT="$LOCAL_DIR/go"

check_go_version() {
    if command -v go &> /dev/null; then
        CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        REQUIRED_MAJ=$(echo $GO_MIN_VERSION | cut -d. -f1)
        REQUIRED_MIN=$(echo $GO_MIN_VERSION | cut -d. -f2)
        CURRENT_MAJ=$(echo $CURRENT_VERSION | cut -d. -f1)
        CURRENT_MIN=$(echo $CURRENT_VERSION | cut -d. -f2)
        
        if [ "$CURRENT_MAJ" -gt "$REQUIRED_MAJ" ] || \
           ([ "$CURRENT_MAJ" -eq "$REQUIRED_MAJ" ] && [ "$CURRENT_MIN" -ge "$REQUIRED_MIN" ]); then
            echo_success "Go $CURRENT_VERSION sufficient"
            return 0
        fi
    fi
    return 1
}

check_required_tools() {
    MISSING=""
    for cmd in wget tar; do
        if ! command -v $cmd &> /dev/null; then
            MISSING="$MISSING $cmd"
        fi
    done
    
    if [ -n "$MISSING" ]; then
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo_warn "Missing required tools:$MISSING"
        echo ""
        echo "Install with: apt-get install wget tar"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        exit 1
    fi
}

install_go_local() {
    echo_info "Installing Go $GO_VERSION to $GO_ROOT"
    
    mkdir -p "$LOCAL_DIR"
    
    GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TARBALL}"
    
    echo_info "Downloading $GO_URL"
    wget -q --show-progress "$GO_URL" -O "/tmp/${GO_TARBALL}"
    
    if [ -d "$GO_ROOT" ]; then
        echo_info "Removing old $GO_ROOT"
        rm -rf "$GO_ROOT"
    fi
    
    echo_info "Extracting to $GO_ROOT"
    tar -C "$LOCAL_DIR" -xzf "/tmp/${GO_TARBALL}"
    rm "/tmp/${GO_TARBALL}"
    
    if ! grep -q "$GO_ROOT/bin" ~/.bashrc; then
        echo_info "Adding Go to PATH in ~/.bashrc"
        echo '' >> ~/.bashrc
        echo '# Go (user-local)' >> ~/.bashrc
        echo "export PATH=\$PATH:$GO_ROOT/bin" >> ~/.bashrc
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
    fi
    
    export PATH=$PATH:$GO_ROOT/bin
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin
    
    echo_success "Go $GO_VERSION installed to $GO_ROOT"
}

main() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Installing Dependencies"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    check_required_tools
    
    if check_go_version; then
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo_success "Ready"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        return 0
    fi
    
    install_go_local
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo_success "Ready"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo_info "Reload shell: source ~/.bashrc"
    echo_info "Then run: make setup"
    echo ""
}

main "$@"
