# 🚀 Backend Deployment Guide (CTRLTower)

This guide explains how to set up and run the **CTRLTower Backend** locally and in production environments.

---

## 📋 Overview
The backend is a Go-based server that handles API requests, manages agent configurations, scrapes Prometheus metrics, and serves pipeline state.

Key components:
- REST API (mux router)
- SQLite for metadata storage
- Background scraper for agent metrics

---

## ⚙️ Prerequisites
- Go >= 1.20
- SQLite3 CLI (optional for debugging)
- Git
- Make (optional)

---

## 🛠️ Environment Variables
Create a `.env` file or export the following in your shell:

```env
BACKEND_PORT=8096
JWT_SECRET=your-secure-secret
```

> 🔐 `JWT_SECRET` is required. Backend will panic if not provided.

---

## 🚧 Running Locally

### 1. Clone and build:
```bash
git clone https://github.com/ctrlb-hq/ctrlb-control-plane.git
cd ctrlb-control-plane/backend
go run main.go
```

The backend will run at `http://localhost:8096` by default.

---

## 🗃️ Database
The backend uses **SQLite** by default. On first run, it will create a DB file in the same directory.

No migration steps are needed at the moment.

---

## 📦 Docker Support
*(Optional)*

Use the existing `Dockerfile` to build and run the backend container.

Then run:
```bash
docker build -t ctrlbtower-backend .
docker run -p 8096:8096 --env JWT_SECRET=xxx ctrlbtower-backend
```

---

## 📈 Logs
Logs are written to `app.log` in the current directory in JSON format.
They are also streamed to the console in human-readable format by default.

---

## 📈 Logs
Logs are printed to stdout. You can pipe to a file using:
```bash
go run main.go > backend.log 2>&1
```

---

## 🛡️ Production Deployment Tips
- Use **Systemd** to run as a service
- Run behind **Nginx** or **Caddy** for TLS termination
- Store secrets securely by configuring environment variables. Currently, the backend reads secrets only from environment variables set via `.env` files, `systemd`, `docker run --env`, or your shell

---

## 🧰 Troubleshooting
| Problem | Solution |
|--------|----------|
| `JWT_SECRET not set` | Ensure the environment variable is exported |
| `port already in use` | Kill existing process on 8096 or change port |
| Agent not syncing | Check backend logs for registration errors |

---

For more details on API routes, refer to the [API Reference](./api-reference.md).

