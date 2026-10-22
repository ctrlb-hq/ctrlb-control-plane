# Install collectors in an existing Kubernetes cluster

This guide installs one CtrlB collector per eligible Linux node and connects the installation to a shared pipeline. The control plane runs outside the cluster. See [Kubernetes collector management](kubernetes.md) for collection defaults and runtime behavior, or the [management API guide](kubernetes-api.md) to automate setup.

## Prerequisites

- A running control plane, a user account, and access to its frontend. See [deployment](deployment.md) to host the service.
- `kubectl` configured for the target cluster, with permission to create a Namespace, Secret, ServiceAccount, ClusterRole, ClusterRoleBinding, and DaemonSet.
- Linux nodes with pod logs under `/var/log/pods`. Cluster policy must permit the generated pod's root user and hostPath mounts. The pod drops all capabilities, disables privilege escalation, and uses a read-only root filesystem.
- Capacity for each collector's requests of `100m` CPU and `128Mi` memory. The manifest limits each collector to one CPU and `512Mi` memory.
- Collector pod access to the Kubernetes API, its node's kubelet on HTTPS port `10250`, the control-plane management URL, and the telemetry destination. Nodes also need access to the image registry.
- An OTLP HTTP destination and its required authorization value. Management check-ins and telemetry export use separate URLs and credentials.

The DaemonSet selects Linux nodes and tolerates the standard `node-role.kubernetes.io/control-plane:NoSchedule` taint. Other taints, scheduling restrictions, or node policies can prevent a collector from running. The generated service account can read pods, namespaces, and `nodes/stats`.

Select the target context explicitly. Replace `your-cluster-context` in this example and keep the variables in the same terminal for the later commands:

```bash
kubectl config get-contexts
collector_context='your-cluster-context'
kubectl --context "$collector_context" cluster-info
kubectl --context "$collector_context" get nodes -o wide
kubectl --context "$collector_context" get nodes \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.nodeInfo.operatingSystem}{"/"}{.status.nodeInfo.architecture}{"\n"}{end}'
```

## Choose the URLs and collector image

Use a management URL that collector pods can reach. A browser-local address such as `localhost` points to the pod itself when used by a collector. Managed collectors accept either a control-plane root or an API base ending in `/api`:

| Setting | Example | Purpose |
| --- | --- | --- |
| Control-plane management URL | `https://ctrlb.explore.ctrlb.ai/control-plane/api` | Installation check-ins and desired configuration |
| CtrlB OTLP HTTP base endpoint | `https://ctrlb.explore.ctrlb.ai/engine/api/default` | Base URL for the telemetry exporter |
| Logs OTLP endpoint | `https://ctrlb.explore.ctrlb.ai/engine/api/default/control_plane/_otel/v1/logs` | Full native OTLP log route |
| Metrics OTLP endpoint | `https://ctrlb.explore.ctrlb.ai/engine/api/default/v1/metrics` | Full native OTLP metrics route |

For another control plane, replace the management URL with its public root or API base. For another OTLP destination, change the base endpoint and both signal endpoints. The setup form prepopulates CtrlB's signal endpoints; leaving them unchanged continues to send those signals to CtrlB even if you change only the base endpoint. Clear optional signal overrides when the destination should derive them from its base endpoint.

Use HTTPS for management. Selecting HTTP causes the generated manifest to set `ALLOW_INSECURE_HTTP=true` for local development. URLs must be absolute and cannot contain embedded credentials, query parameters, or fragments. The telemetry exporter sends protobuf without compression. CtrlB's JSON ingestion route does not accept these OTLP payloads.

Choose a collector image from this repository that supports every eligible node's architecture. Setup and the manifest API currently default to `ghcr.io/ctrlb-hq/ctrlb-otel-collector:latest`, but the checked-in publication workflows do not publish `latest`. Enter a published short commit tag or digest explicitly:

```text
ghcr.io/ctrlb-hq/ctrlb-otel-collector:<published-short-commit>
ghcr.io/ctrlb-hq/ctrlb-otel-collector@sha256:<published-digest>
```

The `containers/*` workflow publishes Linux AMD64 images. The collector's `agent/v*` workflow publishes Linux AMD64 and ARM64 images. Confirm that the selected workflow succeeded and that your cluster can pull its package; an arbitrary source commit does not imply a published image. The generated manifest uses `imagePullPolicy: IfNotPresent` and does not include an image pull secret. See [container images](container-images.md) for builds, publication, and package visibility.

