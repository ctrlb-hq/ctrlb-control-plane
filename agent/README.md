# 🛰️ CTRLB Agent

The **CTRLB Agent** is a flexible telemetry collection agent that supports multiple backends including **OpenTelemetry Collector** and **Fluent Bit**. Once installed, it connects to the CTRLB backend, shares its runtime status, and receives its initial configuration. The agent can be managed remotely through a set of defined HTTP endpoints.

---

## 🔧 Responsibilities

- Register with the backend upon startup.
- Expose runtime metrics in Prometheus format, including CPU utilization, memory utilization, and data transfer rates (sent/received) for logs, traces, and metrics.
- Receive initial and updated configurations.
- Respond to remote lifecycle commands.

---

## 🚀 Installation

The backend provides a platform-specific installation script. Once executed, the agent starts and reaches out to the backend for initial configuration.

You can customize agent behavior using environment variables:

- `AGENT_TYPE`: Type of telemetry backend to use (`otel`, `fluent-bit`, `fluentbit`, `fb`). Default is `otel`
- `AGENT_CONFIG_PATH`: Path to the agent configuration file. Default is `./config.yaml`
- `BACKEND_URL`: URL of the backend server. Default is `http://pipeline.ctrlb.ai:8096`
- `PIPELINE_NAME`: Name of the pipeline (optional)
- `STARTED_BY`: User who started the agent (optional)

### Supported Backends

#### OpenTelemetry Collector (Default)
The OpenTelemetry Collector runs as an embedded library within the agent process.

```bash
export AGENT_TYPE=otel
export AGENT_CONFIG_PATH=./otel-config.yaml
./ctrlb_collector
```

#### Fluent Bit
Fluent Bit runs as a child process controlled by the agent via HTTP APIs.

```bash
export AGENT_TYPE=fluent-bit
export AGENT_CONFIG_PATH=./fluentbit-config.yaml
./ctrlb_collector
```

See [Fluent Bit Integration Guide](../docs/collector/fluentbit-integration.md) for detailed setup instructions.

---

## 🌐 Agent API Endpoints

All endpoints are served under the base path: `/agent/v1`

### Lifecycle Actions

These endpoints control the agent's operational state:

- `POST /agent/v1/start` – Start the telemetry collector (OTEL/Fluent Bit) without restarting the agent process.
- `POST /agent/v1/stop` – Stop the telemetry collector while keeping the agent process alive.
- `POST /agent/v1/shutdown` – Gracefully shut down the agent and collector.

### Configuration

These endpoints manage the agent’s configuration:

- `POST /agent/v1/config` – Push updated configuration to the agent.

> ℹ️ On initial startup, the agent fetches its configuration automatically from the backend.

---

## 🛠️ Tech Stack

- **Go** – Core implementation language.
- **OpenTelemetry Collector** – Embedded telemetry collector (default).
- **Fluent Bit** – Lightweight log processor (alternative backend).
- **HTTP + Gorilla Mux** – Communication protocol and routing.

## 📚 Documentation

- [Fluent Bit Integration Guide](../docs/collector/fluentbit-integration.md)
- [Fluent Bit Quick Start](../docs/collector/fluentbit-quickstart.md)
- [Architecture Overview](../docs/architecture.md)

---

## 📄 License

AGPL License. See [LICENSE](../LICENSE) for more details.

