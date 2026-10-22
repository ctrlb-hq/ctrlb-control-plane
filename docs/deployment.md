# 🚀 Deployment Guide for CtrlB Control Plane

This guide walks you through deploying the CtrlB Control Plane components in a production or staging environment.

For backend and collector container builds, GHCR publication, and persistent backend storage, see [Backend and collector images](container-images.md).

---

## 🧱 Deployment Modes

- **All-in-One (Dev/Staging)**: Run backend and frontend on a single VM.
- **Split (Prod)**: Run backend, frontend, and collector independently with service isolation.

---

## 📦 Backend Deployment

### Option A: Quick Install (Recommended)

The easiest way to install the backend is to use our automated installation script. This will download the latest release binary and configure it as a systemd service. Because the script runs non-interactively when piped into `bash`, you must provide the JWT secret in the command.

**Basic Installation (requires JWT secret):**

```bash
curl -fsSL https://raw.githubusercontent.com/ctrlb-hq/ctrlb-control-plane/main/scripts/backend-install.sh | sudo bash -s -- --jwt-secret "your-secret-key-here"
```

**Advanced Options:**

```bash
curl -fsSL https://raw.githubusercontent.com/ctrlb-hq/ctrlb-control-plane/main/scripts/backend-install.sh | sudo bash -s -- \
  --jwt-secret "your-secret-key-here" \
  --port 8096 \
  --env prod \
  --workers 4 \
  --check-interval 10
```

Available options:
- `--jwt-secret <secret>`: JWT secret key (required; cannot prompt when piped to bash)
- `--port <port>`: Port to run the backend on (default: 8096)
- `--env <env>`: Environment (default: prod)
- `--workers <count>`: Number of workers (default: 4)
- `--check-interval <mins>`: Check interval in minutes (default: 10)

The script will:
- Download the appropriate binary for your OS and architecture
- Install it to `/opt/ctrlb/control-plane-backend/`
- Create a systemd service file
- Configure environment variables
- Enable and start the service automatically

### Option B: Manual Installation

#### 1. Build the Binary

```bash
cd backend
go build -o control-plane-backend cmd/backend/main.go
```

#### 2. Install Binary and Create Service

```bash
sudo mkdir -p /etc/control-plane-backend
sudo cp control-plane-backend /usr/local/bin/
sudo cp scripts/systemd/control-plane-backend.service /etc/systemd/system/
```

#### 3. Create Environment File

```bash
sudo tee /etc/control-plane-backend/env > /dev/null <<EOF
JWT_SECRET=<SECRET_HERE>
# Other environment variables as needed
EOF
```

#### 4. Enable & Start Service

```bash
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable control-plane-backend.service
sudo systemctl start control-plane-backend.service
```

---

## Frontend deployment under `/control-plane/`

The production frontend defaults to `/control-plane/`. It uses the browser's origin to reach `/control-plane/api`, so the same build works at `https://ctrlb.explore.ctrlb.ai/control-plane/` and on a staging hostname.

### Build and proxy

Build the frontend container and expose it to the local reverse proxy:

```bash
docker build -t ctrlb-frontend \
  --build-arg VITE_BASE_PATH=/control-plane/ frontend
docker run -d --name ctrlb-frontend -p 127.0.0.1:8080:80 ctrlb-frontend
```

Run the backend on port `8096` with `JWT_SECRET` set and persistent storage for its database. In the HTTPS nginx server for `ctrlb.explore.ctrlb.ai`, add these locations:

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

The API proxy removes `/control-plane` and preserves `/api`. For example, `/control-plane/api/auth/v1/login` reaches the backend as `/api/auth/v1/login`. The frontend proxy preserves the complete path. Its container serves assets and page refreshes beneath `/control-plane/`.

For a static build, run `npm ci --legacy-peer-deps` and `npm run build` in `frontend`. Serve `frontend/dist` at `/control-plane/`, with unknown frontend routes falling back to `/control-plane/index.html`. Keep the API proxy above separate from the frontend fallback.

Frontend configuration is applied at build time:

| Variable | Meaning | Production default |
| --- | --- | --- |
| `VITE_BASE_PATH` | Frontend asset and router prefix | `/control-plane/` |
| `VITE_API_BASE_URL` | Full API root, absolute or relative to the origin | Same-origin `/control-plane/api` |
| `VITE_API_URL` | Legacy backend root, with `/api` appended | Unset |

Use `VITE_API_BASE_URL=https://ctrlb.explore.ctrlb.ai/control-plane/api` to set the API root explicitly. It takes precedence over `VITE_API_URL`. Docker builds accept the same variables as build arguments. For root hosting, set `VITE_BASE_PATH=/` and proxy `/api/` to the backend. Local Vite development defaults to `http://localhost:8096/api`; existing `VITE_API_URL=http://localhost:18096` demo commands still work.

### Install Kubernetes collectors from a remote machine

After creating a Kubernetes installation, setup displays the manifest YAML with a copy button, plus `curl` and `wget` commands. The existing browser download remains available. Run either command on the VM or host where you use `kubectl`, then run the displayed apply and rollout commands there.

The collector-reachable control-plane URL defaults to the public API root, such as `https://ctrlb.explore.ctrlb.ai/control-plane/api`. Both legacy backend roots and API roots ending in `/api` work for Kubernetes collector check-ins. The cluster nodes must be able to reach this URL and pull the collector image you select.

Terminal downloads authenticate with the installation token in a JSON POST body. They do not need a browser session, and the token is never included in the URL. The server returns the same YAML as the browser preview with `Cache-Control: no-store`. Keep the YAML or a download command before closing setup, because setup only receives the installation token once. Both contain that token.

To verify a deployment, sign in beneath `/control-plane/`, reload `/control-plane/home`, and create an installation. Copy a terminal download command to another machine and check that it returns YAML from `/control-plane/api/agent/v2/installations/<id>/manifest`. Apply it and confirm that node collectors appear in setup.

---

## 🛰️ Collector Installation

Collector installation steps are provided via the UI. Once a new collector is created, a corresponding installation command with a unique token and configuration is displayed.
The control plane will wait for the collector to complete the setup and come online.
Each collector:

- Fetches its configuration from the control plane
- Exposes health metrics on `:8888`

---

## 🔒 Security Notes

- Kubernetes collector check-ins and terminal manifest downloads require an installation token.
- Use HTTPS reverse proxies (e.g., Nginx, Caddy) in production.

---

For more details, refer to:

- [Architecture Overview](architecture.md)
- [API Reference](backend/api-reference.md)
- [Troubleshooting Guide](troubleshooting.md)
