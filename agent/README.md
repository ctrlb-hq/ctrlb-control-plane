# 🛰️ CTRLB Agent

The **CTRLB Agent** is a lightweight wrapper built on top of the OpenTelemetry Collector. Once installed, it connects to the CTRLB backend, shares its runtime status, and receives its initial configuration. The agent can be managed remotely through a set of defined HTTP endpoints.

---

## 🔧 Responsibilities

- Register with the backend upon startup.
- Expose runtime metrics in Prometheus format, including CPU utilization, memory utilization, and data transfer rates (sent/received) for logs, traces, and metrics.
- Receive initial and updated configurations.
- Respond to remote lifecycle commands.

---

## 🚀 Installation

The backend provides a platform-specific installation script. Once executed, the agent starts and reaches out to the backend for initial configuration.

You can customize agent behavior using the following command-line flags:

- `--config`: Path to the agent configuration file. Default is `./config.yaml`
- `--backend`: URL of the backend server. Default is `http://pipeline.ctrlb.ai:8096`
- `--port`: Port on which the agent listens for communication. Default is `443`. This is currently an experimental feature.

---

## 🌐 Agent API Endpoints

All endpoints are served under the base path: `/agent/v1`

### Lifecycle Actions

These endpoints control the agent's operational state:

- `POST /agent/v1/start` – Start the OpenTelemetry Collector instance without restarting the agent process.
- `POST /agent/v1/stop` – Stop the OpenTelemetry Collector instance while keeping the agent process alive.
- `POST /agent/v1/shutdown` – Gracefully shut down the agent.

### Configuration

These endpoints manage the agent’s configuration:

- `POST /agent/v1/config` – Push updated configuration to the agent.

> ℹ️ On initial startup, the agent fetches its configuration automatically from the backend.

---

## 🛠️ Tech Stack

- **Go** – Core implementation language.
- **OpenTelemetry Collector** – Base collector engine.
- **HTTP + Gorilla Mux** – Communication protocol and routing.

---

## 📄 License

AGPL License. See [LICENSE](../LICENSE) for more details.