## Create the pipeline and installation

1. Sign in to the control plane. On **Pipelines**, select **Add new pipeline**, then **Kubernetes**.
2. Enter **Pipeline / installation name** and **Cluster name**. The cluster name identifies this fleet and becomes the `k8s.cluster.name` resource attribute; it does not select a kubectl context.
3. Enter the **Collector-reachable control-plane URL** and the published **Collector image** selected above.
4. Set **CtrlB OTLP HTTP base endpoint**, **Logs OTLP endpoint**, and **Metrics OTLP endpoint** for the telemetry destination.
5. Enter the destination's literal **CtrlB Authorization header value**, including its authentication scheme, if required. This is separate from your user access token and the installation token. The control plane stores it in the pipeline configuration and distributes that configuration to collectors.
6. Leave **Skip kubelet certificate verification for this installation.** unchecked for a cluster with trusted kubelet certificates. This option supports development clusters such as kind with self-signed kubelet certificates.
7. Select **Create pipeline and installation**.

Setup saves the pipeline first, then creates its installation. If installation creation fails after the pipeline is saved, use **Retry installation creation** in the same setup session to reuse that pipeline.

The next view is **Install collectors in** your cluster. It displays **Manifest YAML**, **curl**, **wget**, and **Download installation manifest**. Use the browser download, copy the YAML into the displayed filename, or run one of the generated download commands on your cluster-management machine. Before running a terminal download command, restrict new files to your user:

```bash
umask 077
```

The manifest and terminal download commands contain the installation token. Keep a private copy before closing setup. The token is returned only at installation creation and cannot be recovered from the later status view. It grants check-in and manifest-download access to that installation; the backend stores its hash. The telemetry authorization value is stored in the shared pipeline configuration rather than this installation Secret.

You can correct the management URL or image on this page. Setup regenerates the YAML and download commands when either changes; save the current version. If generation fails, use **Retry manifest generation**. The **View/edit pipeline configuration** button becomes available after a successful browser download or use of **Copy** for **Manifest YAML**, **curl**, or **wget**. Selecting text manually does not enable it.

## Apply the manifest

Use the generated filename and resource name shown in setup. The resource name is `ctrlb-collector-` followed by the first twelve characters of the installation ID. Replace the placeholder below with that name, and run from the directory containing the manifest:

```bash
collector_resource='ctrlb-collector-<first-12-installation-id-characters>'
collector_manifest="${collector_resource}.yaml"
chmod 600 -- "$collector_manifest"

kubectl --context "$collector_context" apply -f "$collector_manifest"
kubectl --context "$collector_context" -n observability rollout status \
  "daemonset/$collector_resource" --timeout=180s
kubectl --context "$collector_context" -n observability get daemonset "$collector_resource"
kubectl --context "$collector_context" -n observability get pods \
  -l "app.kubernetes.io/instance=$collector_resource" -o wide
```

The manifest creates the shared `observability` namespace if needed, then the installation-specific Secret, ServiceAccount, ClusterRole, ClusterRoleBinding, and DaemonSet. It mounts the installation token read-only, pod logs read-only, and a separate writable state directory on each node.

The UI's generated commands use kubectl's current context. Add `--context` as above when running them. Keep the resource name and context with the manifest for later diagnosis and uninstall. Copying a command or downloading the YAML does not deploy collectors; nodes appear in the UI after their first check-in.

## Verify enrollment and telemetry

Wait for the DaemonSet's desired and ready counts to match the eligible node count. In the UI, the installation should show the expected nodes with these independent states:

| State | Expected result | Meaning |
| --- | --- | --- |
| Connection | `connected` | The backend received a check-in within the last 35 seconds |
| Engine | `running` | The embedded Collector is running |
| Configuration | `applied` | The running revision matches the installation's desired revision |

Collector status refreshes every ten seconds. Expand a collector's revision details to compare desired, applied, and attempted revisions. CPU, memory, logs per second, and metric points per second describe the Collector process and its exports. Missing measurements appear as unavailable; rate calculations need successive samples.

Confirm telemetry at the destination separately. Produce a new identifiable stdout/stderr log entry from a workload after the collector is ready, then search for it using `k8s.cluster.name`, `k8s.node.name`, `k8s.namespace.name`, and `k8s.pod.name`. New log files start at their end, so the preset does not backfill existing log contents. The collector excludes its own logs.

