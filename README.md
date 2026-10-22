# CtrlB Control Plane

CtrlB Control Plane manages OpenTelemetry Collectors on VMs and in Kubernetes. Use the React UI to build telemetry pipelines, deploy collectors, and inspect their health and configuration status.

## Current capabilities

- Build receiver, processor, and exporter graphs with schema-based forms and configuration previews.
- Manage VM collectors through registration, configuration pushes, lifecycle commands, and metrics scraping.
- Deploy one Kubernetes collector per eligible node with a generated DaemonSet manifest. Collect pod logs and node, pod, and container metrics through a shared pipeline.
- Track Kubernetes connection status, engine health, and configuration rollout separately. Collectors retain their last working configuration during control-plane outages and attempt to restore it after failed updates.
- Authenticate UI/API users with access and refresh tokens, and Kubernetes installations with installation tokens.
- Store pipeline graphs, users, installations, collector status, and metrics in SQLite.

VM and Kubernetes collectors use different communication models. VM management requires the backend to reach collector ports `3421` and `8888`. Kubernetes collectors initiate outbound check-ins and need no inbound management connection from the backend.

## Run locally

Install Go 1.26 or later for both Go modules, a C compiler for the backend's SQLite driver, and Node.js 24 with npm. Kubernetes installation also requires kubectl and access to a Linux cluster. Docker is optional for container builds.

```bash
git clone https://github.com/ctrlb-hq/ctrlb-control-plane.git
cd ctrlb-control-plane
```

Start the backend:

```bash
cd backend
export JWT_SECRET="$(openssl rand -hex 32)"
go run ./cmd/backend
```

In a second terminal, from the repository root:

```bash
cd frontend
npm ci --legacy-peer-deps
npm run dev
```

Open [http://localhost:3030](http://localhost:3030), create an account, and sign in. The frontend uses `http://localhost:8096/api` during development without an environment file. The backend creates `backend.db` and `app.log` in its working directory. Keep `JWT_SECRET` stable across restarts to preserve existing sessions.

Continue with [getting started](docs/getting-started.md) to register a VM collector, or follow the [Kubernetes setup guide](docs/kubernetes-setup.md) to install collectors in a cluster. The [usage guide](docs/usage.md) walks through the UI with screenshots.

## Repository layout

| Directory | Contents |
| --- | --- |
| `frontend/` | React, TypeScript, and Vite UI |
| `backend/` | Go API, SQLite storage, graph compiler, and VM monitoring workers |
| `agent/` | OpenTelemetry Collector wrapper with VM and Kubernetes modes |
| `scripts/` | Linux installation scripts |
| `docs/` | Setup, architecture, API, deployment, and troubleshooting guides |

The backend and collector are separate Go modules. Run Go commands from the relevant directory.

## Documentation

- [Usage guide with screenshots](docs/usage.md)
- [Getting started](docs/getting-started.md)
- [Architecture](docs/architecture.md)
- [Deployment](docs/deployment.md) and [container images](docs/container-images.md)
- [Backend API reference](docs/backend/api-reference.md), [backend setup](docs/backend/deployment.md), and [testing](docs/backend/testing.md)
- [Frontend overview](docs/frontend/overview.md)
- [Collector modes and configuration](docs/collector/overview.md)
- [Kubernetes setup](docs/kubernetes-setup.md), [collector management](docs/kubernetes.md), and [management API](docs/kubernetes-api.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Implemented capabilities and future work](docs/roadmap.md)

For hosted deployments, production builds default to `/control-plane/` with a same-origin `/control-plane/api` proxy. Container releases use short Git commit tags; select an available image tag during Kubernetes setup. See the deployment and image guides for details.

## Contributing

Read [CONTRIBUTION.md](CONTRIBUTION.md) for setup, checks, and pull request guidance.

## License

The root project is licensed under the GNU Affero General Public License v3.0. See [LICENSE](LICENSE), [backend/LICENSE](backend/LICENSE), and [agent/LICENSE](agent/LICENSE).
