package configcompiler

type Pipeline struct {
	Receivers  []string `json:"receivers"`
	Processors []string `json:"processors"`
	Exporters  []string `json:"exporters"`
}

type Pipelines map[string]Pipeline
