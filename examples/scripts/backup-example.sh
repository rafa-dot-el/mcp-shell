#!/bin/bash
# Example backup script demonstrating parameter validation

SOURCE=""
DESTINATION=""
COMPRESS="true"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --source)
            SOURCE="$2"
            shift 2
            ;;
        --dest)
            DESTINATION="$2"
            shift 2
            ;;
        --compress)
            COMPRESS="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

# Validate required parameters
if [ -z "$SOURCE" ]; then
    echo "Error: --source parameter is required" >&2
    exit 1
fi

if [ -z "$DESTINATION" ]; then
    echo "Error: --dest parameter is required" >&2
    exit 1
fi

# Validate source exists
if [ ! -d "$SOURCE" ]; then
    echo "Error: Source directory '$SOURCE' does not exist" >&2
    exit 1
fi

# Create destination directory if it doesn't exist
mkdir -p "$DESTINATION" 2>/dev/null

# Perform backup (simulation)
echo "[*] Starting backup..."
echo "    Source: $SOURCE"
echo "    Destination: $DESTINATION"
echo "    Compression: $COMPRESS"
echo ""

# Simulate backup process
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="backup_${TIMESTAMP}"

if [ "$COMPRESS" = "true" ]; then
    BACKUP_FILE="${DESTINATION}/${BACKUP_NAME}.tar.gz"
    echo "[*] Creating compressed backup: $BACKUP_FILE"
    echo "[*] This is a simulation - not actually creating archive"
    # In a real script: tar -czf "$BACKUP_FILE" -C "$(dirname "$SOURCE")" "$(basename "$SOURCE")"
else
    BACKUP_DIR="${DESTINATION}/${BACKUP_NAME}"
    echo "[*] Creating uncompressed backup: $BACKUP_DIR"
    echo "[*] This is a simulation - not actually copying files"
    # In a real script: cp -r "$SOURCE" "$BACKUP_DIR"
fi

echo ""
echo "[+] Backup completed successfully"
echo "    This was a simulation - no files were actually modified"
