# Backend testing

Run backend checks from the `backend/` module. Use the Go toolchain declared in [go.mod](../../backend/go.mod), with a C compiler and CGO enabled for SQLite. Some tests create local HTTP servers, so the environment must allow loopback listeners.

## Standard checks

From the repository root:

```bash
cd backend
go test ./...
go vet ./...
```

Tests use mocks, temporary databases, and local HTTP servers. They do not require a running backend, a Kubernetes cluster, or an exported `JWT_SECRET`.

## Focused checks

Run a package or one named test from `backend/`:

```bash
go test ./internal/pkg/configcompiler
go test ./internal/kubernetes -run '^TestCheckInPreservesIdentityAndRollsOutIndependently$' -v
```

Useful packages include:

| Package | Checks |
| --- | --- |
| `./internal/api`, `./internal/middleware` | Routes, authentication, and CORS |
| `./internal/auth`, `./internal/agent` | User authentication and VM registration |
| `./internal/frontend/...` | Browser handlers, services, repositories, and pipeline deletion |
| `./internal/db` | SQLite initialization, migration of existing rows, and bundled schema refresh |
| `./internal/pkg/configcompiler` | Directed graph ordering, signal compatibility, branching rules, and storage extensions |
| `./internal/pkg/queue` | VM monitoring, metric parsing, and repository writes |
| `./internal/kubernetes` | Installation tokens, check-ins, manifests, rollout state, and the Kubernetes preset |

For race detection and a coverage report:

```bash
go test -race -coverprofile=/tmp/ctrlb-backend-coverage.out ./...
go tool cover -html=/tmp/ctrlb-backend-coverage.out
```

## Collector integration

Backend tests validate the preset graph and management API. To check the compiled preset against the collector's actual component factories, follow the cross-module command in [Kubernetes collector management](../kubernetes.md#validate-the-collection-preset). Use the [Kubernetes setup guide](../kubernetes-setup.md) to deploy collectors to a cluster and verify enrollment, telemetry delivery, and configuration rollouts.
