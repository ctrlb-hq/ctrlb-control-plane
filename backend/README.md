# 🧠 CTRLB Backend

The **CTRLB Backend** is the core API and orchestration layer responsible for managing agents, pipelines, and configurations. It exposes a RESTful API and enforces authentication for protected routes.

---

## 🔧 Requirements

The backend requires the following environment variable:

- `JWT_SECRET`: A secret key used for signing and verifying JWT tokens.

---

## 🚀 API Endpoints

### Auth API (v1)

**Base Path:** `/api/auth/v1`

- `POST /register` – Register a new user
- `POST /login` – Authenticate user and issue tokens
- `POST /refresh` – Refresh access tokens

### Agent API (v1)

**Base Path:** `/api/agent/v1`

- `POST /agents` – Register a new agent
- `POST /agents/{id}/config-changed` – Notify backend of agent config changes

### Frontend API (v2, Auth Protected)

**Base Path:** `/api/frontend/v2`

#### Agents

- `GET /agents` – Get all registered agents
- `GET /agents/{id}` – Get agent details
- `DELETE /agents/{id}` – Delete an agent
- `POST /agents/{id}/start` – Start agent
- `POST /agents/{id}/stop` – Stop agent
- `POST /agents/{id}/restart-monitoring` – Restart monitoring logic for an agent
- `GET /agents/{id}/healthmetrics` – Get agent health metrics
- `GET /agents/{id}/ratemetrics` – Get data rate metrics (logs, traces, metrics)
- `POST /agents/{id}/labels` – Add/update agent labels
- `GET /unassigned-agents` – List agents not yet linked to pipelines

#### Pipelines

- `GET /pipelines` – Get all pipelines
- `POST /pipelines` – Create a new pipeline
- `GET /pipelines/{id}` – Get pipeline details
- `DELETE /pipelines/{id}` – Delete pipeline
- `GET /pipelines/{id}/graph` – Get pipeline configuration graph
- `POST /pipelines/{id}/graph` – Sync pipeline configuration graph
- `GET /pipelines/{id}/agents` – List agents attached to a pipeline
- `POST /pipelines/{id}/agent/{agent_id}` – Attach agent to pipeline
- `DELETE /pipelines/{id}/agent/{agent_id}` – Detach agent from pipeline

#### Components

- `GET /component` – Get component info
- `GET /component/schema/{name}` – Get schema for a specific component

---

## 🛠️ Tech Stack

- **Go** – Core implementation language
- **JWT** – Authentication system
- **Gorilla Mux** – HTTP router
- **SQLite** – Supported database

---

## 📄 License

AGPL License. See [LICENSE](../LICENSE) for more details.

