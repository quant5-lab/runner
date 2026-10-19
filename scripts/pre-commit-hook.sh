#!/bin/sh
# Tracked pre-commit hook template — install via: make install-hooks
set -e

export PATH="$HOME/.local/go/bin:/usr/local/go/bin:$PATH"
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"

if ! command -v go >/dev/null 2>&1; then
    echo "✗ Go not found. Run: make install"
    exit 1
fi

echo "Running pre-commit validation..."

REPO_ROOT="$(git rev-parse --show-toplevel)"
"$REPO_ROOT/scripts/golden-guard.sh"

make all
