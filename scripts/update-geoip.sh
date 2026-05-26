#!/bin/bash
set -e

GEOIP_ACCOUNT_ID="${GEOIP_ACCOUNT_ID:?Set GEOIP_ACCOUNT_ID}"
GEOIP_LICENSE_KEY="${GEOIP_LICENSE_KEY:?Set GEOIP_LICENSE_KEY}"
DOWNLOAD_URL="https://download.maxmind.com/geoip/databases/GeoLite2-City-CSV/download?suffix=zip"
TMP_DIR=$(mktemp -d)

cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

# Download
wget --content-disposition \
    --user="$GEOIP_ACCOUNT_ID" \
    --password="$GEOIP_LICENSE_KEY" \
    -P "$TMP_DIR" \
    "$DOWNLOAD_URL"

# Find the downloaded zip
ZIP_FILE=$(find "$TMP_DIR" -name "GeoLite2-City-CSV_*.zip" | head -1)

if [ -z "$ZIP_FILE" ]; then
    echo "Error: download failed" >&2
    exit 1
fi

# Extract
unzip "$ZIP_FILE" -d "$TMP_DIR"

# Find the extracted directory
EXTRACTED_DIR=$(find "$TMP_DIR" -name "GeoLite2-City-CSV_*" -type d | head -1)

if [ -z "$EXTRACTED_DIR" ]; then
    echo "Error: could not find extracted directory" >&2
    exit 1
fi

# Stop service
sudo systemctl stop beep

# Update files
# Locations file
sudo cp "$EXTRACTED_DIR/GeoLite2-City-Locations-en.csv" /usr/share/GeoIP/

# Combine IPv4 and IPv6 blocks into single file (preserve header from one file)
head -1 "$EXTRACTED_DIR/GeoLite2-City-Blocks-IPv4.csv" | sudo tee /usr/share/GeoIP/GeoLite2-City-Blocks.csv > /dev/null
tail -n +2 "$EXTRACTED_DIR/GeoLite2-City-Blocks-IPv4.csv" | sudo tee -a /usr/share/GeoIP/GeoLite2-City-Blocks.csv > /dev/null
tail -n +2 "$EXTRACTED_DIR/GeoLite2-City-Blocks-IPv6.csv" | sudo tee -a /usr/share/GeoIP/GeoLite2-City-Blocks.csv > /dev/null

# Start service
sudo systemctl start beep

echo "GeoIP database updated successfully"
