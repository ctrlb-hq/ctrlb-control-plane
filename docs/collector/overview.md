# Collector guide

The CtrlB collector embeds OpenTelemetry Collector. It receives a pipeline configuration from the control plane and sends collected telemetry to the exporters in that configuration. The management URL and telemetry ingestion URL are separate values.

The binary has two runtimes:

| Mode | Selection | Management |
| --- | --- | --- |
| VM | `INSTALLATION_ID` is absent or empty | Registers at startup; backend calls the collector's HTTP API and scrapes its metrics |
| Kubernetes | `INSTALLATION_ID` is nonempty | Sends authenticated outbound check-ins; receives revisions and reports the result |

Both load an optional `.env` from the working directory without replacing environment variables already set. The program uses environment variables rather than command-line configuration flags.

## Build and installation

Use Go 1.26 to [build and run a VM collector from source](../../agent/README.md#build-and-run-a-vm-collector). Build a container from the repository root with:

```bash
docker build -t ctrlb-otel-collector:local agent
```

The image includes the same binary and a VM bootstrap configuration at `/app/config.yaml`. See [container images](../container-images.md) for publication, platform support, and image references.

For Kubernetes, follow the [setup guide](../kubernetes-setup.md) and use the installation manifest generated during pipeline setup.

The VM installer downloads `ctrlb-collector-<os>-<arch>` from the latest GitHub release. Current release workflows emit `ctrlb-agent-linux-<arch>` for `v*` tags and `ctrlb_collector-<os>-<arch>` for `agent/v*` tags. Check asset compatibility before using the installer; a source build avoids that naming mismatch.

## VM configuration

| Variable | Default | Meaning |
| --- | --- | --- |
| `BACKEND_URL` | Required | Control-plane root, for example `http://localhost:8096` |
| `AGENT_CONFIG_PATH` | `./config.yaml` | Writable Collector configuration file; must exist at startup |
| `PIPELINE_NAME` | Empty | Name for a new default pipeline created during registration |
| `STARTED_BY` | Empty | Initiator sent during registration |

The collector starts its bootstrap configuration, posts its platform and host details to `/api/agent/v1/agents`, then writes the returned default configuration to disk. If `PIPELINE_NAME` is nonempty, the backend creates a new default pipeline and attaches the collector; it does not look up an existing pipeline by that name. With an empty value, assign the enrolled collector through the authenticated [pipeline attachment API](../backend/api-reference.md#pipelines). The collector watches its configuration file for changes and validates updates received through its HTTP API before saving them.

VM registration appends `/api/agent/v1/agents` to `BACKEND_URL`. Use the server root without a trailing slash or `/api`. For a deployment exposed at `/control-plane/api`, use `https://ctrlb.explore.ctrlb.ai/control-plane` as the VM root. The backend needs a route back to the collector's reported hostname or IP.

VM mode listens on `0.0.0.0:3421` and exposes these routes:

| Request | Behavior |
| --- | --- |
| `POST /agent/v1/start` | Start the embedded Collector |
| `POST /agent/v1/stop` | Stop the embedded Collector while keeping the wrapper running |
| `POST /agent/v1/shutdown` | Shut down the collector and its HTTP server |
| `POST /agent/v1/config` | Validate and save a full Collector configuration JSON object |

These collector routes have no authentication middleware. Keep access to trusted control-plane hosts through network policy or firewall rules. The backend also scrapes `http://<collector-host>:8888/metrics`; generated configurations expose that Prometheus endpoint on all interfaces. VM mode writes console logs and `app.log` in its working directory.

## Kubernetes configuration

Generated manifests set these variables. You normally do not need to assemble them yourself.

| Variable | Required by runtime | Meaning |
| --- | --- | --- |
| `INSTALLATION_ID` | Yes | Selects managed mode and identifies the installation |
| `BACKEND_URL` | Yes | Collector-reachable management root or API base URL |
| `INSTALLATION_TOKEN_FILE` or `INSTALLATION_TOKEN` | Yes | Installation bearer token; the file takes precedence and is reread on each check-in |
| `K8S_NODE_NAME` | Yes | Stable collector identity within the installation |
| `STATE_DIR` | No | Defaults to `/var/lib/ctrlb`; stores the working configuration, revisions, and log offsets |
| `ALLOW_INSECURE_HTTP` | No | Set exactly `true` to permit non-loopback HTTP management during local development |
| `K8S_NODE_UID` | No | Optional descriptive node UID |
| `POD_NAMESPACE`, `POD_NAME`, `POD_UID`, `POD_IP` | No | Pod metadata for check-ins; namespace, name, and UID also exclude the collector's logs |
| `K8S_NODE_IP`, `K8S_CLUSTER_NAME` | No | Used by the generated kubelet metrics and resource-metadata configuration |

The generated preset needs the pod metadata, node IP, and cluster name even though the management runtime does not validate them as required environment variables. The manifest supplies these values using the Kubernetes downward API and installation settings. It mounts the token at `/var/run/secrets/ctrlb/token`.

Managed mode requires HTTPS unless the URL is loopback or `ALLOW_INSECURE_HTTP=true`. URLs cannot include credentials, query parameters, or fragments. A root such as `https://control-plane.example.com` gains `/api`; a base ending in `/api`, including `https://ctrlb.explore.ctrlb.ai/control-plane/api`, keeps that prefix. The client refuses HTTP redirects.

Each collector checks in about every ten seconds with twenty percent jitter. Failed check-ins retry with exponential backoff from one to thirty seconds, with jitter. The collector keeps its working configuration during a backend outage. Identity is the installation ID plus node name, so replacing a pod on the same node keeps its collector record.

Before applying a revision, the collector validates its configuration. It then stops the old Collector, starts the candidate, confirms it is running, and persists the working state. A failed startup or persistence step triggers an attempt to restore the previous working configuration. Configuration changes can briefly interrupt collection while the embedded Collector restarts.

The collector records a failed attempted revision and does not repeatedly apply it. A new desired revision allows another attempt. On restart it validates and starts its saved working configuration before contacting the backend. Unusable cached configuration leaves its revisions empty so it can request a fresh copy. The runtime also retries a committed working configuration if the embedded Collector stops unexpectedly.

The generated DaemonSet mounts `/var/lib/ctrlb/installations/<installation-id>` from the node at `/var/lib/ctrlb` in the pod. It stores log offsets under `offsets/` and working state in `state.json`. State and candidate configuration files have mode `0600` and can contain exporter credentials. Node replacement loses this local state.

## Health and metrics

Managed mode serves probes on port `3421`:

| Request | Response |
| --- | --- |
| `GET /healthz` | HTTP 200 while the management server is alive |
| `GET /readyz` | HTTP 200 when the Collector is running; HTTP 503 while waiting or failed |

A collector running cached configuration stays ready during a control-plane outage. Managed mode does not expose the VM lifecycle or configuration routes.

The managed runtime samples `http://127.0.0.1:8888/metrics` and includes CPU percent, memory bytes, and available per-second sent log, metric, and span counts in its check-ins. These are Collector process and export metrics, not host CPU or memory usage. A rate needs two counter samples; missing measurements stay unavailable. The backend does not scrape Kubernetes pod or node addresses.

The Kubernetes preset separately collects node, pod, and container metrics every thirty seconds and new pod stdout/stderr logs across namespaces. It excludes the collector's own logs and adds cluster and Kubernetes resource metadata. See [Kubernetes behavior](../kubernetes.md) and the [API contract](../kubernetes-api.md) for rollout status and manifest details.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| VM startup reports a missing config file | Copy the bootstrap file and set `AGENT_CONFIG_PATH` before starting |
| VM registration fails | Verify `BACKEND_URL`, backend reachability, and the reported hostname/IP |
| VM status or updates fail | Allow backend access to collector ports `3421` and `8888` |
| Kubernetes pod waits for configuration | Check the installation token, management URL, outbound access, and collector logs |
| Kubernetes rollout fails but engine is running | Read `config_error`; the collector may be running its previous working revision |
| Image pull fails | Use a published short commit tag or digest and check registry access and node architecture |
| Host port `3421` or `8888` is busy | Stop the conflicting local process; these listener addresses are not environment-configurable |

Deleting a pipeline removes its installation credentials and inventory, but running collectors can keep their cached configuration. Uninstall the installation's Kubernetes resources before deleting its pipeline. Follow the [uninstall instructions](../kubernetes-setup.md#uninstall-this-installation) to remove installation-specific resources while preserving the shared namespace.
