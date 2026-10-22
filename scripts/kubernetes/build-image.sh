#!/usr/bin/env bash
set -euo pipefail
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo_dir=$(cd -- "$script_dir/../.." && pwd)
image=${COLLECTOR_IMAGE:-ctrlb-otel-collector:k8s-v1}
build_dir=$(mktemp -d "${TMPDIR:-/tmp}/ctrlb-image.XXXXXX")
trap 'rm -rf -- "$build_dir"' EXIT
if [[ -n ${COLLECTOR_BINARY:-} ]]; then
  cp -- "$COLLECTOR_BINARY" "$build_dir/collector"
else
  (cd -- "$repo_dir/agent" && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o "$build_dir/collector" ./cmd/ctrlb_collector)
fi
chmod 755 "$build_dir/collector"
ca_bundle=${SSL_CERT_FILE:-/etc/ssl/certs/ca-certificates.crt}
if [[ ! -f $ca_bundle ]]; then
  echo 'Set SSL_CERT_FILE to a PEM CA bundle for the collector image.' >&2
  exit 1
fi
cp -- "$ca_bundle" "$build_dir/ca-certificates.crt"
cp -- "$repo_dir/agent/internal/config/config.yaml" "$build_dir/config.yaml"
docker build --tag "$image" --file "$script_dir/Dockerfile.prebuilt" "$build_dir"
