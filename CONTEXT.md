# Collector management

Agreed terminology for the Kubernetes v1 collector management model.

## Language

**Control plane**:
The system that manages collector installations, desired pipeline configurations, and collector status.
_Avoid_: Telemetry destination

**Telemetry destination**:
The service that receives collected logs and metrics. It is configured separately from the control plane.
_Avoid_: Control plane, backend when the distinction is unclear

**Pipeline**:
A control-plane telemetry configuration describing collection, processing, and export for one or more signals. It is the shared desired configuration assigned to an installation.
_Avoid_: Installation, collector

**Installation**:
A deployed fleet of collectors in one named Kubernetes cluster, assigned to one pipeline that is shared by every member.
_Avoid_: Cluster when referring to the deployed collector fleet

**Collector**:
A stable per-node member of an installation. Its identity and pipeline assignment survive replacement of its pod.
_Avoid_: Pod or agent process when referring to the stable collector identity
