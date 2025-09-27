#!/bin/bash

# MCP Shell - Build Script
# Copyright (C) 2025 Rafael
# Licensed under GPL-3.0
#
# This script builds the MCP Shell binary with proper version information.

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
BINARY_NAME="mcp-shell"
VERSION_FILE="$PROJECT_ROOT/VERSION"
OUTPUT_DIR="$PROJECT_ROOT/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Get version from VERSION file or default
get_version() {
    if [[ -f "$VERSION_FILE" ]]; then
        cat "$VERSION_FILE"
    else
        echo "dev"
    fi
}

# Get git information
get_git_info() {
    local commit=""
    local branch=""
    local dirty=""

    if git rev-parse --git-dir > /dev/null 2>&1; then
        commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        branch=$(git branch --show-current 2>/dev/null || echo "unknown")

        if ! git diff-index --quiet HEAD -- 2>/dev/null; then
            dirty="-dirty"
        fi
    else
        commit="unknown"
        branch="unknown"
    fi

    echo "$commit $branch$dirty"
}

# Build function
build_binary() {
    local version="$1"
    local goos="${2:-$(go env GOOS)}"
    local goarch="${3:-$(go env GOARCH)}"

    local git_info
    git_info=$(get_git_info)
    local commit
    commit=$(echo "$git_info" | cut -d' ' -f1)
    local branch
    branch=$(echo "$git_info" | cut -d' ' -f2)

    local build_date
    build_date=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

    local output_name="$BINARY_NAME"
    if [[ "$goos" == "windows" ]]; then
        output_name="${output_name}.exe"
    fi

    local output_path
    if [[ "$goos" != "$(go env GOOS)" || "$goarch" != "$(go env GOARCH)" ]]; then
        output_path="$OUTPUT_DIR/${BINARY_NAME}-${goos}-${goarch}"
        if [[ "$goos" == "windows" ]]; then
            output_path="${output_path}.exe"
        fi
    else
        output_path="$OUTPUT_DIR/$output_name"
    fi

    # Ensure output directory exists
    mkdir -p "$OUTPUT_DIR"

    info "Building ${BINARY_NAME} for ${goos}/${goarch}..."
    info "Version: $version"
    info "Commit: $commit"
    info "Branch: $branch"
    info "Build Date: $build_date"
    info "Output: $output_path"

    # Build with version information
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build \
        -ldflags "-s -w \
                  -X main.version=$version \
                  -X main.commit=$commit \
                  -X main.branch=$branch \
                  -X main.buildDate=$build_date \
                  -X main.builtBy=build-script" \
        -o "$output_path" \
        "$PROJECT_ROOT/cmd/$BINARY_NAME"

    if [[ -f "$output_path" ]]; then
        success "Built: $output_path"

        # Show file info
        if command -v file > /dev/null 2>&1; then
            info "File info: $(file "$output_path")"
        fi

        # Show size
        if command -v du > /dev/null 2>&1; then
            info "Size: $(du -h "$output_path" | cut -f1)"
        elif command -v ls > /dev/null 2>&1; then
            info "Size: $(ls -lh "$output_path" | awk '{print $5}')"
        fi
    else
        error "Build failed: output file not found"
        return 1
    fi
}

# Show help
show_help() {
    cat << EOF
MCP Shell Build Script

Usage: $0 [OPTIONS] [GOOS] [GOARCH]

Options:
    -h, --help      Show this help message
    -v, --version   Show version and exit
    -c, --clean     Clean build artifacts before building
    -a, --all       Build for all supported platforms

Examples:
    $0                      # Build for current platform
    $0 linux amd64         # Build for Linux AMD64
    $0 windows amd64       # Build for Windows AMD64
    $0 darwin arm64        # Build for macOS ARM64
    $0 --all               # Build for all platforms
    $0 --clean             # Clean and build

Supported platforms:
    linux/amd64, linux/arm64, linux/386, linux/arm
    darwin/amd64, darwin/arm64
    windows/amd64, windows/386
    freebsd/amd64, freebsd/arm64

EOF
}

# Clean build artifacts
clean_build() {
    info "Cleaning build artifacts..."
    if [[ -d "$OUTPUT_DIR" ]]; then
        rm -rf "$OUTPUT_DIR"
        success "Cleaned: $OUTPUT_DIR"
    else
        info "Nothing to clean"
    fi
}

# Build for all platforms
build_all() {
    local version
    version=$(get_version)

    local platforms=(
        "linux/amd64"
        "linux/arm64"
        "linux/386"
        "linux/arm"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
        "windows/386"
        "freebsd/amd64"
        "freebsd/arm64"
    )

    info "Building for all platforms..."

    for platform in "${platforms[@]}"; do
        IFS='/' read -r goos goarch <<< "$platform"
        if ! build_binary "$version" "$goos" "$goarch"; then
            error "Failed to build for $platform"
            return 1
        fi
        echo # Empty line for readability
    done

    success "All builds completed successfully!"
}

# Main function
main() {
    local clean=false
    local build_all_platforms=false
    local goos=""
    local goarch=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                echo "Build script version: 1.0.0"
                exit 0
                ;;
            -c|--clean)
                clean=true
                shift
                ;;
            -a|--all)
                build_all_platforms=true
                shift
                ;;
            -*)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
            *)
                if [[ -z "$goos" ]]; then
                    goos="$1"
                elif [[ -z "$goarch" ]]; then
                    goarch="$1"
                else
                    error "Too many arguments"
                    show_help
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Change to project root
    cd "$PROJECT_ROOT"

    # Verify we're in a Go module
    if [[ ! -f "go.mod" ]]; then
        error "Not in a Go module directory (go.mod not found)"
        exit 1
    fi

    # Clean if requested
    if [[ "$clean" == true ]]; then
        clean_build
    fi

    # Build
    if [[ "$build_all_platforms" == true ]]; then
        build_all
    else
        local version
        version=$(get_version)
        build_binary "$version" "$goos" "$goarch"
    fi

    success "Build completed successfully!"
}

# Handle script being sourced vs executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi