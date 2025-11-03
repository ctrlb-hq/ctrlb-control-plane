package adapters

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
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

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
)

func componentsFactory() (otelcol.Factories, error) {
	factories := otelcol.Factories{}

	// Receivers
	factories.Receivers = map[component.Type]receiver.Factory{
		otlpreceiver.NewFactory().Type():                  otlpreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory().Type():           hostmetricsreceiver.NewFactory(),
		awscloudwatchmetricsreceiver.NewFactory().Type():  awscloudwatchmetricsreceiver.NewFactory(),
		awscloudwatchreceiver.NewFactory().Type():         awscloudwatchreceiver.NewFactory(),
		azuremonitorreceiver.NewFactory().Type():          azuremonitorreceiver.NewFactory(),
		filelogreceiver.NewFactory().Type():               filelogreceiver.NewFactory(),
		googlecloudmonitoringreceiver.NewFactory().Type(): googlecloudmonitoringreceiver.NewFactory(),
	}

	// Processors
	factories.Processors = map[component.Type]processor.Factory{
		batchprocessor.NewFactory().Type():                batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory().Type():        memorylimiterprocessor.NewFactory(),
		probabilisticsamplerprocessor.NewFactory().Type(): probabilisticsamplerprocessor.NewFactory(),
		attributesprocessor.NewFactory().Type():           attributesprocessor.NewFactory(),
		filterprocessor.NewFactory().Type():               filterprocessor.NewFactory(),
		tailsamplingprocessor.NewFactory().Type():         tailsamplingprocessor.NewFactory(),
	}

	// Exporters
	factories.Exporters = map[component.Type]exporter.Factory{
		otlpexporter.NewFactory().Type():       otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory().Type():   otlphttpexporter.NewFactory(),
		debugexporter.NewFactory().Type():      debugexporter.NewFactory(),
		kafkaexporter.NewFactory().Type():      kafkaexporter.NewFactory(),
		prometheusexporter.NewFactory().Type(): prometheusexporter.NewFactory(),
	}

	// Connectors (empty for now, but initialized)
	factories.Connectors = map[component.Type]connector.Factory{}

	// Extensions
	factories.Extensions = map[component.Type]extension.Factory{
		healthcheckextension.NewFactory().Type(): healthcheckextension.NewFactory(),
	}

	return factories, nil
}

func getNewOTELCollector() (*otelcol.Collector, error) {
	fmp := fileprovider.NewFactory()
	configProviderSettings := otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs:              []string{constants.AGENT_CONFIG_PATH},
			ProviderFactories: []confmap.ProviderFactory{fmp},
		},
	}

	settings := otelcol.CollectorSettings{
		BuildInfo:              component.NewDefaultBuildInfo(),
		Factories:              componentsFactory,
		ConfigProviderSettings: configProviderSettings,
		LoggingOptions: []zap.Option{
			zap.WrapCore(func(_ zapcore.Core) zapcore.Core {
				return logger.Logger.Core()
			}),
		},
	}

	return otelcol.NewCollector(settings)
}
