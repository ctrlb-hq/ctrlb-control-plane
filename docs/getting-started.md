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

## 🖥️ Local Development Setup

### 1. Start the Backend

```bash
cd backend
export JWT_SECRET="your-secret-key"
go run cmd/backend/main.go
```

This will:

- Run the Go backend locally.
- Use SQLite by default

### 2. Start the Frontend

Open a new terminal window:

```bash
cd frontend
npm install --legacy-peer-deps
cp .env.example .env
npm run dev
```

The frontend should now be running at [http://localhost:3030](http://localhost:3030)

### 3. Run the Collector

> ⚠️ **Use this method only if you are actively developing or modifying the collector code.**  
> For regular usage, install the collector using the instructions provided in the UI.

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
