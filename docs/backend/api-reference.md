# Backend API reference

Routes below match [the backend router](../../backend/internal/api/router.go). Paths in each table are relative to its base path. Send JSON request bodies with `Content-Type: application/json` unless stated otherwise.

## Authentication

The `/api/frontend/v2` routes require an access JWT in `Authorization: Bearer <access_token>`. Access tokens expire after 15 minutes; refresh tokens expire after 30 days. Register or log in to receive both tokens, then use `/api/auth/v1/refresh` to renew the access token. Refresh keeps the existing refresh token.

The VM agent API has no authentication middleware. Kubernetes check-ins authenticate with the installation token. The two manifest routes have different requirements, listed below.

### User authentication

Base path: `/api/auth/v1`.

| Method | Path | Request and result |
| --- | --- | --- |
| POST | `/register` | Requires nonempty `name`, `email`, and `password`. Creates a user and returns user details and tokens. |
| POST | `/login` | Requires nonempty `email` and `password`. Returns user details and tokens. |
| POST | `/refresh` | Requires `refresh_token`. Returns `access_token` and the same `refresh_token`. |

Registration and login return `200` on success. The service stores newly registered users with role `user`, regardless of a supplied `role`. Invalid login credentials return `401`. The refresh handler currently returns `500` for invalid or expired refresh tokens, so clients must inspect its response body as well as its status.

```bash
curl -sS http://localhost:8096/api/auth/v1/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Local developer","email":"dev@example.com","password":"replace-with-your-password"}'

curl -sS http://localhost:8096/api/auth/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"dev@example.com","password":"replace-with-your-password"}'

curl -sS http://localhost:8096/api/auth/v1/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'
```

## VM agent API

Base path: `/api/agent/v1`. The collector calls these routes.

| Method | Path | Description |
| --- | --- | --- |
| POST | `/agents` | Register a VM collector and return its `id` and `config`. |
| POST | `/agents/{id}/config-changed` | Trigger a configuration sync from the collector's pipeline. |

Registration requires nonempty `hostname`, `platform`, and `version`. Optional fields are `ip`, `type`, `pipeline_name`, and `started_by`. `type` defaults to `OTEL`. The backend generates the agent name and registration timestamp. If `pipeline_name` is nonempty, it creates a new default pipeline for the agent, even if a pipeline with that name already exists.

```json
{
  "hostname": "collector-01",
  "platform": "linux",
  "version": "0.0.1",
  "ip": "192.0.2.10",
  "type": "OTEL",
  "pipeline_name": "example-pipeline",
  "started_by": "dev@example.com"
}
```

The backend sends VM control and configuration requests to the registered hostname, falling back to the IP, on port `3421`. It scrapes metrics on port `8888`. See [collector operation](../collector/overview.md) for the collector's own endpoints.

## Frontend API

Base path: `/api/frontend/v2`. Every route requires an access JWT.

### Agents

| Method | Path | Description |
| --- | --- | --- |
| GET | `/agents` | List collector inventory and summary metrics. |
| GET | `/agents/{id}` | Get collector details and labels. |
| DELETE | `/agents/{id}` | Delete VM collector inventory. Request shutdown first unless its recorded status is `disconnected`. |
| POST | `/agents/{id}/start` | Start a VM collector's telemetry engine and monitoring. |
| POST | `/agents/{id}/stop` | Stop a VM collector's telemetry engine and monitoring. |
| POST | `/agents/{id}/restart-monitoring` | Add a VM collector back to the monitoring queue. |
| GET | `/agents/{id}/healthmetrics` | Get CPU and memory time series. |
| GET | `/agents/{id}/ratemetrics` | Get stored VM telemetry metric time series. |
| POST | `/agents/{id}/labels` | Add or update labels from a nonempty JSON object of string values. |
| GET | `/unassigned-agents` | List collectors without a pipeline, excluding disconnected collectors. |
| GET | `/latest-agent?since={unix_seconds}` | Get the latest VM registration after `since` that has a pipeline. Returns `null` if none exists. |

For example, labels use `{"environment":"staging","team":"platform"}`. Metric results contain `metric_name` and `data_points`, where each point has a Unix `timestamp` and numeric `value`.

VM monitoring currently stores the scraped exporter counters directly. Fields whose names contain `rate` are not calculated per-second rates. Health fields also depend on the available source metrics: `cpu_utilization` uses `system_cpu_utilization`, falling back to cumulative process CPU seconds when that gauge is missing or zero; `memory_utilization` uses `system_memory_utilization`, falling back to process RSS bytes. These fields do not have consistent percentage units. Kubernetes installation metrics use a separate sampler and explicit units described in the [Kubernetes API contract](../kubernetes-api.md).

Deleting a collector recorded as `disconnected` skips its shutdown request. A successful deletion therefore does not confirm that its process stopped. Stop and disable its service on the host when retiring it. For other recorded states, a failed shutdown request prevents inventory deletion.

Kubernetes collectors use pull management internally. VM start, stop, shutdown, and monitoring operations do not apply to them. Use the installation routes below for their state and rollout status.

### Pipelines

