# Backend deployment

The backend exposes the control-plane API on port `8096`. It stores users, pipeline graphs, VM collectors, and Kubernetes installations in SQLite. Run one backend process against each database file.

## Requirements

- A Go toolchain compatible with [backend/go.mod](../../backend/go.mod), which declares Go `1.23.0` and toolchain `go1.24.4`.
- A C compiler with CGO enabled. The SQLite driver requires it.
- A writable working directory for `backend.db` and `app.log`.

The backend Dockerfile uses Go `1.26` and includes the C build dependencies. A local source build does not require the SQLite command-line tool.

## Environment

The backend loads `.env` from its working directory. Exported environment variables take precedence over values in that file.

| Variable | Default | Purpose |
| --- | --- | --- |
| `JWT_SECRET` | Required | Signs and verifies access and refresh tokens. Keep the value stable across restarts. |
| `PORT` | `8096` | HTTP listening port. The server binds on all interfaces. |
| `WORKER_COUNT` | `4` | Number of workers that scrape VM collector metrics. Use a positive integer. |
| `CHECK_INTERVAL_MINS` | `10` | VM monitoring interval, interpreted as **seconds** by the current implementation despite the variable name. Use a positive integer. |
| `ENV` | `prod` | Environment label loaded at startup. It does not configure TLS, CORS, or logging. |

Unset or nonnumeric worker and interval values fall back to the defaults. Zero and negative values are not validated, so use positive integers. Kubernetes check-ins use the separate poll interval returned by the API, currently 10 seconds.

For example, create `backend/.env` before running from `backend/`:

```dotenv
PORT=8096
JWT_SECRET=replace-with-a-random-secret
WORKER_COUNT=4
CHECK_INTERVAL_MINS=10
ENV=prod
```

The process exits if `JWT_SECRET` is empty. Generate a secret once with `openssl rand -hex 32` and store it in your deployment configuration.

## Run from source

From the repository root:

```bash
cd backend
go mod download
go run ./cmd/backend
```

To build a binary instead:

```bash
cd backend
CGO_ENABLED=1 go build -trimpath -o /tmp/ctrlb-control-plane-backend ./cmd/backend
/tmp/ctrlb-control-plane-backend
```

Each example starts in the repository root. The binary uses the working directory where you launch it, so keep that directory consistent across restarts.

There is no backend `/health` endpoint. A request to an authenticated route without a token should return `401` and confirms that the router is responding:

```bash
curl -i http://localhost:8096/api/frontend/v2/agents
```

## Run in Docker

From the repository root, use the `backend/.env` file created above with your generated secret and mount a named volume:

```bash
docker build -t ctrlb-control-plane-backend:local backend
docker volume create ctrlb-backend-data
docker run -d --name ctrlb-backend \
  -p 127.0.0.1:8096:8096 \
  --env-file backend/.env \
  --mount type=volume,source=ctrlb-backend-data,target=/data \
  ctrlb-control-plane-backend:local
```

The container runs as UID/GID `10001:10001` with `/data` as its working directory. Bind mounts must be writable by that identity. Keep the volume when replacing the container. If you change `PORT` in the environment file, update the container port in `-p` to match.

This example exposes the API only on the host's loopback interface. Remote collectors need a reachable backend URL through a reverse proxy or an appropriate published address. See [container images](../container-images.md) for GHCR publication, image tags, and digest pinning.

## Database and logs

The database path is fixed at `./backend.db`. Startup creates missing tables and applies the Kubernetes migration, including the `management_mode` column on existing agent rows. It also refreshes component and UI schemas from the embedded assets. There is no separate migration command.

Back up `backend.db` before upgrading. Stop the backend before copying the file, or use SQLite's backup facilities. Restore into the same working directory with permissions that allow the backend to write.

Logs go to standard output and to `./app.log`. The file uses JSON encoding. The current logger writes debug-level messages and has no built-in file rotation.

## Service deployment

For a systemd deployment, point `ExecStart` at the compiled binary, set an `EnvironmentFile` containing the values above, and set `WorkingDirectory` to a persistent writable directory. Run the service under an account that owns that directory.

The checked-in [backend installer](../../scripts/backend-install.sh) downloads `control-plane-backend-linux-<arch>` from the latest GitHub release. The current release workflow publishes `ctrlb-control-plane-backend-linux-<arch>` instead. The installer needs that asset naming mismatch resolved before it can install binaries from those releases. Use a source build or container in the meantime. Its intended service is `control-plane-backend`, with configuration in `/opt/ctrlb/control-plane-backend/.env` and data in `/var/lib/control-plane-backend`.

Terminate HTTPS at your reverse proxy. The backend itself serves HTTP and allows all CORS origins. User registration and VM collector registration do not require tokens; the frontend API requires access JWTs, and managed collector check-ins require installation tokens. Configure network access for the intended deployment.

VM collectors require backend-to-collector access on `3421` for commands and `8888` for metrics. Kubernetes collectors initiate outbound check-ins and do not need those ports exposed to the backend. Follow the [Kubernetes setup guide](../kubernetes-setup.md) to create an installation, generate its manifest, and deploy it to a cluster.

See [troubleshooting](../troubleshooting.md) for common startup and connectivity failures.
