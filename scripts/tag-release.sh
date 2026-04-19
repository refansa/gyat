#!/bin/bash
# Create a new git tag with an incremented version number.
# Automatically reads the current version from existing tags or cmd/root.go.

set -e

usage() {
    echo "Usage: $0 [patch|minor|major] [version]"
    echo ""
    echo "Options:"
    echo "  patch  Increment the patch version (default)"
    echo "  minor  Increment the minor version"
    echo "  major  Increment the major version"
    echo ""
    echo "Examples:"
    echo "  $0 patch        # v0.2.0 -> v0.2.1"
    echo "  $0 minor        # v0.2.0 -> v0.3.0"
    echo "  $0 major        # v0.2.0 -> v1.0.0"
    echo "  $0 patch v0.3.0 # Set to specific version"
    exit 1
}

# Get current version from latest tag or from cmd/root.go
get_current_version() {
    local version

    # Try to get version from latest tag
    version=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')
    if [ -n "$version" ]; then
        echo "$version"
        return
    fi

    # Fall back to reading from cmd/root.go
    version=$(grep -m1 'var Version = "' cmd/root.go | sed 's/.*var Version = "\(.*\)".*/\1/')
    if [ -n "$version" ] && [ "$version" != "dev" ]; then
        echo "$version"
        return
    fi

    # Default starting version
    echo "0.1.0"
}

# Parse version components
parse_version() {
    local version=$1
    major=$(echo "$version" | cut -d. -f1)
    minor=$(echo "$version" | cut -d. -f2)
    patch=$(echo "$version" | cut -d. -f3)
}

# Increment version
increment_version() {
    local part=$1

    case "$part" in
        major) ((major++)); minor=0; patch=0 ;;
        minor) ((minor++)); patch=0 ;;
        patch) ((patch++)) ;;
    esac

    new_version="${major}.${minor}.${patch}"
}

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Fetch latest tags from remote to ensure we have the latest version
echo "Fetching latest tags from origin..."
git fetch --tags origin

# Parse arguments
PART="patch"
SPECIFIC_VERSION=""

while [ $# -gt 0 ]; do
    case "$1" in
        patch|minor|major)
            PART="$1"
            ;;
        -h|--help)
            usage
            ;;
        *)
            if [[ "$1" == v*.*.* ]]; then
                SPECIFIC_VERSION="${1#v}"
            else
                SPECIFIC_VERSION="$1"
            fi
            ;;
    esac
    shift
done

# Get current version and compute new version
if [ -n "$SPECIFIC_VERSION" ]; then
    new_version="$SPECIFIC_VERSION"
else
    current_version=$(get_current_version)
    parse_version "$current_version"
    increment_version "$PART"
fi

echo "Current version: v$current_version"
echo "New version:     v$new_version"
echo ""

read -p "Create tag v$new_version? [y/N] " confirm
if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "Aborted."
    exit 1
fi

# Create the tag
git tag -a "v$new_version" -m "Release v$new_version"

echo "Created tag v$new_version"
echo ""
echo "Push with: git push origin v$new_version"