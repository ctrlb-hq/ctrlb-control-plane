---
status: accepted
---

# Kubernetes collectors use outbound check-ins

The Kubernetes v1 control plane can run outside the cluster, so collectors initiate HTTPS check-ins approximately every ten seconds with a small random offset. Each check-in reports collector health, the applied configuration revision, and errors; the backend returns the desired configuration when its revision changes. This accepts a delay of roughly one check-in interval for configuration changes and status updates, keeping v1 communication simple without requiring inbound access to collector pods or nodes.

Collectors continue using their last working configuration during a control-plane outage and retry automatically. A restarted collector resumes from its saved working configuration; a new collector without a valid configuration waits for enrollment. Persistent connections can be considered later if faster updates become necessary.
