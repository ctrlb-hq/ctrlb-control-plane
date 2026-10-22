# Frontend

The frontend is a React 18 and TypeScript application built with Vite 7. It provides login and registration, a pipeline list, a React Flow editor, VM collector installation instructions, and Kubernetes installation and rollout status.

For installation commands, environment variables, checks, and a Docker example, use the [frontend README](../../frontend/README.md). Local development runs at `http://localhost:3030` and uses `http://localhost:8096/api` unless configured otherwise.

## Source layout

```text
frontend/
├── public/ctrlb-logo.png
├── src/
│   ├── App.tsx
│   ├── main.tsx
│   ├── config.ts
│   ├── constants.ts
│   ├── components/
│   │   ├── auth/
│   │   ├── pipelines/
│   │   │   ├── HomePage.tsx
│   │   │   ├── create/
│   │   │   ├── editor/
│   │   │   ├── kubernetes/
│   │   │   └── overview/
│   │   └── ui/
│   ├── context/
│   ├── hooks/
│   ├── services/
│   ├── types/
│   └── utils/
├── .env.production
├── Dockerfile
├── package.json
└── vite.config.ts
```

`main.tsx` mounts the application, the pipeline overview and graph providers, and toast notifications. `App.tsx` defines browser routes. The `@/` import alias resolves to `src/` in both TypeScript and Vite.

`components/pipelines/create/` contains the VM and Kubernetes setup selector and VM installation flow. `editor/` contains the graph editor, component selection, and configuration forms. `kubernetes/` contains installation setup, manifest download, and per-node status views. `overview/` contains the pipeline table, VM details and health charts, YAML preview, and deletion dialog.

`context/useGraphFlowContext.tsx` stores graph nodes, edges, and unsaved changes. `context/usePipelineDetailContext.tsx` stores pipeline overview data. HTTP requests live in `services/`; shared types, URL resolution, manifest command generation, and YAML conversion have separate modules.

## Authentication and routing

| Route       | Behavior                                                   |
| ----------- | ---------------------------------------------------------- |
| `/login`    | Signs in through `/api/auth/v1/login`.                     |
| `/register` | Creates an account through `/api/auth/v1/register`.        |
| `/home`     | Shows pipelines. Requires an `authToken` in local storage. |

These paths are relative to `VITE_BASE_PATH`. A production build with the default prefix uses `/control-plane/login`, `/control-plane/register`, and `/control-plane/home`. Unknown routes redirect to login. Users with a stored access token redirect from the public authentication pages to home or their attempted protected route.

The authentication service stores access and refresh tokens in local storage. `utils/axiosInstance.ts` adds the bearer token to frontend API requests. A `401` response triggers a refresh request; other requests that fail during that refresh wait for its result and retry with the new access token. Logout clears local storage. The backend remains responsible for validating tokens on protected endpoints.

## API URLs and hosting

`config.ts` resolves one API root and derives the service prefixes:

| Service                                    | Prefix appended to the API root |
| ------------------------------------------ | ------------------------------- |
| Authentication                             | `/auth/v1`                      |
| Browser pipeline and installation requests | `/frontend/v2`                  |
| Terminal manifest download commands        | `/agent/v2`                     |

`VITE_API_BASE_URL` includes `/api` and takes precedence over the legacy `VITE_API_URL`, which expects a backend root and appends `/api`. The default API root is `http://localhost:8096/api` in development. Production uses the current browser origin and the app base path plus `/api`, such as `https://example.com/control-plane/api`.

The production base path is `/control-plane/` in `.env.production` and the Docker build. Vite uses it for assets, and React Router uses Vite's `BASE_URL` for navigation. Changing the prefix or backend URL requires a new frontend build.

The Docker image supplies an Nginx fallback to the app's `index.html` for browser routes. It does not proxy API requests. A deployment using the default API root must route `/control-plane/api/*` to the backend's `/api/*` and route UI requests to the frontend. See [deployment](../deployment.md).

## Pipeline workflows

The pipeline list loads both pipelines and Kubernetes installations every 10 seconds. It also refreshes after setup, Kubernetes configuration saves, and deletion. VM configuration changes appear on the regular poll. Failed refreshes retain the last data and show a stale-data message.

