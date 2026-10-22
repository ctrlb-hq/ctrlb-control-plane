# Getting started

This guide runs the control plane from source, then connects a VM collector or a Kubernetes installation. For hosted services and persistent container storage, see [deployment](deployment.md).

For a walkthrough of the UI, including pipeline editing and collector status, see the [usage guide with screenshots](usage.md).

## Prerequisites

- Go 1.26 or later to build both Go modules. The backend module declares Go 1.23 and a Go 1.24.4 toolchain; the collector requires Go 1.26.
- A C compiler and CGO enabled for the backend's `go-sqlite3` driver. Linux builds need development headers, such as those provided by `build-essential` on Debian/Ubuntu.
- Node.js 24 and npm for the frontend.
- Git and OpenSSL for the examples below.

Kubernetes installation also needs kubectl and access to a Linux cluster. Docker is optional for the source-based setup and is required when building container images.

```bash
git clone https://github.com/ctrlb-hq/ctrlb-control-plane.git
cd ctrlb-control-plane
```

Run each component in a separate terminal. Commands start from the repository root unless stated otherwise.

## Start the backend

```bash
cd backend
export JWT_SECRET="$(openssl rand -hex 32)"
go run ./cmd/backend
```

The backend listens on `http://localhost:8096`. It creates `backend.db` and `app.log` in the working directory and loads component schemas embedded in the binary. It also loads a `.env` file in that directory if one exists. Save your development secret in `backend/.env` or export the same value next time if you want existing tokens to stay valid.

`PORT` changes the listening port. [Backend deployment](backend/deployment.md) lists all supported variables, including the legacy `CHECK_INTERVAL_MINS` setting, which the current code interprets as seconds.

## Start the frontend

```bash
cd frontend
npm ci --legacy-peer-deps
npm run dev
```

Open [http://localhost:3030](http://localhost:3030), create an account, and sign in. No `.env.example` file is required or supplied. The development API defaults to `http://localhost:8096/api`.

To use a different API, create `frontend/.env.local` and restart Vite:

```dotenv
VITE_API_BASE_URL=http://localhost:18096/api
```

`VITE_API_BASE_URL` includes `/api`. The older `VITE_API_URL` setting takes a backend root and appends `/api`; prefer the explicit API setting for new deployments. See [frontend configuration](frontend/overview.md) for production base paths.

## Connect a VM collector

Start the collector with `PIPELINE_NAME` set to create a new default pipeline during registration. The backend does not look up an existing pipeline by this name. Alternatively, omit `PIPELINE_NAME` to register an unassigned collector, then use the [authenticated pipeline attachment API](backend/api-reference.md#pipelines) to assign it.

```bash
cd agent
cp internal/config/config.yaml /tmp/ctrlb-vm-config.yaml
export BACKEND_URL=http://localhost:8096
export PIPELINE_NAME=local-vm
export STARTED_BY=dev@example.com
export AGENT_CONFIG_PATH=/tmp/ctrlb-vm-config.yaml
go run ./cmd/ctrlb_collector
```

Choose a new name in place of `local-vm`. The configuration path must exist before startup and remain writable, because registration and later updates replace its contents. Copying the sample keeps these updates out of the tracked source file. Leave `INSTALLATION_ID` unset for VM mode.

VM `BACKEND_URL` is the backend root without `/api`, because the collector appends `/api/agent/v1/...`. For a hosted proxy, use a root such as `https://your-host/control-plane`.

The collector exposes its management API on port `3421` and its configured Prometheus metrics endpoint on `8888`. The backend must reach those ports using the hostname or IP the collector registers. Open the pipeline in the UI, edit its graph, save, and confirm the collector receives the new configuration. VM endpoints use the legacy unauthenticated management protocol; keep them on a trusted network.

The Linux systemd installer is another option, but its expected binary filenames differ from the current release workflows. Use the source command above or verify release asset compatibility before using an installer command from the UI. [Deployment](deployment.md#linux-install-scripts) explains the mismatch.

## Connect Kubernetes collectors

Follow the [Kubernetes setup guide](kubernetes-setup.md). Choose Kubernetes during pipeline creation, set the collector-reachable control-plane URL, and select a published collector image tag or digest. Setup currently suggests `ghcr.io/ctrlb-hq/ctrlb-otel-collector:latest`, while the publication workflows produce short commit tags. Replace that suggestion with a tag that exists. Copy or download the generated manifest and apply it from a machine with kubectl access.

Kubernetes mode uses outbound check-ins with an installation token and a shared desired configuration. Save the generated YAML and uninstall command. Uninstall cluster resources before deleting the pipeline, because deleting control-plane records does not stop the running DaemonSet.

## Check your changes

Run Go checks from each module and frontend checks from `frontend/`. See [contribution guidance](../CONTRIBUTION.md) and [backend testing](backend/testing.md) for the commands and their coverage.

For connection, configuration, or image pull failures, use [troubleshooting](troubleshooting.md).
