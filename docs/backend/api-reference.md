# 📘 API Reference (CTRLTower Backend)

This document provides an overview of the available API endpoints in the CTRLTower backend.

---

## 🔐 Auth APIs (`/api/auth/v1`)

| Method | Endpoint    | Description          |
| ------ | ----------- | -------------------- |
| POST   | `/register` | Register a new user  |
| POST   | `/login`    | Login to get tokens  |
| POST   | `/refresh`  | Refresh access token |

---

## ⚙️ Agent APIs (`/api/agent/v1`)

| Method | Endpoint                      | Description                                |
| ------ | ----------------------------- | ------------------------------------------ |
| POST   | `/agents`                     | Register a new agent                       |
| POST   | `/agents/{id}/config-changed` | Agent notifies that its config has changed |

---

## 🌐 Frontend APIs (`/api/frontend/v2`)

> **Note**: All endpoints below require authentication via `AuthMiddleware`, which checks for a valid bearer token in the `Authorization` header.

### 🔍 Agent Management

| Method | Endpoint                          | Description                                                                |
| ------ | --------------------------------- | -------------------------------------------------------------------------- |
| GET    | `/agents`                         | Get all registered agents                                                  |
| GET    | `/agents/{id}`                    | Get detailed info of a single agent                                        |
| DELETE | `/agents/{id}`                    | Delete an agent                                                            |
| POST   | `/agents/{id}/start`              | Start an agent                                                             |
| POST   | `/agents/{id}/stop`               | Stop an agent                                                              |
| POST   | `/agents/{id}/restart-monitoring` | Restart agent's monitoring                                                 |
| GET    | `/agents/{id}/healthmetrics`      | Get health metrics for a specific agent                                    |
| GET    | `/agents/{id}/ratemetrics`        | Get rate metrics for a specific agent                                      |
| POST   | `/agents/{id}/labels`             | Add or update labels for a specific agent                                  |
| GET    | `/unassigned-agents`              | Retrieve a list of agents that are active but not yet assigned to any pipeline |

### 🔁 Pipeline Management

| Method | Endpoint                            | Description                              |
| ------ | ----------------------------------- | ---------------------------------------- |
| GET    | `/pipelines`                        | List all pipelines                       |
| POST   | `/pipelines`                        | Create a new pipeline                    |
| GET    | `/pipelines/{id}`                   | Get details of a pipeline                |
| DELETE | `/pipelines/{id}`                   | Delete a pipeline                        |
| GET    | `/pipelines/{id}/graph`             | Fetch pipeline graph                     |
| POST   | `/pipelines/{id}/graph`             | Sync pipeline graph                      |
| GET    | `/pipelines/{id}/agents`            | List all agents attached to the pipeline |
| DELETE | `/pipelines/{id}/agents/{agent_id}` | Detach an agent from the pipeline        |
| POST   | `/pipelines/{id}/agents/{agent_id}` | Attach an agent to the pipeline          |

### 🧩 Component Management

| Method | Endpoint                   | Description                                                         |
| ------ | -------------------------- | ------------------------------------------------------------------- |
| GET    | `/component`               | Get all components (optional query param: `type` to filter results) |
| GET    | `/component/schema/{name}` | Get schema for a specific component                                 |

