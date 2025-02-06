package models

// ConfigNew represents the top-level configuration.
type ConfigNew struct {
	Version    string           `yaml:"version" json:"version"`
	Global     GlobalConfig     `yaml:"global" json:"global"`
	Extensions ExtensionsConfig `yaml:"extensions" json:"extensions"`
	Pipelines  []Pipeline       `yaml:"pipelines" json:"pipelines"`
}

// GlobalConfig holds global settings.
type GlobalConfig struct {
	LogLevel    string            `yaml:"log_level" json:"log_level"`
	Credentials map[string]string `yaml:"credentials,omitempty" json:"credentials,omitempty"`
}

// ExtensionsConfig contains extension settings.
type ExtensionsConfig struct {
	HealthCheck *HealthCheckExtension  `yaml:"health_check,omitempty" json:"health_check,omitempty"`
	Extra       map[string]interface{} `yaml:",inline" json:"extra,omitempty"`
}

// HealthCheckExtension captures health check settings.
type HealthCheckExtension struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Endpoint string `yaml:"endpoint" json:"endpoint"`
}

// Pipeline represents a pipeline.
type Pipeline struct {
	Name         string      `yaml:"name" json:"name"`
	Sources      []Component `yaml:"sources" json:"sources"`
	Processors   []Component `yaml:"processors" json:"processors"`
	Destinations []Component `yaml:"destinations" json:"destinations"`
}

// Component is a generic structure for sources, processors, or destinations.
type Component struct {
	Type   string                 `yaml:"type" json:"type"`
	Name   string                 `yaml:"name" json:"name"`
	Config map[string]interface{} `yaml:"config" json:"config"`
}
