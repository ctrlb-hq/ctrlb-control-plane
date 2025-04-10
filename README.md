# 🛠️ CTRLB – Control Plane for Managing Agents

**CTRLB** is a control plane designed to manage telemetry agents across diverse environments. It currently supports the configuration and lifecycle management of **OpenTelemetry Collector** agents.

This monorepo is organized into three primary components:

- **Frontend**: A modern React-based UI to visualize and manage agents and their configurations.
- **Backend**: A Go-powered API and orchestration engine that communicates with agents, handles configuration state, and offers installation scripts that users can run manually to set up agents.
- **Agent**: A lightweight wrapper around the OpenTelemetry Collector, capable of dynamically receiving configurations and reporting status back to the control plane.

---

## ✨ Features

- Centralized management of distributed telemetry agents.
- Declarative, graph-based configuration interface for OpenTelemetry Collector (OTEL) only.
- Real-time agent status monitoring.

---

## 📆 Repository Structure

```
ctrlb/
├── frontend/        # React UI
├── backend/         # Go-based API server and logic
├── agent/           # OTel Collector wrapper with remote config support
└── docs/            # Architecture diagrams, usage guides, etc.
```

---

## 🚀 Getting Started

> 📘️ Prerequisites:
>
> - Go 1.23+
>
> - Node.js 18+
>
> - Docker (for local development)

Clone the repository:

```bash
git clone https://github.com/your-org/ctrlb.git
cd ctrlb
```

Start the dev environment (local setup guide coming soon).

---

## 🏗️ Architecture

CTRLB is built to support agent orchestration at scale. Key components include:

- **Agent Communication Layer**: Simple HTTP communication to and from registered agents. Token-based authentication is a work in progress.
- **Configuration Manager**: Handles creation, versioning, and delivery of agent configs.
- **Storage Backend**: Currently supports SQLite for storing configuration and telemetry metadata. Support for other SQL-based databases is coming soon.

*More details are available in the* **[docs/architecture.md](docs/architecture.md)**

---

## 📖 Documentation

- [Agent Configuration](docs/agent-config.md)
- [API Reference](docs/api.md)
- [Deployment Guide](docs/deployment.md)
- [Troubleshooting](docs/troubleshooting.md)

---

## 🤝 Contributing

Please refer to the [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to get involved.

---

## 📜 License

AGPL License. See [LICENSE](LICENSE) for more details.

