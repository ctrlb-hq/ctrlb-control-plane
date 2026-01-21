#!/bin/bash

set -e

COLLECTOR_NAME="ctrlb-collector"
COLLECTOR_VERSION="v0.2.1"
INSTALL_DIR="/opt/ctrlb/${COLLECTOR_NAME}"
ENV_FILE="${INSTALL_DIR}/.env"
CONFIG_FILE="${INSTALL_DIR}/config.yaml"
SERVICE_FILE="/etc/systemd/system/${COLLECTOR_NAME}.service"

# Fluent Bit configuration (will be installed via package manager)
FLUENTBIT_BIN="/opt/fluent-bit/bin/fluent-bit"

# Read from env or prompt interactively
BACKEND_URL="${BACKEND_URL:-}"
PIPELINE_NAME="${PIPELINE_NAME:-}"
STARTED_BY="${STARTED_BY:-}"
LOG_PATHS="${LOG_PATHS:-/var/log/*.log}"  # Default log paths to monitor

# Require root
if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run this script with sudo or as root."
  exit 1
fi

# Prompt if not provided
[ -z "$BACKEND_URL" ] && read -p "Enter backend URL: " BACKEND_URL
[ -z "$PIPELINE_NAME" ] && read -p "Enter pipeline name: " PIPELINE_NAME
[ -z "$STARTED_BY" ] && read -p "Enter started by (email): " STARTED_BY
read -p "Enter log paths to monitor (default: /var/log/*.log): " USER_LOG_PATHS
[ -n "$USER_LOG_PATHS" ] && LOG_PATHS="$USER_LOG_PATHS"

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
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "❌ Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

echo "🚀 Installing CtrlB Collector with Fluent Bit backend..."
echo ""
echo "📦 Configuration:"
echo "   • Architecture: ${OS}/${ARCH}"
echo "   • Backend URL: ${BACKEND_URL}"
echo "   • Pipeline: ${PIPELINE_NAME}"
echo "   • Log Paths: ${LOG_PATHS}"
echo ""
echo "🏗️  Architecture:"
echo "   Log Files → Fluent Bit (embedded) → CtrlB Collector → Backend"
echo "   Single systemd service manages both components"
echo ""

# Check if Fluent Bit is already installed
if [ -f "$FLUENTBIT_BIN" ]; then
  echo "✅ Fluent Bit is already installed"
else
  echo "❌ Fluent Bit is not installed"
fi

# ========================
# Install Fluent Bit from Official Repository
# ========================
echo "📦 Installing Fluent Bit from official repository..."

# Detect distribution
if [ -f /etc/os-release ]; then
  . /etc/os-release
  DISTRO=$ID
  DISTRO_VERSION=$VERSION_ID
  echo "   Detected: ${PRETTY_NAME:-$ID $VERSION_ID}"
else
  echo "❌ Cannot detect OS distribution"
  exit 1
fi

case "$DISTRO" in
  ubuntu)
    echo "   Installing on Ubuntu..."
    
    # Add Fluent Bit GPG key
    curl -fsSL https://packages.fluentbit.io/fluentbit.key | gpg --dearmor | sudo tee /usr/share/keyrings/fluentbit-keyring.gpg > /dev/null
    
    # Get codename
    codename=$(grep -oP '(?<=VERSION_CODENAME=).*' /etc/os-release 2>/dev/null || lsb_release -cs 2>/dev/null)
    
    # Add Fluent Bit repository
    echo "deb [signed-by=/usr/share/keyrings/fluentbit-keyring.gpg] https://packages.fluentbit.io/ubuntu/$codename $codename main" | sudo tee /etc/apt/sources.list.d/fluent-bit.list > /dev/null
    
    # Update and install
    sudo apt-get update -qq
    sudo apt-get install -y fluent-bit
    
    # Stop the default service (we'll manage it through our collector)
    sudo systemctl stop fluent-bit 2>/dev/null || true
    sudo systemctl disable fluent-bit 2>/dev/null || true
    ;;
    
  debian)
    echo "   Installing on Debian..."
    
    # Add Fluent Bit GPG key
    curl -fsSL https://packages.fluentbit.io/fluentbit.key | gpg --dearmor | sudo tee /usr/share/keyrings/fluentbit-keyring.gpg > /dev/null
    
    # Get codename
    codename=$(grep -oP '(?<=VERSION_CODENAME=).*' /etc/os-release 2>/dev/null || lsb_release -cs 2>/dev/null)
    
    # Add Fluent Bit repository
    echo "deb [signed-by=/usr/share/keyrings/fluentbit-keyring.gpg] https://packages.fluentbit.io/debian/$codename $codename main" | sudo tee /etc/apt/sources.list.d/fluent-bit.list > /dev/null
    
    # Update and install
    sudo apt-get update -qq
    sudo apt-get install -y fluent-bit
    
    # Stop the default service (we'll manage it through our collector)
    sudo systemctl stop fluent-bit 2>/dev/null || true
    sudo systemctl disable fluent-bit 2>/dev/null || true
    ;;
    
  centos|rhel|rocky|almalinux|fedora)
    echo "   Installing on RHEL/CentOS/Rocky/Alma..."
    
    # Determine the major version
    MAJOR_VERSION=$(echo "$DISTRO_VERSION" | cut -d. -f1)
    
    # Handle CentOS 8 EOL
    if [ "$DISTRO" = "centos" ] && [ "$MAJOR_VERSION" = "8" ]; then
      echo "   Configuring CentOS 8 vault mirrors..."
      sudo sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-* 2>/dev/null || true
      sudo sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-* 2>/dev/null || true
    fi
    
    # Create Fluent Bit repo file
    cat <<EOF | sudo tee /etc/yum.repos.d/fluent-bit.repo > /dev/null
