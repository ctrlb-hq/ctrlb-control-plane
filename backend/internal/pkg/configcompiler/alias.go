package configcompiler

import (
	"fmt"
	"strings"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

// GenerateOTELAlias creates a unique OTEL component alias
// Format: componentType/nodeName_configHash
func GenerateOTELAlias(node models.PipelineNodes) string {
	componentType := utils.TrimAfterUnderscore(node.ComponentName)
	nodeName := utils.ToCamelCase(node.Name)
	configHash := utils.HashFromConfig(node.Config)

	return fmt.Sprintf("%s/%s_%s", componentType, nodeName, configHash)
}

// GenerateFBAlias creates a unique Fluent Bit alias
// Format: nodeName_configHash
func GenerateFBAlias(node models.PipelineNodes) string {
	nodeName := utils.ToCamelCase(node.Name)
	configHash := utils.HashFromConfig(node.Config)

	return fmt.Sprintf("%s_%s", nodeName, configHash)
}

// GenerateFBTag creates a tag for a Fluent Bit input
// Format: pluginName.alias
func GenerateFBTag(node models.PipelineNodes, alias string) string {
	pluginName := utils.TrimAfterUnderscore(node.ComponentName)
	return fmt.Sprintf("%s.%s", pluginName, alias)
}

// GetPluginName extracts the plugin name from a component name
func GetPluginName(node models.PipelineNodes) string {
	return utils.TrimAfterUnderscore(node.ComponentName)
}

// ExtractTagPrefix extracts the plugin prefix from a tag
// e.g., "tail.myInput_abc" -> "tail"
func ExtractTagPrefix(tag string) string {
	idx := strings.Index(tag, ".")
	if idx == -1 {
		return tag
	}
	return tag[:idx]
}
