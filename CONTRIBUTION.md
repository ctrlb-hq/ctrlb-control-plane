# Contributing

The repository has separate Go modules for the backend and collector, plus a React frontend. Run checks in the module you change. There is no root Go module.

## Set up

Clone this repository, or your fork if you plan to submit changes from one:

```bash
git clone https://github.com/ctrlb-hq/ctrlb-control-plane.git
cd ctrlb-control-plane
git switch -c docs/describe-your-change
```

Use a short branch name with the appropriate prefix, such as `feat/`, `fix/`, `refactor/`, `sec/`, `test/`, `docs/`, or `chore/`.

Check the toolchain declarations before installing Go. `backend/go.mod` declares Go 1.23.0 with toolchain `go1.24.4`; `agent/go.mod` requires Go 1.26.0. Backend builds use `github.com/mattn/go-sqlite3`, so they require CGO and a C compiler. Use Node.js 24 and npm for the frontend, matching its Docker build.

The [getting started guide](docs/getting-started.md) explains backend startup and collector setup. For the frontend:

```bash
cd frontend
npm ci --legacy-peer-deps
npm run dev
```

The UI runs on port `3030` and defaults to a backend on port `8096`. See the [frontend README](frontend/README.md) for API URL and build configuration.

## Check your changes

For backend changes, run from `backend/`:

```bash
go test ./...
go vet ./...
go build ./cmd/backend
```

For collector changes, run from `agent/`:

```bash
go test ./...
go vet ./...
go build ./cmd/ctrlb_collector
```

Run focused Go checks while developing, for example `go test ./internal/kubernetes` in `backend/` or `go test ./internal/managed` in `agent/`. Run the relevant module suite before submitting. The collector imports many OpenTelemetry components; its first build and test run can take substantial time and memory. `go test -p 1 ./...` limits concurrent package builds on smaller machines.

Format changed Go files with `gofmt`. For concurrency changes, run the affected packages with `go test -race`. See [backend testing](docs/backend/testing.md) for more examples.

For frontend changes, run from `frontend/`:

```bash
npm test
npm run build
npm run lint
npm run format:check
```

The frontend utility tests use Node's test runner. There is no configured browser test suite, so exercise changed screens with a running backend and record the result in the pull request. `npm run build` checks TypeScript before creating `dist/`.

Use the repository's Prettier configuration for frontend files. To format only changed files, pass their paths to `npx prettier --write`; `npm run format:fix` formats the whole frontend. Keep unrelated formatting changes out of the pull request. If a check fails on unchanged code, include the failure and identify it separately from your changes.

For changes to the local OTLP receiver used by the Kubernetes demo, run from the repository root:

```bash
python3 -m unittest discover -s scripts/kubernetes -p 'test_*.py'
```

The [Kubernetes demo guide](docs/kubernetes-v1-demo.md) covers checks with a running cluster.

## Commits and pull requests

Use the existing commit format:

```text
[TYPE]: Short description of change
```

| Type         | Purpose                               |
| ------------ | ------------------------------------- |
| `[FEAT]`     | New behavior                          |
| `[FIX]`      | Bug fix                               |
| `[REFACTOR]` | Refactoring without a behavior change |
| `[SEC]`      | Security change                       |
| `[TEST]`     | Test changes                          |
| `[DOCS]`     | Documentation                         |
| `[CHORE]`    | Build, CI, or dependencies            |
| `[PERF]`     | Performance                           |
| `[STYLE]`    | Formatting or comments                |
| `[WIP]`      | Work not ready for review             |

Before opening a pull request, review the diff for generated binaries, SQLite files, downloaded manifests, and credentials. Kubernetes manifests contain installation credentials; keep them out of commits.

Describe the problem, the resulting behavior, and how you checked it. Include screenshots for visible UI changes and the commands and results for automated checks. Add regression tests when fixing behavior that the existing test setup can exercise. Update the affected setup, API, deployment, or collector documentation when its behavior changes.

Push your branch to your fork or repository and open a pull request against the intended base branch. Keep the change focused enough that a reviewer can check its behavior and documentation together.
