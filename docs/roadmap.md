# 🛣️ CTRLB Control Plane – Public Product Roadmap

> This roadmap outlines planned milestones for CTRLB's OpenTelemetry control plane. Timelines are approximate and subject to change. Community feedback is welcome — feel free to [open issues](https://github.com/ctrlb-hq/ctrlb-control-plane/issues) or contribute via PRs.

---

## 🔹 Phase 1: Developer Preview (Now – \~1 month)

> ⚙️ **Goal:** Build trust among OSS users and make the control plane usable for small teams

### ✅ Features

- Agent install script (with dynamic config pull)
- Config UI with schema-based generation
- Graph-based pipeline builder
- Agent status & metrics dashboard
- Push-based config update via queue
- SQLite metadata store
- Release binaries for backend + agent

### 📌 Milestones

- OSS-ready README, LICENSE, CONTRIBUTING
- GitHub Releases + install docs
- Try-it-now hosted demo or video walkthrough

---

## 🔸 Phase 2: OSS Adoption & Hardening (1–2 months)

> ⚙️ **Goal:** Improve stability, allow community usage at scale

### 🛠 Features

- Backend auth (JWT + API keys)
- Agent auth handshake with secret token
- Config validation + dry run before apply
- RBAC for view/edit pipeline access control
- Remote pipeline templates
- Collector upgrade feature from UI
- Postgres metadata store support
- Improved frontend UI/UX polish

### 🔧 Infrastructure

- Tests: backend services, agent state sync
- GitHub Actions: build/test/release pipeline
- Cross-platform binaries for backend/agent
- Docker images for backend + self-hosted demo

---

## 🟦 Phase 3: Enterprise Features (2–4 months)

> ⚙️ **Goal:** Offer CTRLB as a control plane for teams/orgs with scale and compliance needs

### 🧩 Features

- RBAC and org/user/project separation
- Versioned config history & rollback
- Alerting integration (Slack/Webhook/email)
- Hosted cloud offering (alpha)
- Audit logs and activity tracking

### 🏗 Infrastructure

- Separate operational metadata and time-series telemetry storage:
- Store metadata (agents, pipelines, configs, users) in PostgreSQL
- Store metrics and telemetry data in a time-series DB

---

## 🟨 Phase 4: Enterprise Readiness (5+ months)

> ⚙️ **Goal:** Ensure the platform is secure, scalable, and production-grade for organizations of all sizes while remaining 100% open source

### 🧱 Deployment & Operations

- Air-gapped deployment support
- SSO/SAML authentication
- Advanced logging, backup, and HA setup

### 🔐 Security & Compliance

- SOC2 readiness practices
- Encryption at rest and in transit
- Multi-region backend for high availability

### 🧱 Self-hosted Enterprise

- Air-gapped deployment support
- SSO/SAML auth
- Advanced logging, backup, HA setup

### ☁️ Hosted SaaS

- Usage-based billing (agents, pipelines)
- Team management, usage limits
- Credit-based auto-scaling agents

### 🔐 Security & Compliance

- SOC2 readiness
- Encryption at rest/in transit
- Multi-region backend for HA
