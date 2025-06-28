package gemini

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_GetModelForTask(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		taskType      string
		expectedModel string
	}{
		{
			name: "CV parsing uses specific model",
			config: &Config{
				Model:          "default-model",
				ModelCVParsing: "gemini-1.5-flash",
			},
			taskType:      "cv_parsing",
			expectedModel: "gemini-1.5-flash",
		},
		{
			name: "job analysis uses specific model",
			config: &Config{
				Model:            "default-model",
				ModelJobAnalysis: "gemini-2.5-flash",
			},
			taskType:      "job_analysis",
			expectedModel: "gemini-2.5-flash",
		},
		{
			name: "fallback to default when task model empty",
			config: &Config{
				Model:          "default-model",
				ModelCVParsing: "",
			},
			taskType:      "cv_parsing",
			expectedModel: "default-model",
		},
		{
			name: "unknown task uses default",
			config: &Config{
				Model: "default-model",
			},
			taskType:      "unknown_task",
			expectedModel: "default-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetModelForTask(tt.taskType)
			assert.Equal(t, tt.expectedModel, result)
		})
	}
}
