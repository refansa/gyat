#!/bin/bash
# This script always reinitializes the test workspace from a clean state.
# Any existing tmp/gyat-test directory will be removed and recreated.

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
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

git -C "$TEST_DIR/services/auth" init --quiet
git -C "$TEST_DIR/services/api" init --quiet
git -C "$TEST_DIR/services/web" init --quiet

git -C "$TEST_DIR/services/auth" add README.md
git -C "$TEST_DIR/services/api" add README.md
git -C "$TEST_DIR/services/web" add README.md

git -C "$TEST_DIR/services/auth" commit -m "Initial commit" --no-gpg-sign --quiet
git -C "$TEST_DIR/services/api" commit -m "Initial commit" --no-gpg-sign --quiet
git -C "$TEST_DIR/services/web" commit -m "Initial commit" --no-gpg-sign --quiet

git init "$TEST_DIR" --quiet

# Add services to .gitignore (source repos to track, not commit)
echo "/services" >> "$TEST_DIR/.gitignore"

echo "Initializing gyat workspace..."
cd "$TEST_DIR"
$GYAT_BIN init
$GYAT_BIN track services/auth
$GYAT_BIN track services/api
$GYAT_BIN track services/web
$GYAT_BIN update

echo "# Auth Service (modified)" > "$TEST_DIR/auth/README.md"
echo "# API Service (modified)" > "$TEST_DIR/api/README.md"
echo "# Web Service (modified)" > "$TEST_DIR/web/README.md"

echo "Staging changes in gyat workspace..."
cd "$TEST_DIR"
$GYAT_BIN exec -- git add .
$GYAT_BIN list