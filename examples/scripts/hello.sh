#!/bin/bash
# Simple hello world script demonstrating parameter handling

NAME="${1:-World}"

echo "Hello, ${NAME}!"
echo "This is an example script from MCP Shell"
echo "Current time: $(date)"
echo "Running on: $(hostname)"
