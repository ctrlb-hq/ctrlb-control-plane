package constants

var (
	AGENT_CONFIG_PATH      = "./config.yaml"
	AGENT_TYPE             = "otel"
	AGENT_VERSION          = "3.1.5"
	BACKEND_URL            = "http://controlplane.ctrlb.ai:8096"
	PORT                   = "3421"
	TESTING                = false
	PIPELINE_NAME          = ""
	STARTED_BY             = "Admin"
	HEARTBEAT_INTERVAL_SEC = 30
	SKIP_CONFIG_VALIDATION = false
)

var AGENTID int64

var SUPPORTED_AGENT_TYPES = []string{"otel", "fluent-bit"}

const (
	FluentBitDefaultHTTPPort = "2020"       //default HTTP monitoring port for Fluent Bit
	FluentBitDefaultHTTPHost = "127.0.0.1"  //default HTTP monitoring host for Fluent Bit
	FluentBitBinaryName      = "fluent-bit" //name of the Fluent Bit executable
	FluentBitStartupTimeout  = 30           // maximum time in seconds to wait for Fluent Bit to start
	FluentBitShutdownTimeout = 20           // maximum time in seconds to wait for Fluent Bit to shutdown gracefully
)
