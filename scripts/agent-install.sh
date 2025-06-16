#!/bin/bash

set -e

COLLECTOR_NAME="ctrlb-collector"
VERSION="v1.0.0-alpha"
INSTALL_DIR="/usr/local/bin"
ENV_FILE="/etc/${COLLECTOR_NAME}/env"
SERVICE_FILE="/etc/systemd/system/${COLLECTOR_NAME}.service"

# Read from env or prompt interactively
BACKEND_URL="${BACKEND_URL:-}"
PIPELINE_NAME="${PIPELINE_NAME:-}"
STARTED_BY="${STARTED_BY:-}"

# Require root
if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run this script with sudo or as root."
  exit 1
fi

# Prompt if not provided
[ -z "$BACKEND_URL" ] && read -p "Enter backend URL: " BACKEND_URL
[ -z "$PIPELINE_NAME" ] && read -p "Enter pipeline name: " PIPELINE_NAME
[ -z "$STARTED_BY" ] && read -p "Enter started by (email): " STARTED_BY

# Validate required fields
if [[ -z "$BACKEND_URL" || -z "$PIPELINE_NAME" || -z "$STARTED_BY" ]]; then
  echo "❌ BACKEND_URL, PIPELINE_NAME, and STARTED_BY are required."
  exit 1
fi

# Detect arch/OS
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *)
    echo "❌ Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Download binary
DOWNLOAD_BASE_URL="https://github.com/ctrlb-hq/ctrlb-control-plane/releases/download/${VERSION}"
BINARY_URL="${DOWNLOAD_BASE_URL}/${COLLECTOR_NAME}-${OS}-${ARCH}"
BINARY_PATH="${INSTALL_DIR}/${COLLECTOR_NAME}"

echo "📥 Downloading ${COLLECTOR_NAME} ${VERSION} for ${OS}/${ARCH}..."
curl -L "$BINARY_URL" -o "$BINARY_PATH"
chmod +x "$BINARY_PATH"

# Write environment file
mkdir -p "$(dirname "$ENV_FILE")"
cat <<EOF > "$ENV_FILE"
BACKEND_URL=${BACKEND_URL}
PIPELINE_NAME=${PIPELINE_NAME}
STARTED_BY=${STARTED_BY}
EOF

chmod 600 "$ENV_FILE"

# Create systemd unit
cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=${COLLECTOR_NAME} Service
After=network.target

[Service]
ExecStart=${BINARY_PATH}
Restart=always
EnvironmentFile=${ENV_FILE}

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
echo "🔧 Configuring systemd service..."
systemctl daemon-reexec
systemctl enable "$COLLECTOR_NAME"
systemctl restart "$COLLECTOR_NAME"

echo "✅ ${COLLECTOR_NAME} installed and running!"
systemctl status "$COLLECTOR_NAME" --no-pager
