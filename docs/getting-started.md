# 🚀 Getting Started with CtrlB Control Plane

This guide helps you set up and run **CtrlB Control Plane** in your local development environment.

---

## ⚙️ Prerequisites

Before you begin, make sure you have the following installed:

- **Go** 1.23 or later
- **Node.js** 18 or later
- **Docker** (for running agent containers and dev services)

---

### Fork the repo and clone it

```bash
git clone https://github.com/your-username/ctrlb-control-plane.git
```

--- 

## 🧱 Component Overview

Here's a quick overview of the major components you'll interact with:

- `frontend/`: React-based web UI
- `backend/`: Go-based API server
- `agent/`: Lightweight wrapper around OpenTelemetry Collector
- `scripts/`: Utilities and installer scripts

---

## 🖥️ Quick Start (Production/Staging)

### 1. Install the Backend (Quick)

The fastest way to get started is to use our automated installation script. Because the script runs non-interactively when piped into `bash`, you must provide the JWT secret flag in the command:

```bash
curl -fsSL https://raw.githubusercontent.com/ctrlb-hq/ctrlb-control-plane/main/scripts/backend-install.sh | sudo bash -s -- --jwt-secret "your-secret-key-here"
```

The backend will be running at `http://localhost:8096`

### 2. Install the Frontend

```bash
cd frontend
npm install --legacy-peer-deps
cp .env.example .env
npm run build
npm install -g serve
serve -s dist -l 3030
```

The frontend will be running at [http://localhost:3030](http://localhost:3030)

### 3. Install a Collector

Once the backend is running, you can install collectors on your target machines. Use the automated installation script:

```bash
curl -fsSL https://raw.githubusercontent.com/ctrlb-hq/ctrlb-control-plane/main/scripts/agent-install.sh | sudo bash
```

> **Note:** You'll be prompted to enter the backend URL, pipeline name, and your email during installation.

Or provide all parameters directly:

```bash
curl -fsSL https://raw.githubusercontent.com/ctrlb-hq/ctrlb-control-plane/main/scripts/agent-install.sh | \
  sudo BACKEND_URL="http://your-backend:8096" \
       PIPELINE_NAME="production-pipeline" \
       STARTED_BY="user@example.com" \
  bash
```

> 💡 **Tip:** For easier collector management, use the UI to create collectors and get customized installation commands with all parameters pre-filled.

---

## 🛠️ Local Development Setup

For active development on the codebase, use these steps:

### 1. Start the Backend (Development)

```bash
cd backend
export JWT_SECRET="your-secret-key"
go run cmd/backend/main.go
```

This will:

- Run the Go backend locally.
- Use SQLite by default

### 2. Start the Frontend (Development)

Open a new terminal window:

```bash
cd frontend
npm install --legacy-peer-deps
cp .env.example .env
npm run dev
```

The frontend should now be running at [http://localhost:3030](http://localhost:3030)

### 3. Run the Collector (Development)

> ⚠️ **Use this method only if you are actively developing or modifying the collector code.**  
> For regular usage, install the collector using the installation script above or the instructions provided in the UI.

```bash
cd agent

# Set required environment variables
export BACKEND_URL="http://localhost:8096"
export PIPELINE_NAME="test-pipeline"
export STARTED_BY="dev-user@example.com"

# Run the agent
go run cmd/ctrlb_collector/main.go
```

---

## 🏦 Directory Structure Reference

```
CtrlB Control Plane/
├── frontend/        # React UI
├── backend/         # Go API server
├── agent/           # Telemetry agent wrapper
├── scripts/         # Installer scripts
└── docs/            # Documentation and architecture
```

---

## 🔧 Next Steps

- Explore the [Architecture](./architecture.md)
- Read the [API Reference](./api.md)

---

Need help? Create an issue or join our community at [ctrlb.dev](https://docs.ctrlb.ai/)
