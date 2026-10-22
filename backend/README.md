# CtrlB backend

The backend manages users, collector inventory, pipeline graphs, and collector configurations. It runs an HTTP API on port `8096` and stores metadata in SQLite.

VM collectors register through `/api/agent/v1`; the backend pushes configuration and control commands to them and scrapes their metrics. Kubernetes collectors use `/api/agent/v2` to check in, receive desired configuration, and report rollout status. The browser uses the authenticated `/api/frontend/v2` API for both workflows.

## Run locally

Build with a Go toolchain compatible with [go.mod](go.mod). The module declares Go `1.23.0` and toolchain `go1.24.4`. SQLite uses CGO, so install a C compiler and keep `CGO_ENABLED=1`.

From the repository root:

```bash
cd backend
export JWT_SECRET="$(openssl rand -hex 32)"
go run ./cmd/backend
```

The backend also loads `.env` from its working directory. `JWT_SECRET` is required. It creates `backend.db` and `app.log` in that directory, creates or migrates the database tables, and loads the embedded component schemas at startup.

See the [deployment guide](../docs/backend/deployment.md) for all environment variables, container commands, and persistent storage. The [Kubernetes setup guide](../docs/kubernetes-setup.md) covers installing managed collectors. See the [API reference](../docs/backend/api-reference.md) for routes and request examples, and the [testing guide](../docs/backend/testing.md) for checks.

## Source layout

| Path | Responsibility |
| --- | --- |
| `cmd/backend` | Startup, environment configuration, and service wiring |
| `internal/api` | HTTP routes |
| `internal/auth`, `internal/middleware` | User registration, JWTs, and request middleware |
| `internal/agent` | VM collector registration and configuration notifications |
| `internal/frontend` | Browser APIs for collectors, pipelines, and component schemas |
| `internal/kubernetes` | Installations, manifest generation, collector check-ins, and rollout state |
| `internal/db` | SQLite initialization, migrations, and schema loading |
| `internal/pkg/configcompiler` | Directed pipeline graph compilation |
| `internal/pkg/queue` | VM collector monitoring and metric scraping |
| `internal/assets` | Embedded component schemas and UI schemas |

## License

See [LICENSE](LICENSE) for the backend license and attribution terms.
