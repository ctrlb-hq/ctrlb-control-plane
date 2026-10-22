# Kubernetes v1 demo

Run one collector on each of two Linux nodes. The control plane runs on the Docker host, outside the cluster. The demo checks pod logs, node/pod/container metrics, configuration rollout, failed-update recovery, pod replacement, and restart from cached configuration during a control-plane outage.

## Requirements

Use Linux with Docker, Python 3, Go 1.26, [kind v0.33.0](https://github.com/kubernetes-sigs/kind/releases/tag/v0.33.0), and kubectl v1.35.8 on `PATH`. The checked-in kind configuration pins the release's Kubernetes v1.35.8 node image by digest. Reserve space for two node containers, the node image, and Go build caches. Node.js 24 is needed only for the UI.

The demo creates the isolated cluster `ctrlb-v1-demo` and writes a separate kubeconfig. Every kubectl operation uses that file and the explicit `kind-ctrlb-v1-demo` context. It refuses to reuse an existing cluster of that name. Set `DEMO_CLUSTER` to choose another short lowercase name.

The backend listens on host port 18096 and the local OTLP receiver on 14318. Collector pods use the kind Docker network gateway to reach the host. Host firewall rules must allow that network to reach these two ports. This local demo explicitly enables HTTP management and disables kubelet certificate verification for kind. For another cluster, use the collector-reachable HTTPS control-plane URL and keep kubelet certificate verification enabled.

## Run without external credentials

From the repository root:

```bash
./scripts/kubernetes/demo.sh up
./scripts/kubernetes/demo.sh status
./scripts/kubernetes/demo.sh verify
```

`up` builds a static collector, packages it in a small image, loads the image into both kind nodes, and starts the host backend and local OTLP receiver. It registers a demo user, creates one shared pipeline and installation, downloads the installation manifest, and installs collectors plus a log-producing workload. Both collectors must report the shared desired revision as applied.

The default directory is `/tmp/ctrlb-k8s-demo`. Choose a new directory with `--dir /tmp/my-demo` before the command. The directory must not already exist on the first run. It contains the backend database, a private `state.json`, demo login details, the manifest, logs, and verification evidence. The manifest and state contain credentials and are written with owner-only permissions. The receiver records decoded telemetry and resource metadata, but never request Authorization headers.

To reuse locally built binaries and skip compilation:

```bash
BACKEND_BINARY=/tmp/ctrlb-k8s-backend \
COLLECTOR_BINARY=/tmp/ctrlb-k8s-collector \
./scripts/kubernetes/demo.sh --dir /tmp/ctrlb-k8s-demo-local up
```

The collector binary must be a statically linked Linux executable for the Docker host's CPU. `scripts/kubernetes/build-image.sh` also works by itself with `COLLECTOR_BINARY` and optional `COLLECTOR_IMAGE`. It copies only the binary, bootstrap config, and CA bundle into the build context. Set `SSL_CERT_FILE` if the host's PEM CA bundle is not `/etc/ssl/certs/ca-certificates.crt`.

`verify` performs these checks:

1. Both nodes have connected collectors under the same installation and pipeline.
2. The local receiver receives `ctrlb-demo-marker` log records and node, pod, and container metric names, with both node identities in resource attributes.
3. A graph configuration edit produces a new revision and both collectors apply it.
4. An invalid kubelet collection interval fails on both collectors while their previous revision keeps running. Restoring the valid graph clears the failed rollout.
5. Replacing a collector pod preserves its collector ID and pipeline assignment.
6. With the host backend stopped, another replacement pod becomes ready from its saved configuration. Restarting the backend restores check-ins without new collector identities.

Results go to `verification.json`. Local decoded telemetry is in `otlp-evidence.json`. These files support inspection of the actual pod-log bodies and metric names received by the demo destination.

## Send telemetry to CtrlB

Store the literal Authorization header value, including its authentication scheme, in a private file. Do not put the credential in shell command arguments or source files. Set `CTRLB_AUTH_FILE` to that file and use a fresh demo directory and cluster name:

```bash
CTRLB_AUTH_FILE=/path/to/private-authorization.txt \
DEMO_CLUSTER=ctrlb-v1-real \
./scripts/kubernetes/demo.sh --dir /tmp/ctrlb-k8s-real up --real

./scripts/kubernetes/demo.sh --dir /tmp/ctrlb-k8s-real verify
```

The default ingestion destinations are:

| Signal | OTLP HTTP endpoint |
| --- | --- |
| Logs | `https://ctrlb.explore.ctrlb.ai/engine/api/default/control_plane/_otel/v1/logs` |
| Metrics | `https://ctrlb.explore.ctrlb.ai/engine/api/default/v1/metrics` |

The base endpoint is `https://ctrlb.explore.ctrlb.ai/engine/api/default`. Override it with `CTRLB_OTLP_ENDPOINT`; override the signal-specific URLs with `CTRLB_LOGS_ENDPOINT` and `CTRLB_METRICS_ENDPOINT`. The exporter sends protobuf without compression. See the [API contract](kubernetes-v1-api.md) for the endpoint distinction and credential handling.

In real-destination mode, lifecycle checks still run. Confirm `ctrlb-demo-marker` logs from both nodes and node/pod/container metric series in CtrlB. Use `k8s.cluster.name`, `k8s.node.name`, `k8s.namespace.name`, and `k8s.pod.name` to identify this demo. Backend check-ins and successful export responses alone do not establish that CtrlB has indexed the data.

Allow time for CtrlB's ingestion buffers to flush before querying newly created metric streams. The indexed resource fields use names such as `resource_k8s_cluster_name` and `resource_k8s_node_name`.

## Show the UI

Start the frontend against the host backend:

```bash
cd frontend
npm ci --legacy-peer-deps
VITE_API_URL=http://localhost:18096 npm run dev -- --host 127.0.0.1
```

Open the Vite URL and sign in using the private demo account in `login.json`, or register another demo account. Open the Kubernetes pipeline to view each node's connectivity, engine state, revisions, rollout status, CPU, and memory. Choose **View/edit configuration** to change the shared pipeline.

For a new installation, choose Kubernetes in pipeline setup. Enter the collector-reachable management URL separately from the telemetry destination and supply an image available to the cluster. The default `ctrlb-agent:k8s-v1` is the local image built above. Copy the manifest YAML, use a displayed `curl` or `wget` command on your cluster-management machine, or download the file in the browser. Keep a copy during setup because the installation token is returned only once. Apply it with the intended kubeconfig/context and watch the displayed node count; copying a command does not mark deployment successful. See the [deployment guide](deployment.md) for hosting beneath `/control-plane/` with the API beneath `/control-plane/api`.

For a frontend container, pass the browser-reachable backend address at build time:

```bash
docker build -t ctrlb-frontend:k8s-v1 \
  --build-arg VITE_API_URL=http://localhost:18096 frontend
```

The frontend Dockerfile uses Node 24 and `npm ci --legacy-peer-deps`, then serves the build with SPA routing through nginx. The production collector Dockerfile builds with Go 1.26 and `CGO_ENABLED=0`; `docker build -t ctrlb-agent:k8s-v1 agent` builds from source without the prebuilt helper.

## Uninstall and inspect

Remove this installation and stop its host processes:

```bash
./scripts/kubernetes/demo.sh down
```

The command deletes only the named DaemonSet, Secret, ServiceAccount, ClusterRole, and ClusterRoleBinding, plus the demo log workload. It confirms that `observability` still exists. It keeps the shared namespace, other installations, saved node offsets, and private evidence files. It does not delete the downloaded manifest with `kubectl delete -f`, because that manifest includes the shared namespace.

To remove the isolated kind cluster too:

```bash
./scripts/kubernetes/demo.sh down --delete-cluster
```

For a manual installation, substitute the ID prefix and always select the intended context:

```bash
kubectl --context kind-ctrlb-v1-demo -n observability delete \
  daemonset,secret,serviceaccount ctrlb-collector-<first-12-installation-id-characters>
kubectl --context kind-ctrlb-v1-demo delete \
  clusterrole,clusterrolebinding ctrlb-collector-<first-12-installation-id-characters>
```

For diagnosis, use the private kubeconfig and resource name from `state.json`:

```bash
kubectl --kubeconfig /tmp/ctrlb-k8s-demo/kubeconfig \
  --context kind-ctrlb-v1-demo -n observability get pods -o wide
kubectl --kubeconfig /tmp/ctrlb-k8s-demo/kubeconfig \
  --context kind-ctrlb-v1-demo -n observability logs daemonset/<resource-name> --tail=100
```

The log receiver needs root to read node-owned pod log files. The generated pod drops all capabilities, disables privilege escalation, uses a read-only root filesystem, mounts pod logs read-only, and does not use privileged mode or host networking. Its persistent state is isolated under `/var/lib/ctrlb/installations/<installation-id>` on each node. That state survives pod replacement on the same node; replacing the node loses its local state.

## Validation on 2026-10-21

The static Collector binary and prebuilt image were tested in the two-node Kubernetes v1.35.8 kind cluster with the control plane running on its host. Both collectors enrolled under one pipeline, exported logs and node/pod/container metrics with Kubernetes metadata, applied a graph edit, rejected an invalid kubelet duration, preserved their identities through pod replacement, and started from cached configuration while the control plane was stopped. A separate live port-conflict test caused Collector startup to fail on both nodes; both restored their previous working revision.

The corrected CtrlB hostname is `ctrlb.explore.ctrlb.ai`. Native OTLP export succeeded with the supplied private authorization value. Bounded queries confirmed indexed demo-marker logs and node, pod, and container memory metrics from both nodes. Credentials are absent from repository files.

The uninstall command removed the named installation resources while preserving `observability` and a separately named test Secret. The isolated test cluster was then removed.

The full agent and backend test suites, scoped runtime race checks, frontend production build and focused tests, manifest tests, and local OTLP receiver tests passed. Release workflows were updated but were not published or triggered.

An additional frontend end-to-end run used Chromium against the production UI build, a real backend, and a fresh two-node kind cluster. It covered registration, logout/login, Kubernetes setup, manifest download and installation, shared graph/component-form edits, review/save, discarding unsaved edits, invalid-duration rejection with the previous configuration still running, and recovery after restoring a valid interval. The UI also reflected a stopped control plane, a collector replacement started from cache, and an individual node's disconnection and reconnection without changing collector IDs. The final editor and session runs reported no browser console errors, page errors, or failed HTTP requests. Browser testing exposed missing accessible dialog titles, a nested save button, and an unlabeled historical configuration error after recovery; these were corrected and the affected flows were rerun. A current failure stays prominent, while a retained error from an older rejected revision is labeled and collapsed once the desired revision is applied and running. Six frontend tests passed, including revision comparisons that keep current or uncertain errors prominent.

The same run checked the existing VM overview, graph, and installation wizard, and sent a synthetic OTLP log through a local legacy VM collector to its debug exporter. A separate Kubernetes installation was created and downloaded through the frontend using the real CtrlB destination. Both collectors applied its configuration, and bounded CtrlB queries confirmed indexed marker logs and node, pod, and container memory metrics from both nodes. Browser videos and screenshots used the local telemetry destination with a dummy Authorization value; the real-credential setup was not recorded. Customer secret-variable configuration remains deferred.
