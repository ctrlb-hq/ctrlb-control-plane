# Kubernetes v1 management API

This contract implements the [v1 scope](kubernetes-v1.md), [collector check-ins](adr/0001-kubernetes-collector-check-ins.md), and [per-collector rollouts](adr/0002-kubernetes-configuration-rollouts.md). Terms follow the [glossary](../CONTEXT.md).

Management requests use the external control-plane HTTPS URL. Telemetry uses its separately configured ingestion destination. A collector checks in about every ten seconds with client-side jitter and retries during outages. The management API does not pause pods, upgrade collectors, or uninstall Kubernetes resources.

## Installation creation

Create the editable collection pipeline first with `POST /api/frontend/v2/kubernetes-pipelines`, authenticated with the existing user access token.

```json
{
  "name":"Kubernetes demo",
  "endpoint":"https://ctrlb.explore.ctrlb.ai/engine/api/default",
  "logs_endpoint":"https://ctrlb.explore.ctrlb.ai/engine/api/default/control_plane/_otel/v1/logs",
  "metrics_endpoint":"https://ctrlb.explore.ctrlb.ai/engine/api/default/v1/metrics",
  "headers":{"Authorization":"Basic CHANGE_ME"},
  "kubelet_insecure_skip_verify":false
}
```

The server returns HTTP 201 with `{"pipeline_id":1}`. `name` and the base `endpoint` are required. The per-signal endpoints and headers are optional. All endpoints must use HTTP or HTTPS; embedded URL credentials and fragments are rejected. Headers contain literal string values. `Basic CHANGE_ME` is a placeholder; demo setup supplies the actual authorization value. The CtrlB JSON ingestion route is not an OTLP endpoint. The overrides above use its native protobuf OTLP routes. Export requests use protobuf encoding without compression.

The initial graph collects new pod logs from every namespace and node, pod, and container metrics every thirty seconds. It excludes its own pod logs using `POD_NAMESPACE`, `POD_NAME`, and `POD_UID`. It enriches telemetry with Kubernetes resource metadata, applies memory limiting and batching, and exports to the supplied telemetry destination. The kubelet uses the pod's service account and CA; certificate verification is enabled unless the caller explicitly opts out for a development cluster.

Collector environment values are `K8S_NODE_NAME`, `K8S_NODE_IP`, `K8S_CLUSTER_NAME`, and `STATE_DIR`, plus the pod values above. `STATE_DIR` defaults to `/var/lib/ctrlb` in the managed runtime. The file log receiver refers to `file_storage`, which the compiler configures at `${env:STATE_DIR}/offsets` with directory creation enabled. Storage remains in the generated config after graph edits while any file log receiver refers to it. The installer must mount `/var/log/pods` read-only and persist the state directory on the node for this installation.

Graph compilation follows directed paths per signal. Multiple receivers can share an ordered processor chain and each chain can export to multiple destinations. Cycles, disconnected components, invalid edge endpoints, and competing processor branches for the same signal return an error. Component IDs and graph input order do not alter the serialized configuration; processing order comes from the edges. Role suffixes do not truncate internal underscores in canonical Collector component types.

`POST /api/frontend/v2/installations` uses the existing user access token.

```json
{"name":"demo","cluster_name":"demo-cluster","pipeline_id":1}
```

The pipeline must already exist and contain a configuration. All collectors in the installation inherit that pipeline. The server returns HTTP 201.

```json
{
  "installation": {
    "id":"<installation-id>",
    "name":"demo",
    "cluster_name":"demo-cluster",
    "pipeline_id":1,
    "created_by":"demo@example.com",
    "created_at":1792540800,
    "desired_revision":"sha256:<hex>",
    "collectors":[],
    "rollout":{"total":0,"pending":0,"applied":0,"failed":0,"disconnected":0}
  },
  "token":"<installation-bearer-token>"
}
```

The token is returned only at creation. Store it in a Kubernetes Secret. The server stores its SHA-256 hash, never its plaintext. The token grants check-in access only to the installation in the request URL. A logged-in user's access or refresh token is not a collector token.

## Installation manifest

`POST /api/frontend/v2/installations/{installation_id}/manifest` uses the existing user access token and the installation token returned at creation.

```json
{
  "token":"<installation-bearer-token>",
  "backend_url":"https://control-plane.example.com",
  "image":"ctrlb-agent:k8s-v1"
}
```

Return the installation's resources as `application/yaml`, with `Cache-Control: no-store` and a download filename of `ctrlb-collector-<first 12 installation ID characters>.yaml`. Resource names use the same prefix and ID suffix, in namespace `observability`. The manifest includes the token in a Kubernetes Secret; the API validates it against the stored hash before rendering. Token values are never placed in URLs.

`backend_url` is the collector-reachable management URL, separate from telemetry ingestion. HTTPS is the default; an explicitly supplied HTTP URL enables `ALLOW_INSECURE_HTTP=true` for local demos. `image` identifies the collector built from this implementation; the local demo builds and loads `ctrlb-agent:k8s-v1` rather than assuming an unpublished registry tag exists.

