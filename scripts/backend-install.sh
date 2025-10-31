#!/bin/bash

set -e

BACKEND_NAME="control-plane-backend"
VERSION="v1.0.0"
INSTALL_DIR="/opt/ctrlb/control-plane-backend"
ENV_FILE="${INSTALL_DIR}/.env"
SERVICE_FILE="/etc/systemd/system/${BACKEND_NAME}.service"

# Defaults (can be overridden)
PORT="8096"
ENVIRONMENT="prod"
WORKER_COUNT="4"
CHECK_INTERVAL_MINS="10"
JWT_SECRET=""

usage() {
  echo "Usage: sudo $0 --jwt-secret <secret>"
  echo "Optional:"
  echo "  --port <port> (default: 8096)"
  echo "  --env <env> (default: prod)"
  echo "  --workers <count> (default: 4)"
  echo "  --check-interval <mins> (default: 10)"
}

# Require root
if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run as root or with sudo"
  usage
  exit 1
fi

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --jwt-secret)
      JWT_SECRET="$2"
      shift 2
      ;;
    --port)
      PORT="$2"
      shift 2
      ;;
    --env)
      ENVIRONMENT="$2"
      shift 2
      ;;
    --workers)
      WORKER_COUNT="$2"
      shift 2
      ;;
    --check-interval)
      CHECK_INTERVAL_MINS="$2"
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

# Prompt if JWT_SECRET is missing
if [ -z "$JWT_SECRET" ]; then
  read -p "Enter JWT secret key: " JWT_SECRET
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

# Set download URL after detecting ARCH/OS
DOWNLOAD_BASE_URL="https://github.com/ctrlb-hq/ctrlb-control-plane/releases/download/${VERSION}"
BINARY_URL="${DOWNLOAD_BASE_URL}/${BACKEND_NAME}-${OS}-${ARCH}"
BINARY_PATH="${INSTALL_DIR}/${BACKEND_NAME}"

mkdir -p "$INSTALL_DIR"

echo "📥 Downloading ${BACKEND_NAME} ${VERSION} for ${OS}/${ARCH}..."
if ! curl -fL "$BINARY_URL" -o "$BINARY_PATH"; then
  echo "❌ Failed to download binary from $BINARY_URL"
  exit 1
fi

chmod +x "$BINARY_PATH"

# Verify it is a valid ELF binary (or Mach-O if on macOS)
if ! file "$BINARY_PATH" | grep -q 'ELF'; then
  echo "❌ Downloaded file is not a valid executable (check architecture or release asset)"
  rm -f "$BINARY_PATH"
  exit 1
fi


# Write env vars to file
mkdir -p "$(dirname "$ENV_FILE")"
cat <<EOF > "$ENV_FILE"
PORT=${PORT}
ENV=${ENVIRONMENT}
WORKER_COUNT=${WORKER_COUNT}
CHECK_INTERVAL_MINS=${CHECK_INTERVAL_MINS}
JWT_SECRET=${JWT_SECRET}
EOF

chmod 600 "$ENV_FILE"

# Create systemd unit
cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=${BACKEND_NAME} Service
After=network.target

[Service]
ExecStart=${BINARY_PATH}
Restart=always
EnvironmentFile=${ENV_FILE}
WorkingDirectory=/var/lib/${BACKEND_NAME}

[Install]
WantedBy=multi-user.target
EOF

# Create working directory if not present
mkdir -p /var/lib/${BACKEND_NAME}

echo "🔧 Enabling systemd service..."
sudo systemctl daemon-reexec
sudo systemctl enable "$BACKEND_NAME"
sudo systemctl restart "$BACKEND_NAME"

echo "✅ ${BACKEND_NAME} is running on port ${PORT}"
sudo systemctl status "$BACKEND_NAME" --no-pager
