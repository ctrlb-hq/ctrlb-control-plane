## Input Plugin Mapping

This document provides a reference for mapping Fluent Bit input plugins to their OpenTelemetry (OTel) Collector exporter equivalents. It is intended to help users migrating pipelines from Fluent Bit to the OTel Collector.

| Fluent Bit Input | OTel Collector Equivalent | Status / Notes |
| :--- | :--- | :--- |
| [CollectD](https://docs.fluentbit.io/manual/pipeline/inputs/collectd) | [CollectD Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/collectdreceiver) | 🚧 **Beta** |
| [CPU](https://docs.fluentbit.io/manual/pipeline/inputs/cpu-metrics) | [Host Metrics Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) | 🚧 **Beta** |
| [Disk](https://docs.fluentbit.io/manual/pipeline/inputs/disk-io-metrics) | [Host Metrics Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) | 🚧 **Beta** |
| [Docker log](https://docs.fluentbit.io/manual/pipeline/inputs/docker-metrics) | [Docker Stats Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/dockerstatsreceiver) | 🧪 **Alpha** |
| [Docker events](https://docs.fluentbit.io/manual/pipeline/inputs/docker-events) | [Docker Stats Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/dockerstatsreceiver) | 🧪 **Alpha** |
| [Dummy](https://docs.fluentbit.io/manual/pipeline/inputs/dummy) | No direct support | Workaround: `hostmetrics` or `filelog` can be used |
| [Elasticsearch](https://docs.fluentbit.io/manual/pipeline/inputs/elasticsearch) | [Elasticsearch Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/elasticsearchreceiver) | 🚧 **Beta** |
| [Exec](https://docs.fluentbit.io/manual/pipeline/inputs/exec) | No direct support | |
| [Forward](https://docs.fluentbit.io/manual/pipeline/inputs/forward) | [FluentForward Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/fluentforwardreceiver) | 🚧 **Beta** |
| [Head](https://docs.fluentbit.io/manual/pipeline/inputs/head) | [Filelog Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) | 🚧 **Beta** |
| [Http](https://docs.fluentbit.io/manual/pipeline/inputs/http) | [OTLP Receiver](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/otlpreceiver) | ✅ **Stable** |
| [Health](https://docs.fluentbit.io/manual/pipeline/inputs/health) | No direct support | |
| [Kafka](https://docs.fluentbit.io/manual/pipeline/inputs/kafka) | [Kafka Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kafkareceiver) | 🚧 **Beta** |
| [Kernel logs](https://docs.fluentbit.io/manual/pipeline/inputs/kernel-logs) | No direct support | |
| [Kubernetes](https://docs.fluentbit.io/manual/pipeline/inputs/kubernetes-events) | No direct support | Workaround: `kubeletstats` receiver maybe used |
| [Memory metrics](https://docs.fluentbit.io/manual/pipeline/inputs/memory-metrics) | [Host Metrics Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) | 🚧 **Beta** |
| [MQTT](https://docs.fluentbit.io/manual/pipeline/inputs/mqtt) | No direct support | |
| [Network I/O](https://docs.fluentbit.io/manual/pipeline/inputs/network-io-metrics) | [Host Metrics Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) | 🚧 **Beta** |
| [Nginx](https://docs.fluentbit.io/manual/pipeline/inputs/nginx) | [Nginx Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/nginxreceiver) | 🚧 **Beta** |
| [Node exporter metrics](https://docs.fluentbit.io/manual/pipeline/inputs/node-exporter-metrics) | No direct support | Workaround: `prometheus` receiver maybe used |
| [Podman](https://docs.fluentbit.io/manual/pipeline/inputs/podman-metrics) | [Podman Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/podmanreceiver) | 🧪 **Alpha** |
| [Process log based metrics](https://docs.fluentbit.io/manual/pipeline/inputs/process) | No direct support | Workaround: `filelog` receiver may be used |
| [Process Exporter Metrics](https://docs.fluentbit.io/manual/pipeline/inputs/process-exporter-metrics) | No direct support | |
| [Prometheus Scrape Metrics](https://docs.fluentbit.io/manual/pipeline/inputs/prometheus-scrape-metrics) | [Prometheus Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver) | 🚧 **Beta** |
| [Prometheus Remote Write](https://docs.fluentbit.io/manual/pipeline/inputs/prometheus-remote-write) | [Prometheus Remote Write Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusremotewritereceiver) | 🛠 **Dev** |
| [Random](https://docs.fluentbit.io/manual/pipeline/inputs/random) | No direct support | |
| [Serial Interface](https://docs.fluentbit.io/manual/pipeline/inputs/serial-interface) | No direct support | |
| [Splunk](https://docs.fluentbit.io/manual/pipeline/inputs/splunk) | [Splunk HEC Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/splunkhecreceiver) | 🚧 **Beta** |
| [Standard Input](https://docs.fluentbit.io/manual/pipeline/inputs/standard-input) | No direct support | |
| [StatsD](https://docs.fluentbit.io/manual/pipeline/inputs/statsd) | [StatsD Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/statsdreceiver) | 🚧 **Beta** |
| [Syslog](https://docs.fluentbit.io/manual/pipeline/inputs/syslog) | [Syslog Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/syslogreceiver) | 🧪 **Alpha** |
| [SystemD](https://docs.fluentbit.io/manual/pipeline/inputs/systemd) | [SystemD Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/systemdreceiver) | 🛠 **Dev** |
| [Tail](https://docs.fluentbit.io/manual/pipeline/inputs/tail) | No direct support | Workaround: `filelog` receiver maybe used |
| [TCP](https://docs.fluentbit.io/manual/pipeline/inputs/tcp) | [TCP Logs Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/tcplogreceiver) | 🧪 **Alpha** |
| [Thermal](https://docs.fluentbit.io/manual/pipeline/inputs/thermal) | No direct support | |
| [UDP](https://docs.fluentbit.io/manual/pipeline/inputs/udp) | [UDP Logs Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/udplogreceiver) | 🧪 **Alpha** |
| [Windows Event Log](https://docs.fluentbit.io/manual/pipeline/inputs/windows-event-log) | [Windows Event Log Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/windowseventlogreceiver) | 🧪 **Alpha** |
| [Windows Exporter Metrics](https://docs.fluentbit.io/manual/pipeline/inputs/windows-exporter-metrics) | No direct support | Workaround: `prometheus` receiver may be used |
