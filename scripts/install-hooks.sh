#!/bin/bash
# Install git hooks for golang-port tests

HOOK_SOURCE="golang-port/hooks/pre-commit"
HOOK_TARGET=".git/hooks/pre-commit"

if [ ! -f "$HOOK_SOURCE" ]; then
    echo "Error: Hook source not found: $HOOK_SOURCE"
    exit 1
fi

cp "$HOOK_SOURCE" "$HOOK_TARGET"
chmod +x "$HOOK_TARGET"

echo "✓ Git hook installed: $HOOK_TARGET"
echo "  Runs 'go test ./...' before every commit"
