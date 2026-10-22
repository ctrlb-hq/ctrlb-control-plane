# Backend and collector images

Build both images from the repository root:

```bash
docker build -t ctrlb-control-plane-backend:local backend
docker build -t ctrlb-agent:local agent
```

The collector build includes the OpenTelemetry components used by both VM and Kubernetes collectors. It uses a BuildKit compilation cache and defaults to two concurrent Go builds with a 4 GiB Go memory target. On a smaller build host, pass `--build-arg GO_BUILD_PARALLELISM=1 --build-arg GO_MEMORY_LIMIT=2GiB`.

## Publish to GHCR

The container publication workflow builds Linux AMD64 images and pushes:

- `ghcr.io/ctrlb-hq/ctrlb-control-plane-backend:<tag>`
- `ghcr.io/ctrlb-hq/ctrlb-agent:<tag>`

Commit the source and Docker changes, then push a tag beginning with `containers/`:

```bash
image_tag="hosting-$(git rev-parse --short HEAD)"
git tag "containers/${image_tag}"
git push origin "refs/tags/containers/${image_tag}"
```

The workflow uses its repository `GITHUB_TOKEN` with `packages: write`, so a separate registry secret is not required. It also publishes a `sha-<full commit>` image tag and records the immutable image digest in the Actions summary. It does not update `latest`. A manual workflow dispatch is available after the workflow is present on the default branch.

GitHub makes new GHCR packages private by default, even when the linked repository is public. A package administrator must open each package's settings and change its visibility to **Public** once. Publishing subsequent versions preserves that setting. See [GitHub's package visibility documentation](https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility#configuring-visibility-of-packages-for-an-organization).

Verify the images from a machine with no GHCR credentials:

```bash
docker pull "ghcr.io/ctrlb-hq/ctrlb-control-plane-backend:${image_tag}"
docker pull "ghcr.io/ctrlb-hq/ctrlb-agent:${image_tag}"
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

In Kubernetes setup, set **Collector image** to `ghcr.io/ctrlb-hq/ctrlb-agent:<tag>` and leave the control-plane URL set to `https://ctrlb.explore.ctrlb.ai/control-plane/api`. Copy or download the generated manifest and apply it to the intended cluster. Public images do not require an image pull secret.

Kubernetes collectors use persistent node storage and the installation Secret configured in that manifest. The same image also supports legacy VM registration when no installation ID is supplied. See the [Kubernetes API contract](kubernetes-v1-api.md) for check-in and manifest behavior.