[fluent-bit]
name = Fluent Bit
baseurl = https://packages.fluentbit.io/centos/\$releasever/\$basearch/
gpgcheck=1
gpgkey=https://packages.fluentbit.io/fluentbit.key
repo_gpgcheck=1
enabled=1
EOF
    
    # Install Fluent Bit
    sudo yum install -y fluent-bit
    
    # Stop the default service (we'll manage it through our collector)
    sudo systemctl stop fluent-bit 2>/dev/null || true
    sudo systemctl disable fluent-bit 2>/dev/null || true
    ;;
    
  *)
    echo "❌ Unsupported distribution: $DISTRO"
    echo "   Supported: Ubuntu, Debian, CentOS, RHEL, Rocky Linux, AlmaLinux, Fedora"
    exit 1
    ;;
esac

# Verify Fluent Bit binary exists
if [ ! -f "$FLUENTBIT_BIN" ]; then
  echo "❌ Fluent Bit binary not found at ${FLUENTBIT_BIN}"
  echo "   Installation may have failed. Check the logs above."
  exit 1
fi

echo "✅ Fluent Bit installed successfully"

# ========================
# Download CtrlB Collector Binary
# ========================
echo "📥 Downloading ${COLLECTOR_NAME} ${COLLECTOR_VERSION} for ${OS}/${ARCH}..."

DOWNLOAD_BASE_URL="https://github.com/ctrlb-hq/ctrlb-control-plane/releases/download"
BINARY_URL="${DOWNLOAD_BASE_URL}/${COLLECTOR_VERSION}/${COLLECTOR_NAME}-${OS}-${ARCH}"
BINARY_PATH="${INSTALL_DIR}/${COLLECTOR_NAME}"

echo "🔍 Downloading binary from: $BINARY_URL"

mkdir -p "$INSTALL_DIR"
curl -L "$BINARY_URL" -o "$BINARY_PATH" || {
  echo "⚠️  Failed to download from GitHub releases. Using local build..."
  
  # Fallback: check if binary exists in parent directories
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  LOCAL_BINARY="${SCRIPT_DIR}/../agent/ctrlb-collector"
  
  if [ -f "$LOCAL_BINARY" ]; then
    cp "$LOCAL_BINARY" "$BINARY_PATH"
  else
    echo "❌ Binary not found. Please build or download manually."
    exit 1
  fi
}

chmod +x "$BINARY_PATH"

# ========================
# Create Storage Directory
# ========================
mkdir -p "${INSTALL_DIR}/data"
chmod 755 "${INSTALL_DIR}/data"

# ========================
# Configure CtrlB Collector with Fluent Bit
# ========================
echo "🧩 Creating collector configuration..."

# Parse backend URL to extract host and port
BACKEND_HOST=$(echo "$BACKEND_URL" | sed -E 's|https?://([^:/]+).*|\1|')
BACKEND_PORT=$(echo "$BACKEND_URL" | sed -E 's|.*:([0-9]+).*|\1|')
BACKEND_PROTO=$(echo "$BACKEND_URL" | sed -E 's|(https?)://.*|\1|')

# If port wasn't in URL, use default based on protocol
if [ "$BACKEND_PORT" = "$BACKEND_URL" ]; then
  if [ "$BACKEND_PROTO" = "https" ]; then
    BACKEND_PORT="443"
  else
    BACKEND_PORT="80"
  fi
fi

# Extract URI path from backend URL (default to root if none)
BACKEND_PATH=$(echo "$BACKEND_URL" | sed -E 's|https?://[^/]+(/.*)|\1|')
[ "$BACKEND_PATH" = "$BACKEND_URL" ] && BACKEND_PATH=""

