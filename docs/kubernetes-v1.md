# Kubernetes collector management

Kubernetes collectors run as a DaemonSet and connect to a control plane outside the cluster. The implementation collects pod logs and node, pod, and container metrics, supports shared pipeline configuration changes, and reports collector health and rollout progress. Use the [installation guide](kubernetes-setup.md) for an existing cluster or the [two-node local demo](kubernetes-v1-demo.md) to exercise these behaviors without external credentials.

## Fleet model

An installation is a collector fleet in one named Kubernetes cluster, assigned to one shared pipeline. The generated DaemonSet runs one collector per eligible Linux node and tolerates the standard control-plane node taint. Replacing a collector pod preserves that collector's identity and pipeline assignment; a new node joins the installation with the shared configuration. Identity is the installation ID plus node name.

See the [glossary](../CONTEXT.md) for the meanings of pipeline, installation, and collector.

## Collector communication and configuration

Collectors initiate HTTPS check-ins approximately every ten seconds with jitter. They report health, the applied configuration revision, and errors, and receive a changed desired configuration from the backend. Existing collectors keep their last working configuration during a control-plane outage and retry automatically; a new collector without a valid configuration waits for enrollment.

Each collector validates and applies updates independently, confirms startup before reporting a revision as applied, and attempts to restore its previous working configuration if startup fails. Applying an update restarts the embedded Collector and can briefly interrupt collection. Offline collectors catch up after reconnecting. The UI shows pending, applied, or failed state per collector and progress across the installation; partial rollout does not trigger a rollback of successful collectors.

Check-ins use the installation token and preserve hosted API prefixes such as `/control-plane/api`. The backend does not need inbound access to pod or node addresses. The collector's local port `3421` serves health and readiness probes; it does not expose VM lifecycle controls. See the [collector guide](collector/overview.md) for configuration and metrics.

See the accepted decisions for [outbound check-ins](adr/0001-kubernetes-collector-check-ins.md) and [configuration rollouts](adr/0002-kubernetes-configuration-rollouts.md).

## Collection defaults

- Collect new container stdout/stderr logs from all namespaces, excluding the collector's own logs to prevent feedback loops.
- Collect node, pod, and container metrics every thirty seconds.
- Attach cluster, node, namespace, pod, and container metadata to the corresponding telemetry.
- Save log offsets and the working configuration in node storage isolated by installation ID, so collector pods recreated on the same node can resume collection. Replacing the node loses its local state.
- Namespace selection and filtering controls are deferred; configuration edits use the shared pipeline graph.

## UI and lifecycle

The UI configures pipelines and shows collector health, configuration errors, and rollout progress. On the Pipelines page, choose **Add new pipeline**, then **Kubernetes**. Setup provides manifest YAML, a browser download, and terminal download commands. Kubernetes commands handle installation and uninstall. Follow the [installation guide](kubernetes-setup.md) for the setup fields, context-specific commands, and verification. Pause/resume controls and collector upgrades from the UI are deferred.

Setup and the manifest API default to `ghcr.io/ctrlb-hq/ctrlb-otel-collector:latest`. The checked-in image workflows publish short commit tags, so enter a published tag or digest explicitly. The local demo supplies `ctrlb-otel-collector:k8s-v1` after loading it into the cluster. See [container images](container-images.md) for build and publication details.

Pipeline deletion removes its configuration, installations, credentials, and collector inventory in one transaction. It does not stop collector pods. Uninstall each installation's named Kubernetes resources before deleting the pipeline; the UI's delete confirmation supplies these commands and preserves the shared `observability` namespace. A pod left running can continue exporting with its cached configuration after its token stops working.

## Demo verification

The demo creates two nodes and a workload that produces identifiable logs and metrics. Its verification command checks both collectors joining one pipeline, exporting telemetry, applying a configuration change, rejecting an invalid update while retaining the working revision, and preserving identity and configuration through pod recreation and a temporary control-plane outage. Its uninstall command removes the installation resources while preserving the shared namespace.

## Validation

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

The first command exports the compiled preset with a placeholder credential. The second checks its typed configuration with service-account authentication, then runs a host Collector dry run with kubelet authentication changed to `none`. It does not contact a live cluster or establish that telemetry reaches its destination. Use the [demo verification](kubernetes-v1-demo.md#run-without-external-credentials) for those checks.

## Telemetry destination

By default, the demo sends pod logs and node, pod, and container metrics to a local OTLP receiver and records decoded evidence. To view exported telemetry in CtrlB, use the demo's `--real` option and supply its ingestion authentication settings.

The control-plane management URL and the CtrlB telemetry ingestion endpoint are separate configuration values. Collector check-ins go to the control plane; collected telemetry goes to the telemetry destination.

Use the native OTLP HTTP endpoints with protobuf encoding:

- Logs: `https://ctrlb.explore.ctrlb.ai/engine/api/default/control_plane/_otel/v1/logs`
- Metrics: `https://ctrlb.explore.ctrlb.ai/engine/api/default/v1/metrics`

Both use an `Authorization` header supplied during setup. `Basic CHANGE_ME` is an example placeholder, not a working credential. The JSON log ingestion route is not the destination for OTLP payloads.
