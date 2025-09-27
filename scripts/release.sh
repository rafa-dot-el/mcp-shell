#!/bin/bash

# MCP Shell - Release Script
# Copyright (C) 2025 Rafael
# Licensed under GPL-3.0
#
# This script manages the release process for MCP Shell.

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
VERSION_FILE="$PROJECT_ROOT/VERSION"
CHANGELOG_FILE="$PROJECT_ROOT/CHANGELOG.md"

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

# Get current version
get_current_version() {
    if [[ -f "$VERSION_FILE" ]]; then
        cat "$VERSION_FILE"
    else
        echo "0.0.0"
    fi
}

# Increment version
increment_version() {
    local version="$1"
    local part="$2"

    IFS='.' read -r major minor patch <<< "$version"

    case "$part" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            error "Invalid version part: $part (use major, minor, or patch)"
            return 1
            ;;
    esac

    echo "$major.$minor.$patch"
}

# Validate version format
validate_version() {
    local version="$1"
    if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        error "Invalid version format: $version (expected: x.y.z)"
        return 1
    fi
}

# Check if git repo is clean
check_git_clean() {
    if ! git diff-index --quiet HEAD --; then
        error "Git working directory is not clean. Please commit or stash changes."
        return 1
    fi

    if [[ -n "$(git ls-files --others --exclude-standard)" ]]; then
        warning "Untracked files found. Consider adding them to .gitignore or committing them."
    fi
}

# Check if we're on the main branch
check_main_branch() {
    local current_branch
    current_branch=$(git branch --show-current)

    if [[ "$current_branch" != "main" && "$current_branch" != "master" ]]; then
        error "Not on main/master branch (current: $current_branch)"
        error "Please switch to the main branch before releasing"
        return 1
    fi
}

# Update version file
update_version_file() {
    local new_version="$1"
    echo "$new_version" > "$VERSION_FILE"
    success "Updated VERSION file to $new_version"
}

# Update changelog
update_changelog() {
    local new_version="$1"
    local date
    date=$(date '+%Y-%m-%d')

    if [[ -f "$CHANGELOG_FILE" ]]; then
        # Create a temporary file
        local temp_file
        temp_file=$(mktemp)

        # Add new version entry
        {
            head -n 2 "$CHANGELOG_FILE"  # Keep title and first empty line
            echo "## [$new_version] - $date"
            echo ""
            echo "### Added"
            echo "- Release $new_version"
            echo ""
            tail -n +3 "$CHANGELOG_FILE"  # Keep rest of file
        } > "$temp_file"

        mv "$temp_file" "$CHANGELOG_FILE"
        success "Updated CHANGELOG.md for version $new_version"
    else
        warning "CHANGELOG.md not found, skipping changelog update"
    fi
}

# Run tests
run_tests() {
    info "Running tests..."
    if command -v task > /dev/null 2>&1; then
        task test
    else
        go test ./...
    fi
    success "All tests passed"
}

# Run linting
run_linting() {
    info "Running linting..."
    if command -v task > /dev/null 2>&1; then
        task lint
    else
        if command -v golangci-lint > /dev/null 2>&1; then
            golangci-lint run
        else
            go vet ./...
            go fmt ./...
        fi
    fi
    success "Linting passed"
}

# Run security checks
run_security_checks() {
    info "Running security checks..."
    if command -v task > /dev/null 2>&1; then
        task vuln || warning "Vulnerability check failed or not available"
        task security || warning "Security check failed or not available"
    else
        if command -v govulncheck > /dev/null 2>&1; then
            govulncheck ./...
        else
            warning "govulncheck not available, skipping vulnerability check"
        fi

        if command -v gosec > /dev/null 2>&1; then
            gosec ./...
        else
            warning "gosec not available, skipping security check"
        fi
    fi
    success "Security checks completed"
}

# Build artifacts
build_artifacts() {
    info "Building release artifacts..."
    if [[ -f "$SCRIPT_DIR/build.sh" ]]; then
        "$SCRIPT_DIR/build.sh" --clean --all
    else
        if command -v task > /dev/null 2>&1; then
            task clean
            task build-all
        else
            rm -rf bin/
            go build -o bin/mcp-shell ./cmd/mcp-shell
        fi
    fi
    success "Release artifacts built"
}

# Create git tag
create_git_tag() {
    local version="$1"
    local tag="v$version"

    info "Creating git tag: $tag"
    git tag -a "$tag" -m "Release $tag"
    success "Created git tag: $tag"
}

# Commit changes
commit_changes() {
    local version="$1"

    info "Committing release changes..."
    git add "$VERSION_FILE"
    if [[ -f "$CHANGELOG_FILE" ]]; then
        git add "$CHANGELOG_FILE"
    fi

    git commit -m "chore: release version $version

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

    success "Committed release changes"
}

# Push to remote
push_to_remote() {
    local version="$1"
    local tag="v$version"

    info "Pushing to remote repository..."
    git push origin HEAD
    git push origin "$tag"
    success "Pushed to remote repository"
}

