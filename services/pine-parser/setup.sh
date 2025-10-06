#!/bin/sh
set -e

echo "Installing Python dependencies for Pine Script parser..."
pip3 install --no-cache-dir -r "$(dirname "$0")/requirements.txt"
echo "✓ Python dependencies installed successfully"
