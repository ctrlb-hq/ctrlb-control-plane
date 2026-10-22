# 🚀 Backend Deployment Guide (CtrlB Control Plane)

This guide explains how to set up and run the **CtrlB Control Plane Backend** locally and in production environments.

---

## 📋 Overview

The backend is a Go-based server that handles API requests, manages agent configurations, scrapes Prometheus metrics, and serves pipeline state.

Key components:

- REST API (mux router)
- SQLite for metadata storage
- Background scraper for agent metrics

---

## ⚙️ Prerequisites

- Go >= 1.23
- SQLite3 CLI (optional for debugging)
- Git

---

## 🛠️ Environment Variables

Create a `.env` file or export the following in your shell:

```env
PORT=8096
JWT_SECRET=your-secure-secret
```

> 🔐 `JWT_SECRET` is required. Backend will panic if not provided.

---

## 🚧 Running Locally

### 1. Fork and build:

```bash
git clone https://github.com/<your-username>/ctrlb-control-plane.git
cd ctrlb-control-plane/backend
```

### 2. Set required environment variables

Create a `.env` file or export them in your shell session:

```env
PORT=8096
JWT_SECRET=your-secure-secret
```

### 3. Run the backend

```bash
go run main.go
```

The backend will start at `http://localhost:8096` by default.

---

## 🗃️ Database

The backend uses **SQLite** by default. On first run, it will create a DB file in the same directory.

No migration steps are needed at the moment.

---

## 📦 Docker Support

From the repository root, build the backend and run it with a persistent data volume. Set `JWT_SECRET` in your shell before running the container:

```bash
docker build -t ctrlb-control-plane-backend:local backend
docker volume create ctrlb-backend-data
docker run -d --name ctrlb-backend \
  -p 127.0.0.1:8096:8096 \
  --env JWT_SECRET \
  --mount type=volume,source=ctrlb-backend-data,target=/data \
  ctrlb-control-plane-backend:local
```

The container runs as UID/GID `10001:10001` and writes its database and log to `/data`. Keep that volume when replacing the container. See [Backend and collector images](../container-images.md) for GHCR publication and image tags.

---

## 📈 Logs

Logs are written to `app.log` in the current directory in JSON format.
They are also streamed to the console in human-readable format by default.

---

## 🛡️ Production Deployment Tips

- Use **Systemd** to run as a service
- Run behind **Nginx** or **Caddy** for TLS termination
- Store secrets securely by configuring environment variables. Currently, the backend reads secrets only from environment variables set via `.env` files, `systemd`, `docker run --env`, or your shell

---

## 🧰 Troubleshooting

| Problem               | Solution                                     |
| --------------------- | -------------------------------------------- |
| `JWT_SECRET not set`  | Ensure the environment variable is exported  |
| `port already in use` | Kill existing process on 8096 or change port |
| Agent not syncing     | Check backend logs for registration errors   |

---

For more details on API routes, refer to the [API Reference](./api-reference.md).
