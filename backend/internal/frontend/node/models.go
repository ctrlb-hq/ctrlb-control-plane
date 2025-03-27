package frontendnode

type ComponentInfo struct {
	Name             string   `json:"name"`
	DisplayName      string   `json:"display_name"`
	Type             string   `json:"type"`
	SupportedSignals []string `json:"supported_signals"`
}
