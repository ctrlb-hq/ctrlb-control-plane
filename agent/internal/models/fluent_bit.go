package models

type FluentBitConfig struct {
	Service  map[string]interface{} `json:"service" yaml:"service"`
	Pipeline struct {
		Inputs  []map[string]interface{} `json:"inputs" yaml:"inputs"`
		Filters []map[string]interface{} `json:"filters" yaml:"filters"`
		Outputs []map[string]interface{} `json:"outputs" yaml:"outputs"`
	} `json:"pipeline" yaml:"pipeline"`
}
