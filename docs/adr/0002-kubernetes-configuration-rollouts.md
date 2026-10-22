---
status: accepted
---

# Kubernetes configuration rolls out per collector

Kubernetes v1 collectors apply configuration revisions independently, allowing available nodes to update while other nodes are offline or unable to apply the change. Each collector validates the candidate before replacing its working configuration and reports the revision as applied only after confirming that the Collector starts. If startup fails, it attempts to restore the previous working configuration and reports the failure. If restoration also fails, the collector reports the recovery error.

Offline collectors retain their existing configuration and catch up when they reconnect. A failed or unavailable collector does not trigger a rollback on collectors that successfully applied the revision. The UI shows pending, applied, or failed state per collector and reports progress across the installation.
