# Capabilities and future work

This page records what the repository implements and what remains from the project roadmap. Future items have no committed delivery dates. Use [GitHub issues](https://github.com/ctrlb-hq/ctrlb-control-plane/issues) to discuss scope and priorities.

## Implemented

- React graph editor with component schemas, configuration forms, and YAML previews.
- Go backend with embedded component schemas, graph compilation, SQLite metadata, and collector metric storage.
- User registration, login, access JWTs, and refresh tokens.
- VM collector registration, configuration pushes, start/stop management, and backend metrics scraping.
- Kubernetes installation manifests, installation tokens, and one collector per eligible node with pod logs and node, pod, and container metrics.
- Outbound Kubernetes check-ins, stable per-node identity, configuration revision tracking, local validation, rollback after failed updates, and cached operation during control-plane outages.
- Kubernetes installation views with per-collector connection, engine, and rollout status.
- A local kind demo with OTLP capture and a real-destination mode.
- Backend and collector tests, frontend utility tests, and GitHub Actions workflows for release binaries and container publication.
- Backend, frontend, and collector Dockerfiles; backend and collector GHCR publication using short Git commit tags.
- Hosted frontend paths and same-origin API proxying under `/control-plane/`.

See [Kubernetes scope](kubernetes-v1.md) for the current collection and rollout contract.

## Current operational gaps

- The Linux installers expect release asset names that differ from the current release workflows. Source and container deployment paths are documented in [deployment](deployment.md#linux-install-scripts).
- Kubernetes setup suggests a `latest` image tag, while publication workflows produce short commit tags. Operators must choose an available tag or digest.
- VM collectors use a legacy unauthenticated management protocol and require the backend to reach their management and metrics ports.
- SQLite is the only implemented database. The backend uses local persistent storage; replicated storage and high availability need further work.
- Pipeline deletion removes control-plane records but does not uninstall running collector services or Kubernetes resources.

## Further collector work

- Authenticated VM enrollment and management.
- Collector upgrades from the UI, including signature verification.
- Kubernetes namespace selection and collection filtering.
- Kubernetes pause/resume controls and richer installation lifecycle management.
- User-visible configuration history and explicit rollback controls beyond the collector's automatic recovery after failed updates.
- Remote pipeline templates and support for additional collector types.

## Team and deployment work

- Fine-grained roles, organizations/workspaces, and project isolation.
- API keys, SSO/SAML, audit logs, and alerting integrations.
- PostgreSQL metadata storage and a separate time-series store for operational metrics.
- Helm/Terraform deployment packages, backup and recovery workflows, air-gapped installation, and high availability.
- Hosted service operations, team management, and usage/billing controls.
