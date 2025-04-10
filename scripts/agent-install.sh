#!/bin/bash

set -e

AGENT_NAME="ctrlb-agent"
VERSION="v1.0.0"
INSTALL_DIR="/usr/local/bin"
ENV_FILE="/etc/${AGENT_NAME}/env"
SERVICE_FILE="/etc/systemd/system/${AGENT_NAME}.service"

# Defaults (can be overridden by args)
BACKEND_URL=""
SECRET_KEY=""

# Usage
usage() {
  echo "Usage: sudo $0 --backend-url <url> --secret-key <key>"
  echo "You can also omit these arguments to input them interactively."
}

# Require root
if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run this script with sudo or as root."
  usage
  exit 1
fi

# Parse CLI args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --backend-url)
      BACKEND_URL="$2"
      shift 2
      ;;
    --secret-key)
      SECRET_KEY="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      usage
      exit 1
      ;;
  esac
done

# Prompt if values not passed
if [ -z "$BACKEND_URL" ]; then
  read -p "Enter backend URL (e.g. https://api.yourdomain.com): " BACKEND_URL
fi

if [ -z "$SECRET_KEY" ]; then
  read -p "Enter secret key: " SECRET_KEY
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

# Construct download URL *after* arch is known
DOWNLOAD_BASE_URL="https://github.com/ctrlb-hq/ctrlb-control-plane/releases/download/${VERSION}"
BINARY_URL="${DOWNLOAD_BASE_URL}/${AGENT_NAME}-${OS}-${ARCH}"
BINARY_PATH="${INSTALL_DIR}/${AGENT_NAME}"

echo "📥 Downloading ${AGENT_NAME} ${VERSION} for ${OS}/${ARCH}..."
curl -L "$BINARY_URL" -o "$BINARY_PATH"
chmod +x "$BINARY_PATH"

# Write environment variables
mkdir -p "$(dirname "$ENV_FILE")"
cat <<EOF > "$ENV_FILE"
BACKEND_URL=${BACKEND_URL}
SECRET_KEY=${SECRET_KEY}
EOF

chmod 600 "$ENV_FILE"

# Create systemd unit
cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=${AGENT_NAME} Service
After=network.target

[Service]
ExecStart=${BINARY_PATH}
Restart=always
EnvironmentFile=${ENV_FILE}

[Install]
WantedBy=multi-user.target
EOF

# Enable and start systemd service
echo "🔧 Configuring systemd service..."
systemctl daemon-reexec
systemctl enable "$AGENT_NAME"
systemctl restart "$AGENT_NAME"

echo "✅ ${AGENT_NAME} installed and running!"
systemctl status "$AGENT_NAME" --no-pager
