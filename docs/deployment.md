# 🚀 Deployment Guide for CtrlB Control Plane

This guide walks you through deploying the CtrlB Control Plane components in a production or staging environment.

---

## 🧱 Deployment Modes

- **All-in-One (Dev/Staging)**: Run backend and frontend on a single VM.
- **Split (Prod)**: Run backend, frontend, and collector independently with service isolation.

---

## 📦 Backend Deployment

### Option A: Quick Install (Recommended)

The easiest way to install the backend is to use our automated installation script. This will download the latest release binary and configure it as a systemd service.

**Basic Installation:**

```bash
curl -fsSL https://raw.githubusercontent.com/ctrlb-hq/ctrlb-control-plane/main/scripts/backend-install.sh | sudo bash
```

> **Note:** If you don't provide a JWT secret, the script will prompt you to enter one interactively.

**Installation with JWT Secret:**

```bash
curl -fsSL https://raw.githubusercontent.com/ctrlb-hq/ctrlb-control-plane/main/scripts/backend-install.sh | sudo bash -s -- --jwt-secret "your-secret-key-here"
```

**Advanced Options:**

```bash
curl -fsSL https://raw.githubusercontent.com/ctrlb-hq/ctrlb-control-plane/main/scripts/backend-install.sh | sudo bash -s -- \
  --jwt-secret "your-secret-key-here" \
  --port 8096 \
  --env prod \
  --workers 4 \
  --check-interval 10
```

Available options:
- `--jwt-secret <secret>`: JWT secret key (required, will prompt if not provided)
- `--port <port>`: Port to run the backend on (default: 8096)
- `--env <env>`: Environment (default: prod)
- `--workers <count>`: Number of workers (default: 4)
- `--check-interval <mins>`: Check interval in minutes (default: 10)

The script will:
- Download the appropriate binary for your OS and architecture
- Install it to `/opt/ctrlb/control-plane-backend/`
- Create a systemd service file
- Configure environment variables
- Enable and start the service automatically

### Option B: Manual Installation

#### 1. Build the Binary

```bash
cd backend
go build -o control-plane-backend cmd/backend/main.go
```

#### 2. Install Binary and Create Service

```bash
sudo mkdir -p /etc/control-plane-backend
sudo cp control-plane-backend /usr/local/bin/
sudo cp scripts/systemd/control-plane-backend.service /etc/systemd/system/
```

#### 3. Create Environment File

```bash
sudo tee /etc/control-plane-backend/env > /dev/null <<EOF
JWT_SECRET=<SECRET_HERE>
# Other environment variables as needed
EOF
```

#### 4. Enable & Start Service

```bash
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable control-plane-backend.service
sudo systemctl start control-plane-backend.service
```

---

## 🌐 Frontend Deployment

### 1. Set Up Environment Variables

Create a `.env` file in the `frontend/` directory:

```bash
cd frontend
tee .env > /dev/null <<EOF
VITE_API_URL=http://localhost:8080
EOF
```

This sets the backend URL for API requests during local development.

### 2. Build Frontend Assets

```bash
cd frontend
npm install
npm run build
```

### 2. Serve with Any Static File Server

```bash
npm install -g serve
serve -s dist -l 3000
```

Or deploy via Nginx, Netlify, Vercel, etc.

---

## 🛰️ Collector Installation

Collector installation steps are provided via the UI. Once a new collector is created, a corresponding installation command with a unique token and configuration is displayed.
The control plane will wait for the collector to complete the setup and come online.
Each collector:

- Fetches its configuration from the control plane
- Exposes health metrics on `:8888`

---

## 🔒 Security Notes

- Token-based authentication for collector is under development.
- Use HTTPS reverse proxies (e.g., Nginx, Caddy) in production.

---

For more details, refer to:

- [Architecture Overview](architecture.md)
- [API Reference](backend/api-reference.md)
- [Troubleshooting Guide](troubleshooting.md)