cat <<EOF > "$CONFIG_FILE"
service:
  flush: 1
  log_level: info
  http_server: on
  http_listen: 0.0.0.0
  http_port: 2020
  hot_reload: on

pipeline:
  inputs:
    - name: node_exporter_metrics
      tag: ctrlb_agent_node_metrics
      scrape_interval: 2
    - name: fluentbit_metrics
      tag: ctrlb_agent_internal_metrics
      scrape_interval: 2
  outputs:
    - name: prometheus_exporter
      match: ctrlb_agent_*_metrics
      host: 0.0.0.0
      port: 2021
EOF

# ========================
# Write Environment File
# ========================
cat <<EOF > "$ENV_FILE"
BACKEND_URL=${BACKEND_URL}
PIPELINE_NAME=${PIPELINE_NAME}
STARTED_BY=${STARTED_BY}
AGENT_CONFIG_PATH=${CONFIG_FILE}
AGENT_TYPE=fluent-bit
EOF

chmod 600 "$ENV_FILE"
chmod 644 "$CONFIG_FILE"

# ========================
# Create CtrlB Collector Systemd Service
# ========================
echo "🔧 Creating collector systemd service..."

cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=${COLLECTOR_NAME} Service with Fluent Bit Backend
Documentation=https://github.com/ctrlb-hq/ctrlb-control-plane
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${BINARY_PATH}
EnvironmentFile=${ENV_FILE}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=false
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

# ========================
# Enable and Start Service
# ========================
echo "🔧 Configuring systemd service..."
systemctl daemon-reload

echo "▶️  Starting ${COLLECTOR_NAME}..."
systemctl enable "$COLLECTOR_NAME"
systemctl restart "$COLLECTOR_NAME"

# Wait a moment for service to start
sleep 3

# ========================
# Verify Installation
# ========================
echo ""
echo "🔍 Verifying installation..."

if systemctl is-active --quiet "$COLLECTOR_NAME"; then
  echo "✅ ${COLLECTOR_NAME} is running"
  echo "✅ Fluent Bit backend is embedded in the collector"
else
  echo "❌ ${COLLECTOR_NAME} failed to start"
  echo ""
  echo "Service logs:"
  systemctl status "$COLLECTOR_NAME" --no-pager -l
  echo ""
  echo "Recent logs:"
  journalctl -u "$COLLECTOR_NAME" -n 50 --no-pager
  exit 1
fi

# ========================
# Display Summary
# ========================
echo ""
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  ✅ Installation Complete!                                     ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""
echo "📊 Service Status:"
echo "   • CtrlB Collector:   systemctl status ${COLLECTOR_NAME}"
echo "   • Fluent Bit:        Managed internally by collector"
echo "   • Installation:      System package (apt/yum)"
echo ""
echo "📝 Configuration Files:"
echo "   • Config File:       ${CONFIG_FILE}"
echo "   • Environment:       ${ENV_FILE}"
echo "   • Fluent Bit Binary: ${FLUENTBIT_BIN}"
echo ""
echo "📁 Installation Directory:"
echo "   • Base Directory:    ${INSTALL_DIR}"
echo "   • Data Directory:    ${INSTALL_DIR}/data"
echo "   • Fluent Bit:        /opt/fluent-bit (system package)"
echo ""
echo "📡 Monitoring Endpoints:"
echo "   • Metrics & Health:  http://localhost:2020/api/v1/metrics"
echo ""
echo "🔧 Useful Commands:"
echo "   • View logs:         journalctl -u ${COLLECTOR_NAME} -f"
echo "   • Check status:      systemctl status ${COLLECTOR_NAME}"
echo "   • Restart service:   systemctl restart ${COLLECTOR_NAME}"
echo "   • Stop service:      systemctl stop ${COLLECTOR_NAME}"
echo "   • View config:       cat ${CONFIG_FILE}"
echo ""
echo "🗑️  To Uninstall:"
echo "   systemctl stop ${COLLECTOR_NAME} && systemctl disable ${COLLECTOR_NAME}"
echo "   rm /etc/systemd/system/${COLLECTOR_NAME}.service && systemctl daemon-reload"
echo "   rm -rf ${INSTALL_DIR}"
if [ "$DISTRO" = "ubuntu" ] || [ "$DISTRO" = "debian" ]; then
  echo "   apt-get remove --purge fluent-bit  # Remove Fluent Bit package"
else
  echo "   yum remove fluent-bit  # Remove Fluent Bit package"
fi
echo ""
echo "📚 Documentation: https://github.com/ctrlb-hq/ctrlb-control-plane"
echo ""

