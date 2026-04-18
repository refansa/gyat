#!/bin/bash
set -e

# This script always reinitializes the test workspace from a clean state.
# Any existing tmp/gyat-test directory will be removed and recreated.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_DIR="$PROJECT_ROOT/bin"
GYAT_BIN="$BIN_DIR/gyat"
TEST_DIR="$PROJECT_ROOT/tmp/gyat-test"

echo "Building gyat..."
mkdir -p "$BIN_DIR"
go build -o "$GYAT_BIN" .

echo "Cleaning up existing test directory..."
rm -rf "$TEST_DIR"

echo "Creating test umbrella repository at $TEST_DIR..."

mkdir -p "$TEST_DIR/services/auth"
mkdir -p "$TEST_DIR/services/api"
mkdir -p "$TEST_DIR/services/web"

echo "# Auth Service" > "$TEST_DIR/services/auth/README.md"
echo "# API Service" > "$TEST_DIR/services/api/README.md"
echo "# Web Service" > "$TEST_DIR/services/web/README.md"

git init "$TEST_DIR/services/auth" --quiet
git init "$TEST_DIR/services/api" --quiet
git init "$TEST_DIR/services/web" --quiet

git init "$TEST_DIR" --quiet

# Add services to .gitignore (source repos to track, not commit)
echo "/services" >> "$TEST_DIR/.gitignore"

echo "Initializing gyat workspace..."
(cd "$TEST_DIR" && "$GYAT_BIN" init)
(cd "$TEST_DIR" && "$GYAT_BIN" track services/auth)
(cd "$TEST_DIR" && "$GYAT_BIN" track services/api)
(cd "$TEST_DIR" && "$GYAT_BIN" track services/web)
(cd "$TEST_DIR" && "$GYAT_BIN" list)