| Method | Path | Description |
| --- | --- | --- |
| GET | `/pipelines` | List pipelines and summary metrics. |
| POST | `/pipelines` | Create a pipeline from a graph and optional VM agent IDs. |
| GET | `/pipelines/{id}` | Get pipeline metadata. |
| DELETE | `/pipelines/{id}` | Delete a pipeline and its graph. Detach VM agents and remove linked Kubernetes installations and enrolled collectors. |
| GET | `/pipelines-overview/{id}` | Get pipeline metadata with a representative agent's details and configuration. |
| GET | `/pipelines/{id}/graph` | Get the pipeline graph. |
| POST | `/pipelines/{id}/graph` | Save the graph and compiled configuration, then push it to attached VM collectors. Kubernetes collectors receive it at check-in. |
| GET | `/pipelines/{id}/agents` | List collectors attached to the pipeline. |
| POST | `/pipelines/{id}/agents/{agent_id}` | Attach a VM collector and send the pipeline configuration to it. |
| DELETE | `/pipelines/{id}/agents/{agent_id}` | Detach a VM collector. |

Creating a pipeline requires nonempty `name` and `created_by` and a `pipeline_graph` object with `nodes` and `edges` arrays. The response is `200` with `{"message":"Pipeline created successfully","id":"<id>"}`. For graph updates, send the graph object directly, without the `pipeline_graph` wrapper. A graph update can return an error after saving if delivery to an attached VM collector fails. Inspect the saved graph and collector state before retrying.

This example creates a logs pipeline with an OTLP HTTP receiver and debug exporter:

```json
{
  "name": "example-logs",
  "created_by": "dev@example.com",
  "agent_ids": [],
  "pipeline_graph": {
    "nodes": [
      {
        "component_id": 1,
        "name": "OTLP input",
        "component_role": "receiver",
        "component_name": "otlp_receiver",
        "config": {"protocols": {"http": {"endpoint": "0.0.0.0:4318"}}},
        "supported_signals": ["logs"]
      },
      {
        "component_id": 2,
        "name": "Debug output",
        "component_role": "exporter",
        "component_name": "debug_exporter",
        "config": {"verbosity": "basic"},
        "supported_signals": ["logs"]
      }
    ],
    "edges": [{"source": "1", "target": "2"}]
  }
}
```

Edges are directed. Their `source` and `target` are string forms of node `component_id` values. IDs must be unique positive integers. The compiler preserves processor order, supports receiver fan-in and exporter fan-out, and rejects cycles and ambiguous processor branches. It emits compatible signal paths and requires every component to belong to at least one supported receiver-to-exporter path.

Deleting a pipeline also invalidates its installation tokens through removal of the installation records. It does not delete DaemonSets in a Kubernetes cluster. Remove those resources separately.

### Components

| Method | Path | Description |
| --- | --- | --- |
| GET | `/component` | List components. Optional `type` is `receiver`, `processor`, or `exporter`. |
| GET | `/component/schema/{name}` | Get the component's JSON schema. |
| GET | `/component/ui-schema/{name}` | Get its form UI schema. |

Use the catalog's `name`, such as `otlp_receiver`, for schema lookups and graph `component_name` values. The backend loads these schemas from its embedded assets at startup.

### Kubernetes installations

| Method | Path | Description |
| --- | --- | --- |
| POST | `/kubernetes-pipelines` | Create an editable Kubernetes logs and metrics preset. Returns `pipeline_id`. |
| POST | `/installations` | Create an installation for an existing pipeline. Returns the installation and its token once. |
| GET | `/installations` | List installations, collectors, and rollout totals. |
| GET | `/installations/{installation_id}` | Get installation, desired revision, and per-collector rollout state. |
| GET | `/pipelines/{id}/installations` | List installations linked to a pipeline. |
| POST | `/installations/{installation_id}/manifest` | Download a manifest. Requires the installation token in the JSON body in addition to the access JWT. |

Kubernetes preset creation takes `name`, `endpoint`, optional `logs_endpoint`, `metrics_endpoint`, `headers`, and `kubelet_insecure_skip_verify`. Installation creation takes `name`, `cluster_name`, and a positive `pipeline_id`.

Both creation routes return `201` on success. Installation reads, lists, check-ins, and manifest downloads return `200`.

Manifest creation takes `token`, `backend_url`, and optional `image`. It returns `application/yaml` as an attachment. Set `image` to a published commit tag or digest; the server's fallback is `:latest`, which the image publication workflow does not publish.

Follow the [Kubernetes setup guide](../kubernetes-setup.md) for deployment steps. See the [Kubernetes API contract](../kubernetes-api.md) for complete request and response examples, token handling, revision acknowledgement, metrics units, and manifest generation.

## Managed collector API

Base path: `/api/agent/v2`.

| Method | Path | Authentication and result |
| --- | --- | --- |
| POST | `/installations/{installation_id}/check-in` | Installation token in `Authorization: Bearer <installation_token>`. Enrolls or updates the collector and returns desired revision, optional config, and poll interval. |
| POST | `/installations/{installation_id}/manifest` | Installation token in JSON `token`. Returns the same manifest as the frontend route, without requiring a user JWT. |

Installation tokens belong to one installation. They cannot authenticate frontend requests or check-ins for another installation. These APIs set `Cache-Control: no-store`.

## Responses and errors

Most handlers return JSON. Error helpers usually produce `{"error":"<message>"}`. Authentication handlers also use objects with `error` and `message`, and some malformed-body responses are plain text.

Existing VM and frontend handlers do not consistently use HTTP status codes for missing records. Several return `200` with an `error` field, so inspect the response body as well as the status. Kubernetes handlers use `400` for invalid requests, `401` for invalid installation tokens, `404` for missing installations or pipelines, and `500` for unexpected failures. Kubernetes JSON requests are limited to 1 MiB.

The server handles CORS preflight requests with `204`. It has no dedicated health route or version discovery route.
