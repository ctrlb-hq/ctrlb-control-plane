package constants

var (
	AGENT_CONFIG_PATH = "./config.yaml"
	AGENT_TYPE        = "otel"
	AGENT_VERSION     = "3.1.5"
	BACKEND_URL       = "http://controlplane.ctrlb.ai:8096"
	PORT              = "3421"
	TESTING           = false
	PIPELINE_NAME     = ""
	STARTED_BY        = "Admin"
)

var AGENTID int64

var SUPPORTED_AGENT_TYPES = []string{"otel", "fluent-bit"}

const (
	// FluentBitDefaultHTTPPort is the default HTTP monitoring port for Fluent Bit
	FluentBitDefaultHTTPPort = "2020"

	// FluentBitDefaultHTTPHost is the default HTTP monitoring host for Fluent Bit
	FluentBitDefaultHTTPHost = "127.0.0.1"

	// FluentBitBinaryName is the name of the Fluent Bit executable
	FluentBitBinaryName = "fluent-bit"

	// FluentBitStartupTimeout is the maximum time to wait for Fluent Bit to start
	FluentBitStartupTimeout = 30 // seconds

	// FluentBitShutdownTimeout is the maximum time to wait for Fluent Bit to shutdown gracefully
	FluentBitShutdownTimeout = 20 // seconds
)
