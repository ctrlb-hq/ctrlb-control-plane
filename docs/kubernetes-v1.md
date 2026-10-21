# Kubernetes v1 scope

Build a usable demo of collectors running in Kubernetes and managed by a control plane outside the cluster. The demo collects pod logs and node/pod metrics, supports pipeline configuration changes, and shows collector health and rollout progress.

## Fleet model

An installation is a collector fleet in one named Kubernetes cluster, assigned to one shared pipeline. One collector runs per node. Replacing a collector pod preserves that collector's identity, pipeline assignment, and history; a new node joins the installation with the shared configuration.

See the [glossary](../CONTEXT.md) for the meanings of pipeline, installation, and collector.

## Collector communication and configuration

Collectors initiate HTTPS check-ins approximately every ten seconds with jitter. They report health, the applied configuration revision, and errors, and receive a changed desired configuration from the backend. Existing collectors keep their last working configuration during a control-plane outage and retry automatically; a new collector without a valid configuration waits for enrollment.

Each collector validates and applies updates independently, confirms startup before reporting a revision as applied, and restores its previous working configuration if startup fails. Offline collectors catch up after reconnecting. The UI shows pending, applied, or failed state per collector and progress across the installation; partial rollout does not trigger a rollback of successful collectors.

See the accepted decisions for [outbound check-ins](adr/0001-kubernetes-collector-check-ins.md) and [configuration rollouts](adr/0002-kubernetes-configuration-rollouts.md).

## Collection defaults

- Collect new container stdout/stderr logs from all namespaces, excluding the collector's own logs to prevent feedback loops.
- Collect node, pod, and container metrics every thirty seconds.
- Attach cluster, node, namespace, pod, and container metadata to the corresponding telemetry.
- Save log offsets so collector pods recreated on the same node can resume collection.
- Omit namespace selection and filtering controls from v1.

## UI and lifecycle

The UI configures pipelines and shows collector health, configuration errors, and rollout progress. Provided Kubernetes commands handle installation and uninstall. Pause/resume controls and collector upgrades from the UI are deferred.

## Demo acceptance target

Use a cluster with at least two eligible nodes and a workload that produces identifiable logs and metrics. Demonstrate both collectors joining one pipeline, exporting telemetry, applying a configuration change, recovering from a failed update, and preserving identity and configuration through pod recreation and a temporary control-plane outage. Finish by uninstalling the deployed resources with the documented command.

## Telemetry destination

Send demo pod logs and node, pod, and container metrics to the CtrlB backend, and view the exported telemetry in CtrlB. Demo setup supplies the ingestion endpoint and required authentication settings.

The control-plane management URL and the CtrlB telemetry ingestion endpoint are separate configuration values. Collector check-ins go to the control plane; collected telemetry goes to the telemetry destination.

Use the native OTLP HTTP endpoints with protobuf encoding:

- Logs: `https://ctrlb.explore.ctrlb.ai/engine/api/default/control_plane/_otel/v1/logs`
- Metrics: `https://ctrlb.explore.ctrlb.ai/engine/api/default/v1/metrics`

Both use an `Authorization` header supplied during setup. `Basic CHANGE_ME` is an example placeholder, not a working credential. The JSON log ingestion route is not the destination for OTLP payloads.
