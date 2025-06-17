# 🏗️ CtrlB Control Plane Architecture

This document explains the high-level design and internal components of **CtrlB Control Plane**, the telemetry agent control plane built by **CTRLB**.

---

## 📚 Overview

CtrlB Control Plane is designed to manage distributed telemetry collector (OpenTelemetry Collector instances) across a variety of environments. It includes:

- A **Control Plane** (frontend + backend)
- A **Lightweight Agent** installed on target environments
- **Configuration & Communication Layers** that bind them together

---

## 🧹 Core Components

### 1. Frontend

- Built with **React** + **Vite**
- Provides a **graph-based visual interface** to design and manage OpenTelemetry collector pipelines
- Allows users to:
  - **Edit configuration** and **Delete pipelines**
  - View live **agent metrics**, including:
    - CPU and memory usage
    - Pipeline throughput for logs, metrics, and traces
- Offers configuration previews (YAML) and error validation
- Communicates with the backend via RESTful APIs
- Requires `.env` with `VITE_API_URL` for dev setup

### 2. Backend

- Implemented in **Go**
- Handles:
  - Agent and pipeline **registration**, **lifecycle management**, and **status tracking**
  - **Graph-to-OTel config compilation**
  - A background worker that:
    - Scrapes metrics from `/metrics` endpoint of collector (Prometheus exposition)
    - Updates the internal **SQLite** DB with health and usage info
- Supports embedding config files directly in releases
- API reference is [available here](../backend/api-reference.md)

### 3. Agent

- Lightweight wrapper around the **OpenTelemetry Collector**
- Key roles:
  - **Receives configuration updates pushed by the backend** (push model)
  - **Exposes /metrics** endpoint for Prometheus-style scraping by backend
  - **Auto-reloads** configuration without restarts upon update detection
  - Packaged as a binary with **systemd service** support for Linux installation
- Install script is generated from the UI and bootstraps:
  - Binary download
  - Environment setup
  - Systemd service creation

---

## 🔄 Communication Flow

1. **Collector Bootstraps**:

   - Collector starts and contacts the backend via `/api/agent/v1/agents`, sending metadata like platform, hostname, version, etc.
   - Backend registers the collector and responds with the latest configuration.

2. **Heartbeat & Metrics**:

   - Backend scrapes the Prometheus metrics exposed by the agent on a periodic basis.
   - Health information and operational metrics are stored in database.

3. **Config Change Notification**:

   - Agent watches its config file and notifies the backend if it detects a change.
   - Backend returns the updated configuration, which the agent writes to its local config file before reloading the collector.
   - Collector reloads itself with the new config.

4. **User-Initiated Configuration Updates**:
   - Users manage configurations and monitor collector through the frontend.
   - When a user updates a config, the backend persists the change and queues it for delivery to relevant collector.

---

## 🔐 Security (WIP)

- Token-based auth for collector registration
- User-level RBAC on the frontend/backend

---

## 📈 Scalability Notes

CtrlB Control Plane was initially designed for small-scale systems with around 40–50 agents and pipelines, with simplicity as the guiding principle.

Planned improvements for scalability:

- Stateless backend (backed by external DB)
- Agent polling is lightweight and time-interval based
- Config push model in roadmap
- Scaling architecture to support 1,000+ agents/pipelines in upcoming iterations

---

## 🧱 Diagrams

![CtrlB Control Plane ER Diagram](./assets/CtrlB Control Plane-er-diagram.png)
_Entity-Relationship Diagram of CtrlB Control Plane's internal database schema_

---

## 🚓️ Roadmap (Architecture)

- Support for multiple tenants/workspaces
- Cloud-native deployment (Helm, Terraform modules)
- Plugin system for other agent types
- Agent auto-upgrade and signature validation
