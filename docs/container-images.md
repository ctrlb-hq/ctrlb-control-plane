# Backend and collector images

Build both images from the repository root:

```bash
docker build -t ctrlb-control-plane-backend:local backend
docker build -t ctrlb-otel-collector:local agent
```

The collector build includes the OpenTelemetry components used by both VM and Kubernetes collectors. It uses a BuildKit compilation cache and defaults to two concurrent Go builds with a 4 GiB Go memory target. On a smaller build host, pass `--build-arg GO_BUILD_PARALLELISM=1 --build-arg GO_MEMORY_LIMIT=2GiB`.

## Publish to GHCR

The [`container-images.yml`](../.github/workflows/container-images.yml) workflow builds Linux AMD64 images and pushes:

- `ghcr.io/ctrlb-hq/ctrlb-control-plane-backend:<short-commit>`
- `ghcr.io/ctrlb-hq/ctrlb-otel-collector:<short-commit>`

Commit the source and Docker changes, then push a tag beginning with `containers/`:

```bash
image_tag="$(git rev-parse --short=7 HEAD)"
git tag "containers/${image_tag}"
git push origin "refs/tags/containers/${image_tag}"
```

The workflow uses its repository `GITHUB_TOKEN` with `packages: write`, so a separate registry secret is not required. Each image gets only its short Git commit hash as the version tag, for example `f4b1855`. The full commit remains in the image's revision label, and the Actions summary records the immutable image digest. A manual workflow dispatch is available after the workflow is present on the default branch.

The separate [`agent-release.yml`](../.github/workflows/agent-release.yml) workflow runs on `agent/v*` tags and publishes the collector image for Linux AMD64 and ARM64, also under a short commit tag. It also attaches Linux and macOS binaries for both architectures to the GitHub release. Neither image workflow publishes `latest` or a semantic-version image tag. Use a tag created by a successful run, or pin a known registry digest. An arbitrary source commit does not imply its image has been published.

GitHub makes new GHCR packages private by default, even when the linked repository is public. A package administrator must open each package's settings and change its visibility to **Public** once. Publishing subsequent versions preserves that setting. See [GitHub's package visibility documentation](https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility#configuring-visibility-of-packages-for-an-organization).

Verify the images from a machine with no GHCR credentials:

```bash
docker pull "ghcr.io/ctrlb-hq/ctrlb-control-plane-backend:${image_tag}"
docker pull "ghcr.io/ctrlb-hq/ctrlb-otel-collector:${image_tag}"
```

## Run the backend

The backend requires `JWT_SECRET`. Set it in the shell that starts the container or in your deployment's secret store. Pass the same value when replacing the container so existing sessions remain valid.

```bash
docker volume create ctrlb-backend-data
docker run -d --name ctrlb-backend \
  --restart unless-stopped \
  -p 127.0.0.1:8096:8096 \
  --env JWT_SECRET \
  --mount type=volume,source=ctrlb-backend-data,target=/data \
  "ghcr.io/ctrlb-hq/ctrlb-control-plane-backend:${image_tag}"
```

The container runs as UID/GID `10001:10001`. Its database and application log live in `/data`; keep that volume across upgrades. A named volume starts with the correct directory ownership. If you use a bind mount instead, make the directory writable by that UID/GID. `PORT` can override the default `8096`; update the port mapping to match.

Use the [hosted deployment guide](deployment.md#frontend-deployment-under-control-plane) to route the public `/control-plane/api/` path to the backend. The backend itself continues to serve `/api/`.

## Install collectors

In [Kubernetes setup](kubernetes-setup.md), set **Collector image** to a published `ghcr.io/ctrlb-hq/ctrlb-otel-collector:<short-commit>` or `ghcr.io/ctrlb-hq/ctrlb-otel-collector@sha256:<digest>`. Setup and the manifest API currently default to `:latest`; override that value because the checked-in workflows do not publish it. For the hosted control plane, use `https://ctrlb.explore.ctrlb.ai/control-plane/api` as the management URL. For another deployment, enter its collector-reachable root or API URL. Copy or download the generated manifest and apply it to the intended cluster. Public images do not require an image pull secret.

The generated DaemonSet selects Linux nodes and uses `imagePullPolicy: IfNotPresent`. Choose an image that supports every eligible node's CPU. Images from `containers/*` runs support AMD64; use a collector release build or your own multi-platform image for mixed AMD64/ARM64 clusters. The [local demo](kubernetes-v1-demo.md) loads a locally built image into its kind nodes and does not need a published registry image.

Kubernetes collectors use persistent node storage and the installation Secret configured in that manifest. The same image supports VM registration when no installation ID is supplied. Its entrypoint is `/usr/local/bin/ctrlb-agent`, working directory `/app`, and bootstrap configuration `/app/config.yaml`. VM mode needs `BACKEND_URL` and a writable config and log directory, plus backend connectivity to ports `3421` and `8888`. See the [collector guide](collector/overview.md) for environment variables and networking, and the [Kubernetes API contract](kubernetes-v1-api.md) for check-in and manifest behavior.
