# 🚀 Getting Started with CtrlB Control Plane

This guide helps you set up and run **CtrlB Control Plane** in your local development environment.

---

## ⚙️ Prerequisites

Before you begin, make sure you have the following installed:

- **Go** 1.23 or later
- **Node.js** 18 or later
- **Docker** (for running agent containers and dev services)

---

## 📦 Clone the Repository

```bash
git clone https://github.com/ctrlb-hq/CtrlB Control Plane.git
cd CtrlB Control Plane
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
npm run dev
```

The frontend should now be running at [http://localhost:5173](http://localhost:5173)

### 3. Run the Agent

To run an agent locally with dynamic config support:

```bash
cd agent
go run cmd/ctrlb_collector/main.go -backend=localhost:8096 -config=./internal/config/otel.yaml
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
- Check out [Agent Configuration](./agent-config.md)
- Read the [API Reference](./api.md)

---

Need help? Create an issue or join our community at [ctrlb.dev](https://docs.ctrlb.ai/)
