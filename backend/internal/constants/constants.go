package constants

var (
	PORT                       = "8096"
	ENV                        = "dev"
	AGENT_INACTIVE_TIMEOUT_SEC = 120 // Mark agent as inactive after this many seconds without heartbeat
	STALENESS_CHECK_SEC        = 30  // Check for stale agents every N seconds
)

var JWT_SECRET string
