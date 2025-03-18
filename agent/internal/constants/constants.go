package constants

import "github.com/ctrlb-hq/ctrlb-collector/agent/internal/models"

var (
	AGENT_CONFIG_PATH = "./config.yaml"
	AGENT_TYPE        = "otel"
	AGENT_VERSION     = "3.1.5"
	BACKEND_URL       = "http://controlplane.ctrlb.ai:8096"
	PORT              = "443"
	TESTING           = false
	IS_PIPELINE       = false
)

var AGENT *models.AgentWithConfig
