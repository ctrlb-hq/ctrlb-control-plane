package adapters

import (
	"fmt"

	"go.opentelemetry.io/collector/otelcol"
)

// validateComponents validates individual components
func validateComponents(config *otelcol.Config, factories otelcol.Factories) error {
	if config == nil {
		return fmt.Errorf("collector config is nil")
	}

	// Validate receivers
	if config.Receivers != nil {
		for componentID, receiverConf := range config.Receivers {
			if receiverConf == nil {
				return fmt.Errorf("receiver config for '%s' is nil", componentID)
			}

			// Extract the component type from the component ID
			componentType := componentID.Type()

			if _, exists := factories.Receivers[componentType]; !exists {
				return fmt.Errorf("unknown receiver type: %s", componentType)
			}
		}
	}

	// Validate processors
	if config.Processors != nil {
		for componentID, processorConf := range config.Processors {
			if processorConf == nil {
				return fmt.Errorf("processor config for '%s' is nil", componentID)
			}

			// Extract the component type from the component ID
			componentType := componentID.Type()

			if _, exists := factories.Processors[componentType]; !exists {
				return fmt.Errorf("unknown processor type: %s", componentType)
			}
		}
	}

	// Validate exporters
	if config.Exporters != nil {
		for componentID, exporterConf := range config.Exporters {
			if exporterConf == nil {
				return fmt.Errorf("exporter config for '%s' is nil", componentID)
			}

			// Extract the component type from the component ID
			componentType := componentID.Type()

			if _, exists := factories.Exporters[componentType]; !exists {
				return fmt.Errorf("unknown exporter type: %s", componentType)
			}
		}
	}

	// Validate extensions
	if config.Extensions != nil {
		for componentID, extensionConf := range config.Extensions {
			if extensionConf == nil {
				return fmt.Errorf("extension config for '%s' is nil", componentID)
			}

			// Extract the component type from the component ID
			componentType := componentID.Type()

			if _, exists := factories.Extensions[componentType]; !exists {
				return fmt.Errorf("unknown extension type: %s", componentType)
			}
		}
	}

	// Validate service pipelines
	if err := validateServicePipelines(config); err != nil {
		return fmt.Errorf("invalid service configuration: %w", err)
	}

	return nil
}

// validateServicePipelines validates the service pipeline configuration
func validateServicePipelines(config *otelcol.Config) error {
	// Get all defined component names
	definedReceivers := make(map[string]bool)
	if config.Receivers != nil {
		for name := range config.Receivers {
			definedReceivers[name.String()] = true
		}
	}

	definedProcessors := make(map[string]bool)
	if config.Processors != nil {
		for name := range config.Processors {
			definedProcessors[name.String()] = true
		}
	}

	definedExporters := make(map[string]bool)
	if config.Exporters != nil {
		for name := range config.Exporters {
			definedExporters[name.String()] = true
		}
	}

	// Validate each pipeline
	for pipelineName, pipeline := range config.Service.Pipelines {
		if pipeline == nil {
			return fmt.Errorf("pipeline '%s' configuration is nil", pipelineName)
		}

		// Check receivers
		if len(pipeline.Receivers) == 0 {
			return fmt.Errorf("pipeline '%s' must have at least one receiver", pipelineName)
		}
		for _, receiver := range pipeline.Receivers {
			if !definedReceivers[receiver.String()] {
				return fmt.Errorf("pipeline '%s' references undefined receiver: %s", pipelineName, receiver)
			}
		}

		// Check processors (optional)
		for _, processor := range pipeline.Processors {
			if !definedProcessors[processor.String()] {
				return fmt.Errorf("pipeline '%s' references undefined processor: %s", pipelineName, processor)
			}
		}

		// Check exporters
		if len(pipeline.Exporters) == 0 {
			return fmt.Errorf("pipeline '%s' must have at least one exporter", pipelineName)
		}
		for _, exporter := range pipeline.Exporters {
			if !definedExporters[exporter.String()] {
				return fmt.Errorf("pipeline '%s' references undefined exporter: %s", pipelineName, exporter)
			}
		}
	}

	return nil
}
