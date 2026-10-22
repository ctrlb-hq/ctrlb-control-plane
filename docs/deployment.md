# Deployment guide

Deploy the Go backend and static frontend together behind a reverse proxy, or run them on separate hosts with an explicit frontend API URL. SQLite data belongs to the backend process. Collectors run on the VMs or Kubernetes nodes whose telemetry they collect.

For local development, see [getting started](getting-started.md). For image builds and GHCR publication, see [container images](container-images.md).

## Backend

The backend needs a stable `JWT_SECRET`, a writable working directory, and network access to the collectors it manages in VM mode. It listens on port `8096` by default and serves API routes under `/api/`. It does not serve the frontend or terminate TLS.

Build and run a container from the repository root:

```bash
export JWT_SECRET="$(openssl rand -hex 32)"
docker build -t ctrlb-control-plane-backend:local backend
docker volume create ctrlb-backend-data
docker run -d --name ctrlb-backend \
  --restart unless-stopped \
  -p 127.0.0.1:8096:8096 \
  --env JWT_SECRET \
  --mount type=volume,source=ctrlb-backend-data,target=/data \
  ctrlb-control-plane-backend:local
```

Store the generated secret and reuse it when replacing the container. The image runs as UID/GID `10001:10001` and writes `backend.db` and `app.log` in `/data`. Preserve the volume across upgrades. A bind-mounted directory must be writable by that UID/GID.

For a binary deployment, build from `backend/` with CGO enabled and run from a persistent writable directory:

```bash
cd backend
CGO_ENABLED=1 go build -o /tmp/ctrlb-control-plane-backend ./cmd/backend
```

The binary embeds component schemas. It loads `.env` from its working directory, then opens `./backend.db`. Supervise it with your service manager, set `JWT_SECRET`, and keep that directory when replacing the binary. See [backend deployment](backend/deployment.md) for environment variables, logs, and systemd configuration.

`WORKER_COUNT` defaults to `4`. The legacy variable `CHECK_INTERVAL_MINS` defaults to `10` and controls VM monitoring in seconds in the current implementation. Kubernetes check-ins have their own interval.

## Frontend deployment under `/control-plane/`

Production builds default to `/control-plane/`. Without an explicit API override, the frontend calls `/control-plane/api` on the browser's origin. A build can therefore run behind the same proxy configuration on different hostnames.

Build and expose the frontend to a local reverse proxy:

```bash
docker build -t ctrlb-frontend \
  --build-arg VITE_BASE_PATH=/control-plane/ frontend
docker run -d --name ctrlb-frontend \
  --restart unless-stopped \
  -p 127.0.0.1:8080:80 ctrlb-frontend
```

In your HTTPS nginx server, add these locations:

```nginx
location = /control-plane {
    return 308 /control-plane/;
}

location ^~ /control-plane/api/ {
    proxy_pass http://127.0.0.1:8096/api/;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}

location /control-plane/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

The API proxy removes `/control-plane` and preserves `/api`. For example, `/control-plane/api/auth/v1/login` reaches the backend as `/api/auth/v1/login`. The frontend proxy preserves the complete path; the frontend container handles page refreshes below `/control-plane/`.

For a static build, run `npm ci --legacy-peer-deps` and `npm run build` in `frontend/`. Serve `frontend/dist` at `/control-plane/`, with unknown frontend routes falling back to `/control-plane/index.html`. Keep the API location separate from that fallback.

Frontend configuration is applied at build time:

| Variable | Meaning | Production default |
| --- | --- | --- |
| `VITE_BASE_PATH` | Frontend asset and router prefix | `/control-plane/` |
| `VITE_API_BASE_URL` | Complete API root, absolute or relative to the origin | Same-origin `/control-plane/api` |
| `VITE_API_URL` | Legacy backend root, with `/api` appended | Unset |

`VITE_API_BASE_URL` takes precedence over `VITE_API_URL`. For example, use `VITE_API_BASE_URL=https://your-host/control-plane/api` for an explicit hosted API. Docker builds accept these variables as build arguments. For root hosting, set `VITE_BASE_PATH=/` and proxy `/api/` to the backend. Local Vite development defaults to `http://localhost:8096/api`.

## Kubernetes collectors

