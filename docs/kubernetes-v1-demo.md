# Local Kubernetes demo

This workflow creates an isolated two-node kind cluster and runs the control plane on its Docker host. It checks pod logs, node, pod, and container metrics, configuration rollout, failed-update recovery, pod replacement, and startup from cached configuration during a control-plane outage. To install into a cluster you already manage, use the [existing-cluster installation guide](kubernetes-setup.md).

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

The base endpoint is `https://ctrlb.explore.ctrlb.ai/engine/api/default`. Override it with `CTRLB_OTLP_ENDPOINT`; override the signal-specific URLs with `CTRLB_LOGS_ENDPOINT` and `CTRLB_METRICS_ENDPOINT`. Changing only the base endpoint leaves the default signal-specific CtrlB URLs in place. The exporter sends protobuf without compression. See the [API contract](kubernetes-v1-api.md) for the endpoint distinction and credential handling.

In real-destination mode, lifecycle checks still run. Confirm `ctrlb-demo-marker` logs from both nodes and node/pod/container metric series in CtrlB. Use `k8s.cluster.name`, `k8s.node.name`, `k8s.namespace.name`, and `k8s.pod.name` to identify this demo. Backend check-ins and successful export responses alone do not establish that CtrlB has indexed the data.

Allow time for CtrlB's ingestion buffers to flush before querying newly created metric streams. The indexed resource fields use names such as `resource_k8s_cluster_name` and `resource_k8s_node_name`.

## Show the UI

Start the frontend against the host backend:

```bash
cd frontend
npm ci --legacy-peer-deps
VITE_API_BASE_URL=http://localhost:18096/api npm run dev -- --host 127.0.0.1
```

Open the Vite URL and sign in using the private demo account in `login.json`, or register another demo account. Open the Kubernetes pipeline to view each node's connectivity, engine state, revisions, rollout status, CPU, and memory. Choose **View/edit configuration** to change the shared pipeline.

The setup flow for an existing cluster is documented in [Kubernetes installation](kubernetes-setup.md). For another installation in this local kind cluster, use its loaded image, `ctrlb-otel-collector:k8s-v1`, and the collector-reachable host address recorded in the private demo state. The browser-local backend URL is not reachable from collector pods.

## Uninstall and inspect

From the repository root, remove the default demo installation and stop its host processes:

```bash
./scripts/kubernetes/demo.sh down
```

If you used `--dir` during setup, pass the same directory to every later command, including cleanup. For the custom and real-destination examples above, use the matching command:

```bash
./scripts/kubernetes/demo.sh --dir /tmp/ctrlb-k8s-demo-local down
./scripts/kubernetes/demo.sh --dir /tmp/ctrlb-k8s-real down
```

Each command operates on the installation and processes recorded in that directory. Omitting `--dir` selects `/tmp/ctrlb-k8s-demo`, or `DEMO_DIR` if that environment variable is set.

The command deletes only the named DaemonSet, Secret, ServiceAccount, ClusterRole, and ClusterRoleBinding, plus the demo log workload. It confirms that `observability` still exists. It keeps the shared namespace, other installations, saved node offsets, and private evidence files. It does not delete the downloaded manifest with `kubectl delete -f`, because that manifest includes the shared namespace.

To remove the isolated kind cluster too:

```bash
./scripts/kubernetes/demo.sh down --delete-cluster
```

For a custom directory, include it here as well, for example `./scripts/kubernetes/demo.sh --dir /tmp/ctrlb-k8s-real down --delete-cluster`.

For resources installed outside this demo, follow the [context-specific uninstall commands](kubernetes-setup.md#uninstall-this-installation).

For diagnosis, use the private kubeconfig and resource name from `state.json`:

```bash
kubectl --kubeconfig /tmp/ctrlb-k8s-demo/kubeconfig \
  --context kind-ctrlb-v1-demo -n observability get pods -o wide
kubectl --kubeconfig /tmp/ctrlb-k8s-demo/kubeconfig \
  --context kind-ctrlb-v1-demo -n observability logs 'daemonset/<resource-name>' --tail=100
```

The log receiver needs root to read node-owned pod log files. The generated pod drops all capabilities, disables privilege escalation, uses a read-only root filesystem, mounts pod logs read-only, and does not use privileged mode or host networking. Its persistent state is isolated under `/var/lib/ctrlb/installations/<installation-id>` on each node. That state survives pod replacement on the same node; replacing the node loses its local state.

If you also want to delete the pipeline, uninstall its Kubernetes resources first. Pipeline deletion removes installation credentials and collector records from the control plane; it does not remove pods or stop a collector using cached configuration. The pipeline delete confirmation provides the same installation-specific uninstall commands. Keep the private manifest or demo state until cleanup is complete.

## Automated validation

Run the local OTLP receiver tests from the repository root:

```bash
python3 -m unittest discover -s scripts/kubernetes -p 'test_*.py'
```

The [Kubernetes validation commands](kubernetes-v1.md#validation) check the generated preset against the collector's component factories. These tests do not deploy a cluster. Run `up` and `verify` as described above to collect current deployment evidence, and inspect the telemetry destination when using real-destination mode.
