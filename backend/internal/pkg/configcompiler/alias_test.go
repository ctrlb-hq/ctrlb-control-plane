package configcompiler

import (
	"testing"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Alias Generation Tests
// =============================================================================

func TestGenerateOTELAlias(t *testing.T) {
	node := models.PipelineNodes{
		Name:          "my_receiver",
		ComponentName: "otlp_receiver",
		Config:        map[string]any{"endpoint": "0.0.0.0:4317"},
	}

	alias := GenerateOTELAlias(node)

	// Verify the alias has the expected format: type/name_hash
	assert.Contains(t, alias, "otlp/")
	assert.Contains(t, alias, "my_receiver") // The actual name format
	assert.Contains(t, alias, "_")           // Should have underscore before hash
}

func TestGenerateFBAlias(t *testing.T) {
	node := models.PipelineNodes{
		Name:          "my_input",
		ComponentName: "tail",
		Config:        map[string]any{"path": "/var/log/*.log"},
	}

	alias := GenerateFBAlias(node)

	// Verify the alias has the expected format: name_hash
	assert.Contains(t, alias, "my_input") // The actual name format
	assert.NotContains(t, alias, "/")     // FB aliases don't have type prefix
	// Should have format: name_hash
	assert.Regexp(t, `^my_input_[a-f0-9]+$`, alias)
}

func TestGenerateFBTag(t *testing.T) {
	node := models.PipelineNodes{
		ComponentName: "tail",
	}

	tag := GenerateFBTag(node, "myInput_abc123")

	assert.Equal(t, "tail.myInput_abc123", tag)
}

func TestExtractTagPrefix(t *testing.T) {
	tests := []struct {
		tag      string
		expected string
	}{
		{"tail.myInput_abc", "tail"},
		{"stdin.input", "stdin"},
		{"notag", "notag"},
		{"a.b.c", "a"},
	}

	for _, tt := range tests {
		result := ExtractTagPrefix(tt.tag)
		assert.Equal(t, tt.expected, result)
	}
}

// =============================================================================
// FB Match Pattern Tests
// =============================================================================

func TestComputeFBMatchPattern(t *testing.T) {
	t.Run("no tags", func(t *testing.T) {
		compatible, pattern := computeFBMatchPattern([]FBTagState{})
		assert.True(t, compatible)
		assert.Equal(t, "*", pattern)
	})

	t.Run("single tag", func(t *testing.T) {
		tags := []FBTagState{{Tag: "tail.myInput_abc123"}}
		compatible, pattern := computeFBMatchPattern(tags)
		assert.True(t, compatible)
		assert.Equal(t, "tail.myInput_abc123*", pattern)
	})

	t.Run("multiple tags same plugin with common prefix", func(t *testing.T) {
		tags := []FBTagState{
			{Tag: "tail.input1_abc"},
			{Tag: "tail.input2_def"},
		}
		compatible, pattern := computeFBMatchPattern(tags)
		assert.True(t, compatible)
		assert.True(t, pattern == "tail.input*" || pattern == "tail*")
	})

	t.Run("multiple tags different plugins", func(t *testing.T) {
		tags := []FBTagState{
			{Tag: "tail.input1_abc"},
			{Tag: "stdin.input2_def"},
		}
		compatible, pattern := computeFBMatchPattern(tags)
		assert.False(t, compatible)
		assert.Equal(t, "", pattern)
	})
}

