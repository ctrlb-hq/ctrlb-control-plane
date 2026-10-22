# Kubernetes collector management

Kubernetes collectors run as a DaemonSet and connect to a control plane outside the cluster. They collect pod logs and node, pod, and container metrics, apply shared pipeline configuration changes, and report collector health and rollout progress. Use the [installation guide](kubernetes-setup.md) to connect a cluster or the [management API guide](kubernetes-api.md) to automate setup and inspect status.

## Fleet model

A pipeline defines how collectors receive, process, and export telemetry. An installation connects a collector fleet in one named Kubernetes cluster to a shared pipeline. Each collector runs the telemetry engine on one node.

The generated DaemonSet runs one collector per eligible Linux node and tolerates the standard control-plane node taint. Replacing a collector pod preserves that collector's identity and pipeline assignment; a new node joins the installation with the shared configuration. Identity is the installation ID plus node name.

## Collector communication and configuration

Collectors initiate HTTPS check-ins approximately every ten seconds with jitter. They report health, the applied configuration revision, and errors, and receive a changed desired configuration from the backend. Existing collectors keep their last working configuration during a control-plane outage and retry automatically; a new collector without a valid configuration waits for enrollment.

Each collector validates and applies updates independently, confirms startup before reporting a revision as applied, and attempts to restore its previous working configuration if startup fails. Applying an update restarts the embedded Collector and can briefly interrupt collection. Offline collectors catch up after reconnecting. The UI shows pending, applied, or failed state per collector and progress across the installation; partial rollout does not trigger a rollback of successful collectors.

Check-ins use the installation token and preserve hosted API prefixes such as `/control-plane/api`. The backend does not need inbound access to pod or node addresses. The collector's local port `3421` serves health and readiness probes; it does not expose VM lifecycle controls. See the [collector guide](collector/overview.md) for configuration and metrics.

## Collection defaults

- Collect new container stdout/stderr logs from all namespaces, excluding the collector's own logs to prevent feedback loops.
- Collect node, pod, and container metrics every thirty seconds.
- Attach cluster, node, namespace, pod, and container metadata to the corresponding telemetry.
- Save log offsets and the working configuration in node storage isolated by installation ID, so collector pods recreated on the same node can resume collection. Replacing the node loses its local state.
- Configure collection and filtering through the shared pipeline graph. Setup does not provide a namespace selector.

## UI and lifecycle

The UI configures pipelines and shows collector health, configuration errors, and rollout progress. On the Pipelines page, choose **Add new pipeline**, then **Kubernetes**. Setup provides manifest YAML, a browser download, and terminal download commands. Use Kubernetes commands to install, stop, upgrade, or uninstall the DaemonSet. Follow the [installation guide](kubernetes-setup.md) for the setup fields, context-specific commands, and verification.

Setup and the manifest API default to `ghcr.io/ctrlb-hq/ctrlb-otel-collector:latest`. The checked-in image workflows publish short commit tags, so enter a published tag or digest explicitly. See [container images](container-images.md) for build and publication details.

Pipeline deletion removes its configuration, installations, credentials, and collector inventory in one transaction. It does not stop collector pods. Uninstall each installation's named Kubernetes resources before deleting the pipeline; the UI's delete confirmation supplies these commands and preserves the shared `observability` namespace. A pod left running can continue exporting with its cached configuration after its token stops working.

## Validate the collection preset

The backend and collector are separate Go modules. Validate the actual compiled Kubernetes preset against the collector's component factories from the repository root:

```bash
preset_fixture="$(mktemp /tmp/ctrlb-kubernetes-preset.XXXXXX)"
KUBERNETES_PRESET_FIXTURE="$preset_fixture" \
  go -C backend test ./internal/kubernetes \
  -run '^TestKubernetesPresetCollectsBothSignalsAndPersistsOffsets$' -count=1
KUBERNETES_PRESET_FIXTURE="$preset_fixture" \
  go -C agent test ./internal/adapters \
  -run '^TestGeneratedKubernetesPresetCompatibility$' -count=1
rm -- "$preset_fixture"
```

The first command exports the compiled preset with a placeholder credential. The second checks its typed configuration with service-account authentication, then runs a host Collector dry run with kubelet authentication changed to `none`. It does not contact a live cluster or establish that telemetry reaches its destination. Follow [enrollment and telemetry verification](kubernetes-setup.md#verify-enrollment-and-telemetry) to check a deployed installation.

## Telemetry destination

The control-plane management URL and the CtrlB telemetry ingestion endpoint are separate configuration values. Collector check-ins go to the control plane; collected telemetry goes to the telemetry destination.

Use the native OTLP HTTP endpoints with protobuf encoding:

- Logs: `https://ctrlb.explore.ctrlb.ai/engine/api/default/control_plane/_otel/v1/logs`
- Metrics: `https://ctrlb.explore.ctrlb.ai/engine/api/default/v1/metrics`

Both use an `Authorization` header supplied during setup. `Basic CHANGE_ME` is an example placeholder, not a working credential. The JSON log ingestion route is not the destination for OTLP payloads.
