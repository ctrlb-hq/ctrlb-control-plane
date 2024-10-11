package models

type OTELConfig struct {
	Receivers  map[string]interface{} `json:"receivers" yaml:"receivers"`
	Processors map[string]interface{} `json:"processors" yaml:"processors"`
	Exporters  map[string]interface{} `json:"exporters" yaml:"exporters"`
	Service    struct {
		Pipelines map[string]struct {
			Receivers  []string `json:"receivers" yaml:"receivers"`
			Processors []string `json:"processors" yaml:"processors"`
			Exporters  []string `json:"exporters" yaml:"exporters"`
		} `json:"pipelines" yaml:"pipelines"`
	} `json:"service" yaml:"service"`
}
