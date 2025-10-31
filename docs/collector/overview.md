## 📦 Collector – Technical Docs

### 📝 Overview

A lightweight agent built on top of OpenTelemetry Collector. It registers with the backend, receives pipeline config, and streams telemetry (logs, metrics, traces).

---

### ⚙️ Installation

#### 1. Prerequisites

* Linux (x86\_64)
* Go (only if running from source)

#### 2. Binary Installation

```bash
wget https://github.com/ctrlb-hq/ctrlb-control-plane/releases/download/<version>/ctrlb-collector-linux-amd64
chmod +x ctrlb-collector-linux-amd64
sudo mv ctrlb-collector-linux-amd64 /usr/local/bin/ctrlb-collector
```

#### 3. Environment Variables

| Variable        | Required | Description                          |
| --------------- | -------- | ------------------------------------ |
| `BACKEND_URL`   | ✅        | Backend API endpoint                 |
| `PIPELINE_NAME` | ✅        | Name of the pipeline to attach to    |
| `STARTED_BY`    | ✅        | Email or identifier of the initiator |

---

### 🚀 Running the Collector

```bash
BACKEND_URL="http://localhost:8096" \
PIPELINE_NAME="test-pipeline" \
STARTED_BY="dev@example.com" \
ctrlb-collector
```

> ✅ *Only needed if you’re developing the collector. For regular use, use the install script shown in the UI.*

---

### 📡 How It Works

* Registers itself with the backend
* Receives a pipeline config from control plane (push-based)
* Applies config dynamically without restart
* Exposes `/metrics` for Prometheus scraping

---

### 📊 Metrics

* Default endpoint: `localhost:8888/metrics`
* Includes OTEL collector internal metrics and host metrics

---

### 🧼 Troubleshooting

| Issue                             | Possible Fix                             |
| --------------------------------- | ---------------------------------------- |
| `Failed to register with backend` | Check `BACKEND_URL`, pipeline name       |
| Port `:3421` already in use        | Update config or run on a different port |
| Config not applying               | Check logs, ensure backend is reachable  |
