#!/bin/bash
# Script to download certificates from Cadence repository
# Based on: https://github.com/cadence-workflow/cadence/tree/master/config/credentials

set -e

echo "Downloading certificates from Cadence repository..."

BASE_URL="https://raw.githubusercontent.com/cadence-workflow/cadence/master/config/credentials"

# Download all certificate files
echo "Downloading ca.cert..."
curl -L -o ca.cert "${BASE_URL}/ca.cert"

echo "Downloading server.cert..."
curl -L -o server.cert "${BASE_URL}/server.cert"

echo "Downloading server.key..."
curl -L -o server.key "${BASE_URL}/server.key"

echo "Downloading client.cert..."
curl -L -o client.cert "${BASE_URL}/client.cert"

echo "Downloading client.key..."
curl -L -o client.key "${BASE_URL}/client.key"

# Also try to get .pem versions if they exist
echo "Trying to download .pem versions..."
curl -L -o ca.pem "${BASE_URL}/ca.pem" 2>/dev/null || echo "ca.pem not found, skipping"
curl -L -o server.pem "${BASE_URL}/server.pem" 2>/dev/null || echo "server.pem not found, skipping"
curl -L -o client.pem "${BASE_URL}/client.pem" 2>/dev/null || echo "client.pem not found, skipping"

# Create symlinks for compatibility with existing code
if [ -f "client.cert" ]; then
    ln -sf client.cert client.crt 2>/dev/null || cp client.cert client.crt
fi

if [ -f "ca.cert" ]; then
    ln -sf ca.cert keytest.crt 2>/dev/null || cp ca.cert keytest.crt
fi

echo ""
echo "✓ Certificates downloaded successfully from Cadence repository!"
echo ""
echo "Downloaded files:"
ls -lh *.cert *.key 2>/dev/null || true
ls -lh *.pem 2>/dev/null || true
ls -lh *.crt 2>/dev/null || true
echo ""
echo "These are the official Cadence test certificates."
echo "Reference: https://github.com/cadence-workflow/cadence/tree/master/config/credentials"

