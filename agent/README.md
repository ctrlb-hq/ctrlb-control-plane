# CtrlB collector

The collector embeds OpenTelemetry Collector and applies pipeline configurations from the CtrlB control plane. The same binary supports VM collectors managed through inbound HTTP requests and Kubernetes collectors managed through outbound check-ins.

`INSTALLATION_ID` selects the runtime. A nonempty value starts Kubernetes managed mode; otherwise the collector starts VM mode. Configuration comes from environment variables and an optional `.env` in the working directory.

## Build and run a VM collector

Use Go 1.26. From the repository root, build the binary and copy the bootstrap configuration to a writable location:

```bash
CGO_ENABLED=0 go -C agent build -o /tmp/ctrlb-collector ./cmd/ctrlb_collector
cp agent/internal/config/config.yaml /tmp/ctrlb-collector-config.yaml

BACKEND_URL=http://localhost:8096 \
PIPELINE_NAME=test-pipeline \
STARTED_BY=dev@example.com \
AGENT_CONFIG_PATH=/tmp/ctrlb-collector-config.yaml \
/tmp/ctrlb-collector
```

The collector starts with the bootstrap file, registers with the backend, and saves the returned configuration to `AGENT_CONFIG_PATH`. A nonempty `PIPELINE_NAME` asks the backend to create a new default pipeline and attach this collector. If you omit that variable, use the authenticated [pipeline attachment API](../docs/backend/api-reference.md#pipelines) to assign the enrolled collector. Its file watcher applies later configuration changes. The working directory must also be writable because VM mode writes `app.log` there.

`BACKEND_URL` is required. `AGENT_CONFIG_PATH` defaults to `./config.yaml`, which must exist before startup. `PIPELINE_NAME` and `STARTED_BY` default to empty strings in the binary; the VM installer requires both. The binary does not accept `--config` or `--backend` flags.

The backend must reach VM collectors on port `3421` for configuration and lifecycle requests, and on port `8888` for the generated Prometheus metrics endpoint. See the [collector guide](../docs/collector/overview.md) for the environment variables, APIs, and networking details.

The VM installation script expects release assets named `ctrlb-collector-<os>-<arch>`. The checked-in release workflows currently produce different names. Use a source build or a [container image](../docs/container-images.md) until the selected release provides compatible assets.

## Install in Kubernetes

Choose Kubernetes during pipeline setup, enter the collector-reachable control-plane URL and telemetry destination, and download or copy the generated installation manifest. Apply it to the intended cluster. The manifest runs one collector per eligible Linux node and provides its installation token, Kubernetes metadata, probes, and persistent node storage.

Setup and the manifest API currently default to `ghcr.io/ctrlb-hq/ctrlb-otel-collector:latest`, but the checked-in image workflows publish short commit tags. Enter a published tag or digest explicitly.

- [Kubernetes setup](../docs/kubernetes-setup.md)
- [Kubernetes behavior](../docs/kubernetes.md)
- [Management API and manifest](../docs/kubernetes-api.md)
- [Image builds and publication](../docs/container-images.md)

## Development

Run the collector tests from its Go module:

```bash
go -C agent test ./...
```

The component factories in [otel_factory.go](internal/adapters/otel_factory.go) define which OpenTelemetry receivers, processors, exporters, and extensions this binary can run. It is not the full upstream contrib distribution.

See [LICENSE](LICENSE) for the collector license.