VM setup asks for a pipeline name and Linux or macOS, then generates an installation command using the resolved API root. Registration with `PIPELINE_NAME` creates a new default pipeline and attaches the collector to it; it does not join an existing pipeline by name. The generated command uses a Linux systemd installer whose expected binary filenames differ from the current release workflows. Use the [source setup](../getting-started.md#connect-a-vm-collector) or verify compatible assets before running that command. VM details show the associated collector, pipeline configuration, and health metrics. The refresh control restarts backend monitoring for collectors whose status needs rechecking.

Kubernetes setup creates a pipeline preset and an installation. It accepts a collector-reachable management URL, collector image, OTLP endpoints, an optional Authorization header, and a kubelet certificate verification setting. The management URL must be reachable from cluster nodes; a browser's `localhost` usually does not meet that requirement. Management requests and exported telemetry use separate endpoints. Follow [Kubernetes setup](../kubernetes-setup.md) for the field values and installation commands.

The form currently suggests `ghcr.io/ctrlb-hq/ctrlb-otel-collector:latest`. The publication workflows use short Git commit tags, so select a published tag or digest before generating the manifest. Kubelet certificate verification is enabled by default; the checkbox disables it for installations that require it, such as a local kind cluster with a self-signed certificate.

The setup view generates a manifest for download and also provides `curl` and `wget` download commands. The installation token stays in component state and appears in the generated manifest and terminal download request body. Save the manifest before closing setup. The UI does not persist that token for later retrieval. The telemetry Authorization value is sent to the backend with the pipeline and cleared from the setup form after the pipeline is saved.

The `View/edit pipeline configuration` button becomes available after a manifest download or a successful `Copy` action on the YAML or a terminal download command. Copying the text manually does not enable the button. Changing the management URL or collector image regenerates the manifest and requires saving it again. The UI provides apply and rollout commands for the generated file; it does not apply resources to the cluster.

The graph editor loads component configuration and UI schemas from the backend. Users add receivers, processors, and exporters, connect them, edit their configuration, and review changes before saving. Saving a Kubernetes graph stores the desired configuration and its revision. Collectors apply a changed revision at their next check-in, so a successful save means the configuration is stored, with rollout status following separately.

Kubernetes details refresh every 10 seconds and report connection status, engine status, desired/applied/attempted revisions, configuration errors, and available self-metrics for each node. A running engine can still have a failed configuration rollout after restoring its previous configuration. Confirm exported telemetry at the destination. See [Kubernetes collector management](../kubernetes-v1.md) for the rollout behavior.

Deleting a pipeline removes control-plane records and revokes its Kubernetes installation credentials. Running collectors may continue with their last configuration. The deletion dialog provides per-installation Kubernetes uninstall commands, which remove that installation's resources while preserving the shared `observability` namespace.

## Verification and troubleshooting

`npm test` runs Node-based utility tests for API URL handling, shell-safe manifest downloads, installation-specific resource commands, rollout summaries, metric formatting, and YAML output. `npm run build` checks TypeScript and bundles the app. Use `npm run lint` and `npm run format:check` for the existing static checks. The frontend has no configured browser test suite.

| Symptom                                                      | Check                                                                                                                  |
| ------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------- |
| Dependency installation fails on peer dependencies           | Use `npm ci --legacy-peer-deps`.                                                                                       |
| Node version fails during Vite startup or tests              | Use Node.js 24, matching the Docker build.                                                                             |
| API requests return `404`                                    | Inspect the request URL. `VITE_API_BASE_URL` must include `/api`; `VITE_API_URL` must omit the suffix the app appends. |
| Production preview sends requests to the preview server      | Build with `VITE_API_BASE_URL` pointing to the backend, or route the default API path through a reverse proxy.         |
| Refreshing `/control-plane/home` returns a server `404`      | Configure an SPA fallback to `/control-plane/index.html`.                                                              |
| Browser reaches the backend but Kubernetes collectors do not | Check the management URL from a cluster node, including its path prefix.                                               |
| Status shows a refresh error                                 | Check the backend and network request. The last successful response remains visible while polling retries.             |
