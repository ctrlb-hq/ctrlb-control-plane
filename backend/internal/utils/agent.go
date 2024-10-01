package utils

import (
	"fmt"
)

func GenerateAgentName(typ string, version string, hostname string) string {
	return fmt.Sprintf("%s_%s@%s", typ, version, hostname)
}
