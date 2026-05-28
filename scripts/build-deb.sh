#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Determine version from git tag or default
GIT_DESC=$(git -C "$PROJECT_ROOT" describe --tags 2>/dev/null || echo "")
if echo "$GIT_DESC" | grep -q '^v\?[0-9]\+\.[0-9]'; then
    VERSION=$(echo "$GIT_DESC" | sed 's/^v//')
else
    VERSION="0.1.0"
fi
ARCH="amd64"
PACKAGE_NAME="beep-analytics"
STAGING="$PROJECT_ROOT/build/${PACKAGE_NAME}_${VERSION}_${ARCH}"

echo "Building ${PACKAGE_NAME} ${VERSION}..."

# Clean previous build
rm -rf "$PROJECT_ROOT/build"
mkdir -p "$STAGING"

# Build production binary
echo "Compiling binary..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o "$STAGING/usr/bin/beep-analytics" "$PROJECT_ROOT/cmd/beep-analytics"

# Copy files to staging
echo "Assembling package..."

# Scripts
mkdir -p "$STAGING/usr/lib/beep-analytics"
cp "$PROJECT_ROOT/scripts/update-geoip.sh" "$STAGING/usr/lib/beep-analytics/"
cp "$PROJECT_ROOT/scripts/backup.sh" "$STAGING/usr/lib/beep-analytics/"

# Systemd service
mkdir -p "$STAGING/etc/systemd/system"
	cp "$SCRIPT_DIR/beep-analytics.service" "$STAGING/etc/systemd/system/"

# Documentation
mkdir -p "$STAGING/usr/share/doc/beep-analytics"
cp "$PROJECT_ROOT/README.md" "$STAGING/usr/share/doc/beep-analytics/"
cp "$PROJECT_ROOT/LICENSE" "$STAGING/usr/share/doc/beep-analytics/"

# Debian control files
mkdir -p "$STAGING/DEBIAN"

# Generate control file with version
INSTALLED_SIZE=$(du -sk "$STAGING" | cut -f1)
sed -e "s/^Version: .*/Version: $VERSION/" \
    -e "/^Description:/i Installed-Size: $INSTALLED_SIZE" \
    "$PROJECT_ROOT/debian/control" > "$STAGING/DEBIAN/control"
cp "$PROJECT_ROOT/debian/postinst" "$STAGING/DEBIAN/"
cp "$PROJECT_ROOT/debian/prerm" "$STAGING/DEBIAN/"
cp "$PROJECT_ROOT/debian/postrm" "$STAGING/DEBIAN/"

# Set permissions
chmod 755 "$STAGING/DEBIAN/postinst"
chmod 755 "$STAGING/DEBIAN/prerm"
chmod 755 "$STAGING/usr/bin/beep-analytics"
chmod 755 "$STAGING/usr/lib/beep-analytics/update-geoip.sh"
chmod 755 "$STAGING/usr/lib/beep-analytics/backup.sh"

# Build the deb package
echo "Creating .deb package..."
dpkg-deb --build "$STAGING" "$PROJECT_ROOT/build/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"

echo "Done: build/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
