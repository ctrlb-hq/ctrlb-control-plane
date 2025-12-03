# Migration Guide: Fluent Bit Outputs to OpenTelemetry Collector

This document provides a reference for mapping Fluent Bit output plugins to their OpenTelemetry (OTel) Collector exporter equivalents. It is intended to help users migrating pipelines from Fluent Bit to the OTel Collector.

## Output Plugin Mapping

| Fluent Bit Output | OTel Collector Equivalent | Status / Notes |
| :--- | :--- | :--- |
| [Amazon CloudWatch](https://docs.fluentbit.io/manual/pipeline/outputs/cloudwatch) | [AWS CloudWatch Logs Exporter](https://docs.fluentbit.io/manual/pipeline/outputs/cloudwatch) | 🚧 **Beta** |
| [Amazon Kinesis Data Firehose](https://docs.fluentbit.io/manual/pipeline/outputs/firehose) | [Kinesis Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/awskinesisexporter) | 🚧 **Beta** |
| [Amazon Kinesis Data Streams](https://docs.fluentbit.io/manual/pipeline/outputs/kinesis) | [Kinesis Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/awskinesisexporter) | 🚧 **Beta** |
| [Amazon S3](https://docs.fluentbit.io/manual/pipeline/outputs/s3) | [AWS S3 Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/awss3exporter) | 🧪 **Alpha** |
| [Azure Data Explorer](https://docs.fluentbit.io/manual/pipeline/outputs/azure_kusto) | [Azure Data Explorer Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/azuredataexplorerexporter) | 🚧 **Beta** |
| [Azure Blob](https://docs.fluentbit.io/manual/pipeline/outputs/azure_blob) | *None listed* | |
| [Azure Logs Ingestion API](https://docs.fluentbit.io/manual/pipeline/outputs/azure_logs_ingestion) | *None listed* | |
| [Counter](https://docs.fluentbit.io/manual/pipeline/outputs/counter) | No direct support | Workaround: Use Prometheus and OTLP exporters |
| [Datadog](https://docs.fluentbit.io/manual/pipeline/outputs/datadog) | [Datadog Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/datadogexporter) | 🚧 **Beta** |
| [Elasticsearch](https://docs.fluentbit.io/manual/pipeline/outputs/elasticsearch) | [Elasticsearch Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/elasticsearchexporter) | 🚧 **Dev/Beta** |
| [File](https://docs.fluentbit.io/manual/pipeline/outputs/file) | [File Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/fileexporter) | 🧪 **Alpha** |
| [FlowCounter](https://docs.fluentbit.io/manual/pipeline/outputs/flowcounter) | No direct support | |
| [Forward](https://docs.fluentbit.io/manual/pipeline/outputs/forward) | No direct support | |
| [GELF](https://docs.fluentbit.io/manual/pipeline/outputs/gelf) | No direct support | Workaround: Maybe with OTLP |
| [Google Chronicle](https://docs.fluentbit.io/manual/pipeline/outputs/chronicle) | No direct support | Workaround: Maybe with HTTP |
| [Google Cloud BigQuery](https://docs.fluentbit.io/manual/pipeline/outputs/bigquery) | No direct support | Workaround: Maybe with Google Cloud Pubsub Exporter |
| [HTTP](https://docs.fluentbit.io/manual/pipeline/outputs/http) | [OTLP Exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter) | ✅ **Stable** |
| [InfluxDB](https://docs.fluentbit.io/manual/pipeline/outputs/influxdb) | [InfluxDB Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/influxdbexporter) | 🚧 **Beta** |
| [Kafka](https://docs.fluentbit.io/manual/pipeline/outputs/kafka) | [Kafka Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/kafkaexporter) | 🚧 **Beta** |
| [Kafka REST Proxy](https://docs.fluentbit.io/manual/pipeline/outputs/kafka-rest-proxy) | [Kafka Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/kafkaexporter) | 🚧 **Beta** |
| [LogDNA](https://docs.fluentbit.io/manual/pipeline/outputs/logdna) | No direct support | Workaround: Maybe with HTTP exporter |
| [Loki](https://docs.fluentbit.io/manual/pipeline/outputs/loki) | [Loki Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/lokiexporter) | ⚠️ **Deprecated logs** |
| [NATS](https://docs.fluentbit.io/manual/pipeline/outputs/nats) | No direct support | Workaround: Maybe with Prometheus exporter |
| [New Relic](https://docs.fluentbit.io/manual/pipeline/outputs/new-relic) | No direct support | Workaround: Maybe with OTLP |
| [NULL](https://docs.fluentbit.io/manual/pipeline/outputs/null) | No exporter support | Workaround: Use NOP Processor |
| [OpenObserve](https://docs.fluentbit.io/manual/pipeline/outputs/openobserve) | No direct support | Workaround: Maybe with OTLP |
| [Observe](https://docs.fluentbit.io/manual/pipeline/outputs/observe) | No direct support | Workaround: Maybe with OTLP |
| [Oracle Log Analytics](https://docs.fluentbit.io/manual/pipeline/outputs/oci-logging-analytics) | No direct support | Workaround: Maybe with OTLP |
| [OpenSearch](https://docs.fluentbit.io/manual/pipeline/outputs/opensearch) | [OpenSearch Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/opensearchexporter) | 🛠 **Dev** |
| [OpenTelemetry](https://docs.fluentbit.io/manual/pipeline/outputs/opentelemetry) | [OTLP Exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter) | ✅ **Stable** |
| [PostgreSQL](https://docs.fluentbit.io/manual/pipeline/outputs/postgresql) | No direct support | |
| [Prometheus Exporter](https://docs.fluentbit.io/manual/pipeline/outputs/prometheus-exporter) | [Prometheus Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/prometheusexporter) | 🚧 **Beta** |
| [Prometheus Read Write](https://docs.fluentbit.io/manual/pipeline/outputs/prometheus-remote-write) | [Prometheus Remote Write Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/prometheusremotewriteexporter) | |
| [SkyWalking](https://docs.fluentbit.io/manual/pipeline/outputs/skywalking) | No direct support | Workaround: Maybe with OTLP |
| [Slack](https://docs.fluentbit.io/manual/pipeline/outputs/slack) | No direct support | Workaround: Maybe with HTTP |
| [Splunk](https://docs.fluentbit.io/manual/pipeline/outputs/splunk) | [Splunk HEC Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/splunkhecexporter) | 🚧 **Beta** |
| [Stackdriver](https://docs.fluentbit.io/manual/pipeline/outputs/stackdriver) | [Google Cloud Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/googlecloudexporter) | 🚧 **Beta** |
| [Standard Output](https://docs.fluentbit.io/manual/pipeline/outputs/standard-output) | [Debug Exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/debugexporter) | *Logging exporter is deprecated* |
| [Syslog](https://docs.fluentbit.io/manual/pipeline/outputs/syslog) | [Syslog Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/syslogexporter) | 🧪 **Alpha** |
| [TCP & TLS](https://docs.fluentbit.io/manual/pipeline/outputs/tcp-and-tls) | [OTLP Exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter) | |
| [Treasure Data](https://docs.fluentbit.io/manual/pipeline/outputs/treasure-data) | No direct support | Workaround: Maybe with OTLP |
| [Vivo Exporter](https://docs.fluentbit.io/manual/pipeline/outputs/vivo-exporter) | No direct support | |
| [WebSocket](https://docs.fluentbit.io/manual/pipeline/outputs/websocket) | No direct support | Workaround: Maybe with HTTP |