Create a Kubernetes pipeline and installation in the UI. Set a control-plane URL the cluster can reach and select an available collector image tag or digest. Setup currently defaults to `ghcr.io/ctrlb-hq/ctrlb-otel-collector:latest`; the publication workflows tag images with short Git commits, so replace `latest` with an available tag. See [container images](container-images.md).

For a hosted deployment, the control-plane URL can include the API root, such as `https://your-host/control-plane/api`. Managed collectors accept both backend roots and API roots ending in `/api`. They use HTTPS unless HTTP is explicitly enabled for local development. The telemetry destination URL and its Authorization header are separate settings.

Setup displays the manifest YAML, a browser download, terminal `curl` and `wget` download commands, and apply, rollout, and uninstall commands. Run them on a machine with kubectl access to the target cluster.

Terminal downloads POST the installation token in a JSON body to `/api/agent/v2/installations/<id>/manifest`. The token is not placed in the URL, and the response has `Cache-Control: no-store`. The creation response returns the token once; retain the manifest and uninstall command before closing setup. The manifest and terminal download commands contain the token.

The generated DaemonSet uses host storage for collector state and log offsets. Collectors initiate check-ins approximately every ten seconds and continue with their last working configuration during a control-plane outage. See [Kubernetes setup](kubernetes-setup.md) and the [API contract](kubernetes-api.md) for installation steps, RBAC, host mounts, rollout behavior, and status details.

Uninstall with the saved command before deleting the pipeline. Pipeline deletion removes installation credentials and collector records but leaves Kubernetes resources running. The uninstall command preserves the shared namespace and the node's hostPath state.

## VM collectors

Build and run the collector using [getting started](getting-started.md#connect-a-vm-collector), or use a compatible binary under systemd. Leave `INSTALLATION_ID` unset. Set `BACKEND_URL` to the control-plane root without `/api`, such as `https://your-host/control-plane`.

The VM collector requires an existing writable `AGENT_CONFIG_PATH`. It registers with `/api/agent/v1/agents`, accepts management requests on `3421`, and exposes configured collector metrics on `8888`. The backend must resolve or route to the registered hostname or IP on both ports. These legacy collector routes have no token authentication and belong on a trusted network.

Deleting a VM pipeline detaches its collector records. It does not stop or uninstall the collector service, and it does not push an empty configuration. Stop and disable `ctrlb-collector` on the host when retiring a collector installed by the script.

## Linux install scripts

The scripts in `scripts/` configure Linux systemd services, but their download filenames do not match the current release workflows:

| Component | Installer expects | Current workflow uploads |
| --- | --- | --- |
| Backend | `control-plane-backend-linux-<arch>` | `ctrlb-control-plane-backend-linux-<arch>` on `v*` tags |
| Collector | `ctrlb-collector-linux-<arch>` | `ctrlb-agent-linux-<arch>` on `v*` tags; `ctrlb_collector-linux-<arch>` on `agent/v*` tags |

Both installers download from the repository's latest release. Verify that the release contains the expected asset before using them, or build from source/use containers as described above. The scripts support `amd64` and `arm64`; systemd is required even though they detect the operating system.

If compatible assets are available, the backend script requires `--jwt-secret` when run through a pipe. Its optional flags are `--port`, `--env`, `--workers`, and `--check-interval`. The collector script needs `BACKEND_URL`, `PIPELINE_NAME`, and `STARTED_BY` supplied as environment variables when piped to bash; it cannot reliably prompt with the script occupying standard input.

The backend script stores its binary and `.env` in `/opt/ctrlb/control-plane-backend`, uses `/var/lib/control-plane-backend` as its working directory, and creates `control-plane-backend.service`. The collector script uses `/opt/ctrlb/ctrlb-collector` for its binary, environment, and config, and creates `ctrlb-collector.service`.

## Verify a deployment

1. Sign in at `/control-plane/`, then reload `/control-plane/home` to check the frontend route fallback.
2. Create a pipeline and save a graph change to check authenticated API routing.
3. For Kubernetes, download the manifest from another machine, apply it, and confirm each eligible node appears with separate connection, engine, and rollout status.
4. For VMs, confirm the backend can read port `8888` and deliver a configuration change to port `3421`.
5. Replace the backend container with the same volume and secret, then confirm pipelines and sessions remain available.

Use [troubleshooting](troubleshooting.md) for startup, image pull, proxy, and configuration failures.
