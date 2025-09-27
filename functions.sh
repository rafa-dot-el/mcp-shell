#!/bin/bash

# MCP Shell Development Helper Functions
# This file provides convenience functions for development.
# These helpers are for development convenience only - CI/CD must not depend on them.

# Project-specific aliases
alias ll='ls -la'
alias ..='cd ..'
alias gst='git status'
alias gcm='git commit -m'
alias gp='git push'

# Go development aliases
alias gor='go run'
alias gob='go build'
alias got='go test'
alias gom='go mod'
alias gof='go fmt'

# Development helper functions

build-watch() {
    echo "[*] Starting build watcher for Go files..."
    find . -name "*.go" | entr -c task build
}

test-watch() {
    echo "[*] Starting test watcher for Go files..."
    find . -name "*.go" | entr -c task test
}

lint-watch() {
    echo "[*] Starting lint watcher for Go files..."
    find . -name "*.go" | entr -c task lint
}

clean-all() {
    echo "[*] Cleaning all build artifacts..."
    task clean
    git clean -fdx -e .env -e .envrc -e functions.sh
    echo "[+] Clean complete"
}

quick-check() {
    echo "[*] Running quick development checks..."
    task lint && task test-unit
}

full-check() {
    echo "[*] Running full development pipeline..."
    task all
}

# Go-specific development shortcuts

go-deps() {
    echo "[*] Updating Go dependencies..."
    go mod download
    go mod tidy
    echo "[+] Dependencies updated"
}

go-coverage() {
    echo "[*] Running tests with coverage..."
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    echo "[+] Coverage report generated: coverage.html"
}

go-benchmark() {
    echo "[*] Running benchmarks..."
    go test -bench=. -benchmem ./...
}

go-race() {
    echo "[*] Running race condition tests..."
    go test -race ./...
}

go-profile() {
    echo "[*] Running with CPU profiling..."
    go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
    echo "[+] Profiles generated: cpu.prof, mem.prof"
    echo "[*] Analyze with: go tool pprof cpu.prof"
}

# Security and quality helpers

security-scan() {
    echo "[*] Running security scans..."
    gosec ./...
    trivy fs .
    echo "[+] Security scan complete"
}

vuln-check() {
    echo "[*] Checking for vulnerabilities..."
    govulncheck ./...
    echo "[+] Vulnerability check complete"
}

# Project setup helpers

dev-setup() {
    echo "[*] Setting up development environment..."
    echo "[*] Installing dependencies..."
    go mod download
    echo "[*] Running initial build..."
    task build
    echo "[*] Running tests..."
    task test
    echo "[+] Development setup complete"
}

release-prep() {
    echo "[*] Preparing for release..."
    echo "[*] Running full test suite..."
    task test
    echo "[*] Running security checks..."
    task vuln
    echo "[*] Running linting..."
    task lint
    echo "[*] Building release artifacts..."
    task build
    echo "[+] Release preparation complete"
}

# Documentation helpers

docs-serve() {
    echo "[*] Starting documentation server..."
    if [ -d "docs" ]; then
        cd docs && hugo serve
    else
        echo "[-] No docs directory found"
    fi
}

docs-build() {
    echo "[*] Building documentation..."
    task docs-build
}

# Version management helpers

version-info() {
    echo "[*] Current version information:"
    echo "Project version: $(cat VERSION 2>/dev/null || echo 'No VERSION file')"
    echo "Go version: $(go version)"
    echo "Git commit: $(git rev-parse --short HEAD 2>/dev/null || echo 'Not a git repo')"
    echo "Git branch: $(git branch --show-current 2>/dev/null || echo 'Not a git repo')"
}

# Git helpers

git-clean-branches() {
    echo "[*] Cleaning merged branches..."
    git branch --merged | grep -v "\*\|main\|master\|develop" | xargs -n 1 git branch -d
    echo "[+] Merged branches cleaned"
}

git-sync() {
    echo "[*] Syncing with upstream..."
    git fetch --all
    git pull origin $(git branch --show-current)
    echo "[+] Sync complete"
}

# Help function
dev-help() {
    echo "MCP Shell Development Helper Functions:"
    echo ""
    echo "Build & Test:"
    echo "  build-watch      - Watch for changes and rebuild"
    echo "  test-watch       - Watch for changes and retest"
    echo "  lint-watch       - Watch for changes and relint"
    echo "  quick-check      - Run lint + unit tests"
    echo "  full-check       - Run complete pipeline (task all)"
    echo ""
    echo "Go Specific:"
    echo "  go-deps          - Update Go dependencies"
    echo "  go-coverage      - Run tests with coverage report"
    echo "  go-benchmark     - Run benchmark tests"
    echo "  go-race          - Run race condition tests"
    echo "  go-profile       - Run with profiling"
    echo ""
    echo "Security:"
    echo "  security-scan    - Run security scanners"
    echo "  vuln-check       - Check for vulnerabilities"
    echo ""
    echo "Project Management:"
    echo "  dev-setup        - Initial development setup"
    echo "  release-prep     - Prepare for release"
    echo "  clean-all        - Clean all artifacts"
    echo ""
    echo "Documentation:"
    echo "  docs-serve       - Serve documentation locally"
    echo "  docs-build       - Build documentation"
    echo ""
    echo "Git & Version:"
    echo "  version-info     - Show version information"
    echo "  git-clean-branches - Clean merged branches"
    echo "  git-sync         - Sync with upstream"
    echo ""
    echo "Use 'task' to see available Taskfile commands"
}

# Display help on load
dev-help