The manifest can create the shared namespace during installation. Uninstall commands remove only the named DaemonSet, Secret, ServiceAccount, ClusterRole, and ClusterRoleBinding. They preserve the namespace and other installations.

## Collector check-in

`POST /api/agent/v2/installations/{installation_id}/check-in` requires `Authorization: Bearer <installation-bearer-token>`.

```json
{
  "node_name":"worker-1",
  "node_uid":"optional-node-uid",
  "namespace":"observability",
  "pod_name":"ctrlb-collector-abcde",
  "pod_uid":"replacement-pod-uid",
  "pod_ip":"10.244.0.10",
  "version":"0.163.0",
  "platform":"linux/amd64",
  "engine_status":"running",
  "applied_revision":"sha256:<hex>",
  "attempted_revision":"sha256:<hex>",
  "config_error":"",
  "metrics":{
    "cpu_percent":2.5,
    "memory_bytes":104857600,
    "logs_sent_per_second":12.0,
    "metrics_sent_per_second":20.0,
    "spans_sent_per_second":0.0
  }
}
```

`node_name` is required. The stable identity is installation ID plus node name. Repeating a check-in or replacing the pod updates the same collector record and keeps its pipeline assignment. Node UID and pod fields are descriptive metadata. The server takes cluster and pipeline assignment from the installation, never from collector-supplied fields.

`engine_status` is required and is `waiting`, `running`, or `failed`. Version, platform, metadata, revisions, error, and metrics are optional. Optional metrics use the units in their field names. Missing values mean unavailable. Metrics must be finite and nonnegative. CPU percent may exceed 100 for a process using multiple cores.

Omitting `metrics` or sending `null` preserves the previous sample and its timestamp. Sending a metrics object replaces the previous sample; an empty object records a sample with no available measurements. Null fields inside a metrics object are omitted from the stored sample.

`attempted_revision` is the latest revision whose application it attempted, including failed attempts. `applied_revision` is the revision currently running successfully. The client persists the working configuration and its revision. It preserves the failed attempted revision across restarts when its saved working configuration remains usable. A fresh collector without usable saved configuration sends empty revisions and waits for its desired configuration.

The server returns HTTP 200.

```json
{
  "collector_id":12,
  "desired_revision":"sha256:<hex>",
  "config":{"receivers":{},"processors":{},"exporters":{},"service":{}},
  "poll_interval_seconds":10,
  "server_time":1792540800
}
```

`config` is a flexible JSON object containing the full desired Collector configuration. It appears only when the desired revision differs from both `applied_revision` and `attempted_revision`. An unchanged failed revision is not sent repeatedly. Editing the desired configuration produces another revision and allows another attempt. A collector missing its saved working configuration clears its revisions to request the desired configuration again.

Revisions are `sha256:` plus the SHA-256 hex digest of canonical JSON. Object key order and JSON formatting do not change the revision. Array order is meaningful. The server retains only the current desired configuration. Collectors own their last working configuration and rollback.

Invalid or cross-installation tokens return HTTP 401 without collector enrollment. Invalid request fields return HTTP 400. Unknown installations also return HTTP 401 to collector callers. Backend errors return HTTP 500. Clients retry with bounded exponential backoff and jitter; transient errors do not discard their working configuration.

## Installation and pipeline status

These routes use the existing user access token:

- `GET /api/frontend/v2/installations` returns an array of installations.
- `GET /api/frontend/v2/installations/{installation_id}` returns one installation, or HTTP 404.
- `GET /api/frontend/v2/pipelines/{pipeline_id}/installations` returns the pipeline's installations, or HTTP 404 if the pipeline does not exist.

Each installation has the shape in the creation response, without the token. Each collector includes:

```json
{
  "id":12,
  "installation_id":"<installation-id>",
  "node_name":"worker-1",
  "node_uid":"optional-node-uid",
  "namespace":"observability",
  "pod_name":"ctrlb-collector-abcde",
  "pod_uid":"replacement-pod-uid",
  "pod_ip":"10.244.0.10",
  "version":"0.163.0",
  "platform":"linux/amd64",
  "last_seen":1792540800,
  "connection_status":"connected",
  "engine_status":"running",
  "applied_revision":"sha256:<hex>",
  "attempted_revision":"sha256:<hex>",
  "config_error":"",
  "rollout_status":"applied",
  "metrics":{"cpu_percent":2.5,"memory_bytes":104857600},
  "metrics_updated_at":1792540800
}
```

Connectivity and rollout are independent. `connection_status` is `connected` when the latest server-received check-in is no more than 35 seconds old, otherwise `disconnected`. `rollout_status` is `applied` when the engine is running the desired revision; `failed` when the attempted revision equals the desired revision and `config_error` is nonempty or the engine is failed; otherwise `pending`. A disconnected collector still reports its last known rollout state. Rollout totals count these states separately from disconnected count. Server times are Unix seconds.

Kubernetes collectors have `management_mode = pull` in the existing agent inventory. Their desired configuration is never pushed to pod/node addresses, and the backend's VM metrics queue does not scrape them. Legacy VM lifecycle and assignment APIs reject pull collectors. Pipeline deletion fails while an installation references it. Use the documented Kubernetes uninstall command for deployed resources.
