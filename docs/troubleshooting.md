# Troubleshooting

## Backend does not start

If the log says `JWT_SECRET is not set in environment`, export `JWT_SECRET` or put it in `.env` in the backend's working directory. The backend exits when it is empty. Keep the same secret across restarts so existing tokens remain valid.

Run the entry point from the backend module:

```bash
cd backend
go run ./cmd/backend
```

For `Failed to open log file` or SQLite permission errors, check that the process can write to its working directory. Both `app.log` and `backend.db` use relative paths. In the backend container that directory is `/data`, owned by UID/GID `10001:10001`. A bind mount must allow that identity to write.

If a source build fails with a SQLite or CGO error, install a C compiler and build with `CGO_ENABLED=1`. See [backend deployment](backend/deployment.md) for toolchain and storage details.

## Backend installer cannot find a release asset

The current `scripts/backend-install.sh` requests `control-plane-backend-linux-<arch>`, while the current release workflow publishes `ctrlb-control-plane-backend-linux-<arch>`. Use a [source build or container](backend/deployment.md) until the names agree. An HTTP 404 for that asset is separate from backend startup or network configuration.

## Check whether the API is reachable

The backend's default port is `8096`. It does not expose a dedicated health endpoint. This request should reach the router and return `401` without a token:

```bash
curl -i http://localhost:8096/api/frontend/v2/agents
```

For the backend container:

```bash
docker logs ctrlb-backend
```

For a systemd installation using the checked-in script's service name:

```bash
sudo systemctl status control-plane-backend --no-pager
sudo journalctl -u control-plane-backend -n 100 --no-pager
```

A remote machine or pod cannot use `localhost` to reach the backend on your laptop. Use an address reachable from that machine. If a reverse proxy adds a path prefix, include that prefix in the configured backend URL. The backend container example binds to `127.0.0.1`; provide a reachable proxy or adjust the published address for remote collectors.

## Browser cannot log in or load data

Check the browser's network requests against the backend URL configured for the frontend. Follow [frontend configuration](frontend/overview.md) for the current Vite variables and proxy behavior. Verify that a proxy forwards `/api` requests to the backend and that an HTTPS frontend does not call a plain HTTP backend directly.

Protected frontend API calls need `Authorization: Bearer <access_token>`. Access JWTs expire after 15 minutes; refresh JWTs expire after 30 days. Log in again if the refresh token is invalid or if `JWT_SECRET` changed. Some older handlers return HTTP `200` with an `error` field, so inspect the response body too.

## VM collector fails to register

For `Failed to register with backend server`, check the collector's log for the HTTP status or connection error. In VM mode, `BACKEND_URL` is required. `PIPELINE_NAME` and `STARTED_BY` are optional; omitting the pipeline name leaves the collector unassigned.

The file at `AGENT_CONFIG_PATH` must exist before startup. Its default is `./config.yaml`. A collector process can fail before registration if that file is missing or its telemetry configuration is invalid.

From the collector host, check backend reachability:

```bash
curl -i http://your-backend:8096/api/frontend/v2/agents
```

A `401` response confirms the API is reachable. It is expected here because this check omits the browser token. VM collector registration itself uses `/api/agent/v1/agents` and does not require a JWT. See [collector operation](collector/overview.md) for configuration and service commands.

## VM collector registers but commands or metrics fail

Registration travels from collector to backend. Commands, configuration updates, and VM metrics need connections in the other direction. From the backend host, verify the registered hostname or IP:

```bash
curl -i http://collector-host:3421/agent/v1/config
curl -sS http://collector-host:8888/metrics
```

The configuration route accepts only `POST`, so the read-only `GET` check above should return `405` when the collector listener is reachable. The backend tries the registered hostname first and then the IP. Check DNS, routing, firewall rules, and whether port `3421` is listening. Metrics require the collector's Prometheus endpoint on `8888`.

After three failed monitoring attempts, the backend removes a VM collector from the monitoring queue. It updates an existing metrics row to `disconnected`; a newly registered collector that has never been scraped successfully has no metrics row and can remain `unknown` in API results. Either state can require restarting monitoring. Once connectivity is restored, use the UI's restart-monitoring action or `POST /api/frontend/v2/agents/{id}/restart-monitoring` with an access JWT. Kubernetes collectors use check-ins instead of this monitoring queue.

Deleting inventory for a collector recorded as `disconnected` skips the shutdown request. Stop and disable its service on the host to confirm it has stopped exporting telemetry.

`CHECK_INTERVAL_MINS` controls the VM scrape interval in seconds in the current implementation. Its default `10` means 10 seconds.

## A port is already in use

Check the process that owns the port:

```bash
sudo lsof -nP -iTCP:8096 -sTCP:LISTEN
sudo lsof -nP -iTCP:3421 -sTCP:LISTEN
sudo lsof -nP -iTCP:8888 -sTCP:LISTEN
```

The backend uses `8096` by default and accepts a `PORT` environment variable. The collector management listener uses `3421`; it does not read a `PORT` environment override. Stop the duplicate collector or isolate it in a separate network namespace. Port `8888` belongs to collector self-metrics and can also conflict with another process.

## Kubernetes collector or rollout is unhealthy

Start with the installation's collector details in the UI. `connection_status`, `engine_status`, `applied_revision`, `attempted_revision`, and `config_error` describe different failures. The backend marks a collector disconnected after 35 seconds without a check-in.

- For no enrollment or `401`, confirm the installation ID and installation token match. A user JWT cannot authenticate collector check-ins.
- For startup validation errors, confirm `BACKEND_URL`, `K8S_NODE_NAME`, and the installation token or token file. Managed mode requires HTTPS except for loopback URLs or an explicit `ALLOW_INSECURE_HTTP=true` setting for local development.
- For image pull errors, use a published commit tag or digest. The manifest API's default `:latest` is not published by the current image workflow. Private GHCR packages also need image pull credentials or public package visibility.
- For a failed rollout, inspect `config_error`. The collector keeps its last working configuration and attempts to restore it if a replacement fails to start. A rejected revision is not retried on every check-in. Save a corrected configuration that produces a new desired revision; saving the same compiled configuration does not create a new revision.
- For readiness failures, inspect pod logs and check whether the telemetry engine is running. Readiness reports local engine state; it does not prove that exported telemetry reached its destination.
- For missing pod logs or kubelet metrics, check host log mounts, RBAC, exporter endpoints, and kubelet TLS errors.

The generated DaemonSet runs in namespace `observability`. Inspect its collectors with:

```bash
kubectl -n observability get pods -l app.kubernetes.io/name=ctrlb-collector
kubectl -n observability logs -l app.kubernetes.io/name=ctrlb-collector --all-containers --tail=100
```

Use the [Kubernetes setup guide](kubernetes-setup.md) and [API contract](kubernetes-api.md) for detailed checks. Follow the [uninstall instructions](kubernetes-setup.md#uninstall-this-installation) before deleting the pipeline. The installation-specific commands preserve the shared namespace. Pipeline deletion removes its installation records and tokens from the backend but leaves Kubernetes resources running.
