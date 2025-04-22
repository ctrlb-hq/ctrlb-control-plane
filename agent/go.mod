module github.com/ctrlb-hq/ctrlb-collector/agent

go 1.23.0

require (
	github.com/fsnotify/fsnotify v1.8.0
	github.com/gorilla/mux v1.8.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchmetricsreceiver v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchreceiver v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azuremonitorreceiver v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudmonitoringreceiver v0.122.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.122.0
	go.opentelemetry.io/collector/component v1.28.0
	go.opentelemetry.io/collector/confmap v1.28.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.28.0
	go.opentelemetry.io/collector/connector v0.122.0
	go.opentelemetry.io/collector/exporter v0.122.0
	go.opentelemetry.io/collector/exporter/debugexporter v0.122.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.122.0
	go.opentelemetry.io/collector/extension v1.28.0
	go.opentelemetry.io/collector/otelcol v0.122.0
	go.opentelemetry.io/collector/processor v0.122.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.122.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.122.0
	go.opentelemetry.io/collector/receiver v1.28.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.122.0
	go.uber.org/zap v1.27.0
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/stretchr/objx v0.5.2 // indirect

require (
	cloud.google.com/go/auth v0.15.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.7 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/monitoring v1.24.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.17.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.8.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor v0.11.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.2.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.3.3 // indirect
	github.com/IBM/sarama v1.45.1 // indirect
	github.com/alecthomas/participle/v2 v2.1.1 // indirect
	github.com/antchfx/xmlquery v1.4.4 // indirect
	github.com/antchfx/xpath v1.3.3 // indirect
	github.com/apache/thrift v0.21.0 // indirect
	github.com/aws/aws-msk-iam-sasl-signer-go v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.9 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.62 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.46.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.29.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.17 // indirect
	github.com/aws/smithy-go v1.22.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v4 v4.8.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/ebitengine/purego v0.8.2 // indirect
	github.com/elastic/go-grok v0.3.1 // indirect
	github.com/elastic/lunes v0.1.0 // indirect
	github.com/expr-lang/expr v1.17.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.5 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jaegertracing/jaeger-idl v0.5.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/joho/godotenv v1.5.1
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.2 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-syslog/v4 v4.2.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20190525184631-5f46317e436b // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magefile/mage v1.15.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kafka v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/core/xidutils v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/kafka/topic v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sampling v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.122.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin v0.122.0 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus-community/windows_exporter v0.27.2 // indirect
	github.com/prometheus/client_golang v1.21.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/shirou/gopsutil/v4 v4.25.2 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/stretchr/testify v1.10.0
	github.com/tilinna/clock v1.1.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/ua-parser/uap-go v0.0.0-20240611065828-3a4781585db6 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector v0.122.0 // indirect
	go.opentelemetry.io/collector/client v1.28.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.122.0 // indirect
	go.opentelemetry.io/collector/component/componenttest v0.122.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.122.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.28.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.122.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.122.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.28.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.28.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.28.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.122.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.28.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.122.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.122.0 // indirect
	go.opentelemetry.io/collector/connector/xconnector v0.122.0 // indirect
	go.opentelemetry.io/collector/consumer v1.28.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.122.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.122.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.122.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.122.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.122.0 // indirect
	go.opentelemetry.io/collector/exporter/exportertest v0.122.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.122.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v0.122.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.122.0 // indirect
	go.opentelemetry.io/collector/extension/extensiontest v0.122.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.122.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.28.0 // indirect
	go.opentelemetry.io/collector/filter v0.122.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.122.0 // indirect
	go.opentelemetry.io/collector/internal/memorylimiter v0.122.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.122.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.122.0 // indirect
	go.opentelemetry.io/collector/pdata v1.28.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.122.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.122.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.122.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.122.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper v0.122.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.122.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.122.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.122.0 // indirect
	go.opentelemetry.io/collector/receiver/receivertest v0.122.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.122.0 // indirect
	go.opentelemetry.io/collector/scraper v0.122.0 // indirect
	go.opentelemetry.io/collector/scraper/scraperhelper v0.122.0 // indirect
	go.opentelemetry.io/collector/semconv v0.122.0 // indirect
	go.opentelemetry.io/collector/service v0.122.0 // indirect
	go.opentelemetry.io/collector/service/hostcapabilities v0.122.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.10.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/contrib/otelconf v0.15.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.35.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.11.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.11.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.57.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.11.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.35.0 // indirect
	go.opentelemetry.io/otel/log v0.11.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.11.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/oauth2 v0.28.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	gonum.org/v1/gonum v0.15.1 // indirect
	google.golang.org/api v0.225.0 // indirect
	google.golang.org/genproto v0.0.0-20250122153221-138b5a5a4fd4 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/grpc v1.71.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
