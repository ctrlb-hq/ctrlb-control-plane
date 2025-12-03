## Filter & Processor Plugin Mapping
This document provides a reference for mapping Fluent Bit filters plugins to their OpenTelemetry (OTel) Collector exporter equivalents. It is intended to help users migrating pipelines from Fluent Bit to the OTel Collector.
| Fluent Bit Filter/Processor | OTel Collector Equivalent | Status / Notes |
| :--- | :--- | :--- |
| [AWS Metadata](https://docs.fluentbit.io/manual/pipeline/filters/aws-metadata) | [Attribute Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor) | 🚧 **Beta** |
| [CheckList](https://docs.fluentbit.io/manual/pipeline/filters/checklist) | No direct support | Workaround: Maybe with FilterProcessor |
| [ECS Metadata](https://docs.fluentbit.io/manual/pipeline/filters/ecs-metadata) | No direct support | Workaround: Maybe with AttributeProcessor |
| [Expect](https://docs.fluentbit.io/manual/pipeline/filters/expect) | No direct support | Workaround: Filter, Attribute, and Transform processor |
| [GeoIP2](https://docs.fluentbit.io/manual/pipeline/filters/geoip2-filter) | No direct support | |
| [Grep](https://docs.fluentbit.io/manual/pipeline/filters/grep) | No direct support | Workaround: Maybe with Filter processor |
| [Kubernetes](https://docs.fluentbit.io/manual/pipeline/filters/kubernetes) | [K8s Attribute Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/k8sattributesprocessor) | 🚧 **Beta** |
| [Log to Metrics](https://docs.fluentbit.io/manual/pipeline/filters/log_to_metrics) | No direct support | Workaround: Maybe with Transform processor |
| [Lua](https://docs.fluentbit.io/manual/pipeline/filters/lua) | No direct support | Workaround: Maybe with Transform processor |
| [Parser](https://docs.fluentbit.io/manual/pipeline/filters/parser) | No direct support | Workaround: Maybe with AttributeProcessor |
| [Record Modifier](https://docs.fluentbit.io/manual/pipeline/filters/record-modifier) | [Attribute Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor) | 🚧 **Beta** |
| [Modify](https://docs.fluentbit.io/manual/pipeline/filters/modify) | [Attribute Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor) | 🚧 **Beta** |
| [Multiline or Stack trace](https://docs.fluentbit.io/manual/pipeline/filters/multiline-stacktrace) | No direct support | Workaround: Maybe with Attribute, Batch and Transform processor |
| [Nest](https://docs.fluentbit.io/manual/pipeline/filters/nest) | No direct support | Workaround: Maybe with Transform processor |
| [Nightfall](https://docs.fluentbit.io/manual/pipeline/filters/nightfall) | No direct support | Workaround: Maybe with Attributes and Batch processor |
| [Rewrite-tag](https://docs.fluentbit.io/manual/pipeline/filters/rewrite-tag) | [Attribute Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor) | 🚧 **Beta** |
| [Standard-output](https://docs.fluentbit.io/manual/pipeline/filters/standard-output) | No direct support | Primarily handled with exporter |
| [Sysinfo](https://docs.fluentbit.io/manual/pipeline/filters/sysinfo) | No direct support | Workaround: Maybe with Hostmetrics receiver |
| [Throttle](https://docs.fluentbit.io/manual/pipeline/filters/throttle) | [Filter Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor) | Use rate-limit config |
| [Type-converter](https://docs.fluentbit.io/manual/pipeline/filters/type-converter) | [Attribute Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor) | 🚧 **Beta** |
| [Tensorflow](https://docs.fluentbit.io/manual/pipeline/filters/tensorflow) | No direct support | Workaround: Attribute, Batch and filter processor |
| [Wasm](https://docs.fluentbit.io/manual/pipeline/filters/wasm) | No direct support | Workaround: Maybe with Transform processor |
| [Content-modifier](https://docs.fluentbit.io/manual/pipeline/processors/content-modifier) | [Attribute Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor) | 🚧 **Beta** |
| [Labels](https://docs.fluentbit.io/manual/pipeline/processors/labels) | [Attribute Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor) | 🚧 **Beta** |
| [Metrics Selector](https://docs.fluentbit.io/manual/pipeline/processors/metrics-selector) | [Metrics Transform Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor) | 🚧 **Beta** |
| [SQL](https://docs.fluentbit.io/manual/pipeline/processors/sql) | No direct support | Workaround: Attribute and transform processor |
