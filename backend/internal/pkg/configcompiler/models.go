package configcompiler

type Pipeline struct {
	Receivers  []string
	Processors []string
	Exporters  []string
}

type Pipelines map[string]Pipeline
