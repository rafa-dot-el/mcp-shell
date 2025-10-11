#!/bin/bash
# System information script with format parameter

FORMAT="text"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --format)
            FORMAT="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

# Gather system information
OS=$(uname -s)
KERNEL=$(uname -r)
HOSTNAME=$(hostname)
UPTIME=$(uptime -p 2>/dev/null || uptime)
CPU_COUNT=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "unknown")
MEMORY=$(free -h 2>/dev/null | awk '/^Mem:/ {print $2}' || sysctl -n hw.memsize 2>/dev/null | awk '{print $1/1024/1024/1024 " GB"}' || echo "unknown")
LOAD=$(uptime | awk -F'load average:' '{print $2}' | xargs)

# Output in requested format
if [ "$FORMAT" = "json" ]; then
    cat <<EOF
{
  "operating_system": "$OS",
  "kernel": "$KERNEL",
  "hostname": "$HOSTNAME",
  "uptime": "$UPTIME",
  "cpu_count": "$CPU_COUNT",
  "memory": "$MEMORY",
  "load_average": "$LOAD"
}
EOF
else
    cat <<EOF
System Information
==================
Operating System: $OS
Kernel: $KERNEL
Hostname: $HOSTNAME
Uptime: $UPTIME
CPU Count: $CPU_COUNT
Memory: $MEMORY
Load Average: $LOAD
EOF
fi