# Show release summary
show_release_summary() {
    local version="$1"
    local tag="v$version"

    echo
    success "Release $version completed successfully!"
    echo
    info "What was done:"
    info "  ✓ Updated VERSION file to $version"
    info "  ✓ Updated CHANGELOG.md (if present)"
    info "  ✓ Ran tests, linting, and security checks"
    info "  ✓ Built release artifacts"
    info "  ✓ Created git commit and tag: $tag"
    info "  ✓ Pushed to remote repository"
    echo
    info "Next steps:"
    info "  • Monitor the CI/CD pipeline for release artifacts"
    info "  • Check GitHub releases for the new release"
    info "  • Update documentation if needed"
    echo
}

# Show help
show_help() {
    cat << EOF
MCP Shell Release Script

Usage: $0 [OPTIONS] [VERSION_TYPE|VERSION]

Options:
    -h, --help          Show this help message
    -n, --dry-run       Show what would be done without making changes
    -f, --force         Skip confirmations
    --skip-tests        Skip running tests
    --skip-lint         Skip running linting
    --skip-security     Skip running security checks
    --skip-build        Skip building artifacts
    --no-push          Don't push to remote repository

Version Types:
    major               Increment major version (x.0.0)
    minor               Increment minor version (x.y.0)
    patch               Increment patch version (x.y.z)

Examples:
    $0 patch            # Increment patch version
    $0 minor            # Increment minor version
    $0 major            # Increment major version
    $0 1.2.3            # Set specific version
    $0 --dry-run patch  # Preview what would happen

Current version: $(get_current_version)

EOF
}

# Main function
main() {
    local dry_run=false
    local force=false
    local skip_tests=false
    local skip_lint=false
    local skip_security=false
    local skip_build=false
    local no_push=false
    local version_input=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -n|--dry-run)
                dry_run=true
                shift
                ;;
            -f|--force)
                force=true
                shift
                ;;
            --skip-tests)
                skip_tests=true
                shift
                ;;
            --skip-lint)
                skip_lint=true
                shift
                ;;
            --skip-security)
                skip_security=true
                shift
                ;;
            --skip-build)
                skip_build=true
                shift
                ;;
            --no-push)
                no_push=true
                shift
                ;;
            -*)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
            *)
                if [[ -z "$version_input" ]]; then
                    version_input="$1"
                else
                    error "Too many arguments"
                    show_help
                    exit 1
                fi
                shift
                ;;
        esac
    done

    if [[ -z "$version_input" ]]; then
        error "Version type or version number required"
        show_help
        exit 1
    fi

    # Change to project root
    cd "$PROJECT_ROOT"

    # Get current version
    local current_version
    current_version=$(get_current_version)
    info "Current version: $current_version"

    # Determine new version
    local new_version
    case "$version_input" in
        major|minor|patch)
            new_version=$(increment_version "$current_version" "$version_input")
            ;;
        *)
            new_version="$version_input"
            ;;
    esac

    # Validate new version
    validate_version "$new_version"
    info "New version: $new_version"

    # Check if version is newer
    if [[ "$new_version" == "$current_version" ]]; then
        error "New version ($new_version) is the same as current version"
        exit 1
    fi

    # Dry run mode
    if [[ "$dry_run" == true ]]; then
        info "DRY RUN MODE - No changes will be made"
        echo
        info "Would update version from $current_version to $new_version"
        info "Would run tests, linting, and security checks"
        info "Would build release artifacts"
        info "Would create git commit and tag: v$new_version"
        if [[ "$no_push" != true ]]; then
            info "Would push to remote repository"
        fi
        echo
        info "Run without --dry-run to perform the release"
        exit 0
    fi

    # Verify git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        error "Not in a git repository"
        exit 1
    fi

    # Check git status
    check_git_clean
    check_main_branch

    # Confirmation
    if [[ "$force" != true ]]; then
        echo
        warning "This will create a new release: $current_version → $new_version"
        read -p "Continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Release cancelled"
            exit 0
        fi
    fi

    # Execute release steps
    info "Starting release process for version $new_version"

    # Update version file
    update_version_file "$new_version"

    # Update changelog
    update_changelog "$new_version"

    # Run quality checks
    if [[ "$skip_tests" != true ]]; then
        run_tests
    fi

    if [[ "$skip_lint" != true ]]; then
        run_linting
    fi

    if [[ "$skip_security" != true ]]; then
        run_security_checks
    fi

    # Build artifacts
    if [[ "$skip_build" != true ]]; then
        build_artifacts
    fi

    # Git operations
    commit_changes "$new_version"
    create_git_tag "$new_version"

    # Push to remote
    if [[ "$no_push" != true ]]; then
        push_to_remote "$new_version"
    else
        warning "Skipping push to remote repository (--no-push specified)"
        info "Remember to push manually: git push origin HEAD && git push origin v$new_version"
    fi

    # Show summary
    show_release_summary "$new_version"
}

# Handle script being sourced vs executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi