package adapters

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/service/telemetry/otelconftelemetry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/constants"
	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azuremonitorreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudmonitoringreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
)

func componentsFactory() (otelcol.Factories, error) {
	factories := otelcol.Factories{
		Telemetry: otelconftelemetry.NewFactory(),
	}
	var err error

	// MakeFactoryMap retains upstream aliases for configurations created before
	// component types adopted names such as file_log and host_metrics.
	factories.Receivers, err = otelcol.MakeFactoryMap[receiver.Factory](
		otlpreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		awscloudwatchmetricsreceiver.NewFactory(),
		awscloudwatchreceiver.NewFactory(),
		azuremonitorreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
		googlecloudmonitoringreceiver.NewFactory(),
		kubeletstatsreceiver.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Processors, err = otelcol.MakeFactoryMap[processor.Factory](
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		probabilisticsamplerprocessor.NewFactory(),
		attributesprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		tailsamplingprocessor.NewFactory(),
		k8sattributesprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Exporters, err = otelcol.MakeFactoryMap[exporter.Factory](
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
		debugexporter.NewFactory(),
		kafkaexporter.NewFactory(),
		prometheusexporter.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	// This distribution does not register connectors.
	factories.Connectors = map[component.Type]connector.Factory{}

	factories.Extensions, err = otelcol.MakeFactoryMap[extension.Factory](
		healthcheckextension.NewFactory(),
		filestorage.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	return factories, nil
}

func getNewOTELCollector() (*otelcol.Collector, error) {
	return newCollector(constants.AGENT_CONFIG_PATH, false)
}

// NewManagedCollector creates a Collector whose lifecycle belongs to the
// management loop. Its config path is independent of the legacy VM globals.
func NewManagedCollector(configPath string) (*otelcol.Collector, error) {
	return newCollector(configPath, true)
}

func collectorConfigProviderSettings(configPath string) otelcol.ConfigProviderSettings {
	return otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs:              []string{configPath},
			ProviderFactories: []confmap.ProviderFactory{fileprovider.NewFactory(), envprovider.NewFactory()},
			DefaultScheme:     "env",
		},
	}

}

func newCollector(configPath string, managed bool) (*otelcol.Collector, error) {
	settings := otelcol.CollectorSettings{
		BuildInfo:               component.NewDefaultBuildInfo(),
		Factories:               componentsFactory,
		ConfigProviderSettings:  collectorConfigProviderSettings(configPath),
		DisableGracefulShutdown: managed,
		SkipSettingGRPCLogger:   managed,
		LoggingOptions: []zap.Option{
			zap.WrapCore(func(_ zapcore.Core) zapcore.Core {
				return logger.Logger.Core()
			}),
		},
	}

	return otelcol.NewCollector(settings)
}
