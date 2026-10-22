# Control plane frontend

The React 18 and TypeScript application manages VM collectors and Kubernetes collector installations. Vite builds the app, React Flow provides the pipeline editor, and the backend supplies component schemas and collector status.

For the cluster installation workflow, see [Kubernetes setup](../docs/kubernetes-setup.md).
For a walkthrough of the UI, see the [usage guide with screenshots](../docs/usage.md).

## Run locally

Use Node.js 24, as the frontend Docker build does, and npm. The utility tests use Node's built-in TypeScript support.

From the repository root:

```bash
cd frontend
npm ci --legacy-peer-deps
npm run dev
```

Open [http://localhost:3030](http://localhost:3030). Start the backend separately using the [getting started guide](../docs/getting-started.md). Development requests go to `http://localhost:8096/api` by default, so a frontend environment file is optional.

To use a different backend, create `frontend/.env.local`:

```dotenv
VITE_API_BASE_URL=http://localhost:18096/api
```

Restart Vite after changing environment files. `npm ci` needs `--legacy-peer-deps` for the current dependency tree.

## Configuration

| Variable            | Meaning                                                                                   | Default                                                                                             |
| ------------------- | ----------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| `VITE_API_BASE_URL` | API root, including `/api`. Accepts an absolute HTTP URL or a path on the current origin. | `http://localhost:8096/api` in development; the app base path plus `api` in production.             |
| `VITE_API_URL`      | Legacy backend root. The app appends `/api`.                                              | Unset. `VITE_API_BASE_URL` takes precedence.                                                        |
| `VITE_BASE_PATH`    | URL prefix for assets and browser routes. Use a leading and trailing slash.               | `/` in development; `/control-plane/` in the checked-in production environment file and Dockerfile. |

Vite embeds these values when it builds the application. Set them before `npm run build`, or pass the corresponding Docker build arguments. Browser configuration is public; keep credentials out of `VITE_*` variables.

## Checks and builds

Run these commands in `frontend/`:

```bash
npm test
npm run build
npm run lint
npm run format:check
```

`npm test` runs the utility tests in `src/utils/*.test.ts`. These cover API URL resolution, installation commands, manifest requests, rollout summaries, metrics, and YAML output. There is no browser test runner in the package scripts. Check UI changes manually with a running backend.

`npm run build` runs TypeScript checks and writes the production bundle to `dist/`. To inspect it locally:

```bash
npm run preview
```

Open the URL Vite prints, including `/control-plane/` for the default production build. Preview uses the production API configuration, so provide an API URL at build time when testing against a separate local backend:

```bash
VITE_API_BASE_URL=http://localhost:8096/api npm run build
npm run preview
```

## Deploy

From the repository root:

```bash
docker build -t ctrlb-control-plane-frontend:local frontend
docker run --rm -p 8080:80 ctrlb-control-plane-frontend:local
```

The image serves the UI at `http://localhost:8080/control-plane/` and includes the fallback needed to refresh browser routes. Its Nginx configuration serves static files only. Put a reverse proxy in front of it to send `/control-plane/api/*` to the backend's `/api/*`, or build with `--build-arg VITE_API_BASE_URL=http://localhost:8096/api` for local browser access to a separate backend.

To serve the app at the domain root, build with `--build-arg VITE_BASE_PATH=/`. See [frontend architecture and behavior](../docs/frontend/overview.md) and [deployment](../docs/deployment.md) for more detail.
