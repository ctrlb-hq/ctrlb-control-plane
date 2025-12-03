# 🛠️ CtrlB Control Plane – Manage Telemetry Collectors at Scale

**CtrlB Control Plane** is an orchestration system developed by **CTRLB** to manage telemetry collectors across diverse environments. It currently supports lifecycle and configuration management for the **OpenTelemetry Collector**.

This monorepo is organized into three primary components:

- **Frontend**: A modern React-based UI to visualize and manage collectors and their configurations.
- **Backend**: A Go-powered API server that communicates with agents, manages configuration state, and provides installation scripts for manual setup.
- **Agent**: A lightweight wrapper around the OpenTelemetry Collector with dynamic config fetching and status reporting.

---

## ✨ Features

- 🧹 Centralized management of distributed OpenTelemetry Collectors
- ⚙️ Declarative, graph-based configuration interface
- ⌚ Real-time health and status monitoring of connected collectors

---

## 🗖️ Repository Structure

```text
ctrlb-control-plane/
├── frontend/        # React + Vite-based UI for configuration and control
├── backend/         # Go-based API server for orchestration, config delivery, and agent registration
├── agent/           # Wrapper on OpenTelemetry Collector with dynamic config support
├── scripts/         # Shell scripts to install and manage backend or collectors
└── docs/            # Architecture diagrams, development guides, usage examples
```

---

## 🚀 Getting Started

> 📘️ Prerequisites:
>
> - Go 1.23+
> - Node.js 18+
> - Docker (for local development)

Clone the repository:

```bash
git clone https://github.com/ctrlb-hq/ctrlb-control-plane.git
cd ctrlb-control-plane
```

Start the development environment:

```bash
# Backend
cd backend && go run cmd/backend/main.go

# Frontend
cd ../frontend && npm install && npm run dev
```

> 📘 Local development guide coming soon in `docs/development.md`

---

## 💧 Architecture

**CtrlB Control Plane** is built for scalable collector orchestration. Key architectural components include:

- **Collector Communication Layer**: Simple HTTP interface for agents to register, fetch config, and report health
- **Configuration Manager**: Tracks, versions, and delivers pipeline configs
- **Storage Backend**: Uses SQLite for metadata and state storage (PostgreSQL support coming soon)
- **Authentication**: Token-based auth for agent registration and communication (in progress)

More details in [docs/architecture.md](docs/architecture.md)

---

## 📖 Documentation

- [Collector Configuration](docs/collector/configuration.md)
- [API Reference](docs/backend/api-reference.md)
- [Deployment Guide](docs/deployment.md)
- [Troubleshooting](docs/troubleshooting.md)

---

## 🤝 Contributing

We welcome contributions! Please read [CONTRIBUTION.md](CONTRIBUTION.md) for setup instructions, coding guidelines, and more.

---

## 📜 License

This project is licensed under the **GNU Affero General Public License v3.0**. See [LICENSE](LICENSE) for details.