Allow at least one thirty-second metrics collection interval, plus destination ingestion time, then check node, pod, and container metric series for the expected nodes. Resource names may be transformed by the destination's index. A running engine, an applied revision, or a successful export response alone does not establish that the destination has indexed the data.

## Change configuration and understand recovery

From setup, use **View/edit pipeline configuration**. Later, open the pipeline from the list and use **View/edit configuration**. Enable **Edit Mode**, change the graph or component forms, select **Review**, and then **Save configuration**. A change to the compiled configuration creates a new desired revision shared by every collector in the installation.

Each collector validates and applies the candidate independently. Applying it restarts the embedded Collector and can briefly interrupt collection. If startup fails, the collector attempts to restore its previous working configuration and reports the failed attempted revision. Successful collectors continue running; there is no fleet-wide automatic rollback.

An unchanged rejected revision is not repeatedly retried. Correct the configuration and save a compiled configuration that produces a different revision. A retained error from an older failed attempt can appear under **Previous configuration error** once the desired revision is running and applied.

The pod serves health and readiness probes on port `3421`. Readiness checks whether the Collector is running, so a collector using its cached working configuration remains ready during a control-plane outage. A reachable control plane marks a collector disconnected after the check-in timeout while preserving its last recorded engine and rollout state. If the backend itself is unavailable, the UI retains its most recent result and reports the refresh error.

Working configuration and log offsets persist on the node at `/var/lib/ctrlb/installations/<installation-id>`, mounted as `/var/lib/ctrlb` in the pod. Pod replacement on that node preserves the cache and collector identity. Replacing the node loses its local state. Cached configurations can contain exporter credentials, so protect that directory along with the manifest and backend database.

## Diagnose an installation

List the installation's pods, then select a pod from that output:

```bash
kubectl --context "$collector_context" -n observability get pods \
  -l "app.kubernetes.io/instance=$collector_resource" -o wide
collector_pod='pod-name-from-the-output'
kubectl --context "$collector_context" -n observability describe pod "$collector_pod"
kubectl --context "$collector_context" -n observability logs "$collector_pod" --tail=100
kubectl --context "$collector_context" -n observability get events --sort-by=.lastTimestamp
```

| Symptom | Check |
| --- | --- |
| `ImagePullBackOff` or `ErrImagePull` | Published image tag, package access, registry credentials, and CPU architecture |
| Collector remains `Pending` | Node taints, policy rejection of root or hostPath, and available CPU/memory |
| Pod runs but never becomes ready | Collector logs, management URL, installation token, and initial configuration errors |
| UI shows no collectors | Whether the manifest was applied to the intended context and pods can reach the management URL |
| UI shows `disconnected` | Pod status, outbound management connectivity, and whether backend records were deleted |
| Rollout is `failed` while the engine runs | The collector may have retained its previous working revision; read its current configuration error |
| Logs are missing | Generate new logs, check `/var/log/pods` access, and confirm the configured OTLP logs endpoint |
| Node/pod/container metrics are missing | Kubelet access on port `10250`, service-account authorization, kubelet certificate trust, and metrics destination |

If the container restarted, add `--previous` to the logs command to inspect its preceding run. See [troubleshooting](troubleshooting.md) and the [management API guide](kubernetes-api.md) for status and error details.

## Uninstall this installation

Run installation-specific deletion commands in the selected context before deleting the pipeline:

```bash
kubectl --context "$collector_context" -n observability delete \
  daemonset,secret,serviceaccount "$collector_resource" --ignore-not-found
kubectl --context "$collector_context" delete \
  clusterrole,clusterrolebinding "$collector_resource" --ignore-not-found
kubectl --context "$collector_context" get namespace observability
kubectl --context "$collector_context" -n observability get pods \
  -l "app.kubernetes.io/instance=$collector_resource"
```

These commands preserve the shared namespace, other installations, and the node's saved state. Avoid `kubectl delete -f` on the generated manifest because it includes the shared Namespace. The same resource-specific commands are available under **Uninstall this installation**, the pipeline's **Kubernetes rollout and uninstall commands**, and the pipeline delete confirmation.

After cluster cleanup, delete the pipeline in the UI if it is no longer needed. Deletion removes its installation credentials and collector inventory, but it does not remove Kubernetes resources. Collectors left running can continue exporting with cached configuration after their credentials stop working. Deleting the pipeline does not erase saved node configuration or log offsets.
