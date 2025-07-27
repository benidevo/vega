package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		envValue      string
		fileContent   string
		setupFile     bool
		expectedValue string
		expectWarning bool
	}{
		{
			name:          "returns default when no env vars set",
			key:           "TEST_VAR",
			defaultValue:  "default",
			expectedValue: "default",
		},
		{
			name:          "returns env var when set",
			key:           "TEST_VAR",
			defaultValue:  "default",
			envValue:      "from-env",
			expectedValue: "from-env",
		},
		{
			name:          "reads from file when _FILE var is set",
			key:           "TEST_VAR",
			defaultValue:  "default",
			fileContent:   "from-file",
			setupFile:     true,
			expectedValue: "from-file",
		},
		{
			name:          "trims whitespace from file content",
			key:           "TEST_VAR",
			defaultValue:  "default",
			fileContent:   "  from-file\n\t",
			setupFile:     true,
			expectedValue: "from-file",
		},
		{
			name:          "file takes precedence over env var",
			key:           "TEST_VAR",
			defaultValue:  "default",
			envValue:      "from-env",
			fileContent:   "from-file",
			setupFile:     true,
			expectedValue: "from-file",
		},
		{
			name:          "falls back to env var when file read fails",
			key:           "TEST_VAR",
			defaultValue:  "default",
			envValue:      "from-env",
			setupFile:     false, // This will cause file read to fail
			expectedValue: "from-env",
			expectWarning: true,
		},
		{
			name:          "empty file returns empty string",
			key:           "TEST_VAR",
			defaultValue:  "default",
			fileContent:   "",
			setupFile:     true,
			expectedValue: "",
		},
		{
			name:          "multiline file content is trimmed",
			key:           "TEST_VAR",
			defaultValue:  "default",
			fileContent:   "line1\nline2\nline3",
			setupFile:     true,
			expectedValue: "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			os.Unsetenv(tt.key + "_FILE")
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			var tempFile string
			if tt.setupFile || tt.fileContent != "" {
				tempDir := t.TempDir()
				tempFile = filepath.Join(tempDir, "secret")

				if tt.setupFile {
					err := os.WriteFile(tempFile, []byte(tt.fileContent), 0600)
					if err != nil {
						t.Fatalf("Failed to create temp file: %v", err)
					}
				}

				os.Setenv(tt.key+"_FILE", tempFile)
				defer os.Unsetenv(tt.key + "_FILE")
			} else if tt.expectWarning {
				os.Setenv(tt.key+"_FILE", "/nonexistent/file")
				defer os.Unsetenv(tt.key + "_FILE")
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expectedValue {
				t.Errorf("Expected %q, got %q", tt.expectedValue, result)
			}
		})
	}
}

func TestGetEnvConcurrency(t *testing.T) {
	tempDir := t.TempDir()
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("CONCURRENT_TEST_%d", i)
		content := fmt.Sprintf("value-%d", i)

		tempFile := filepath.Join(tempDir, fmt.Sprintf("secret-%d", i))
		err := os.WriteFile(tempFile, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		os.Setenv(key+"_FILE", tempFile)
		defer os.Unsetenv(key + "_FILE")
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			key := fmt.Sprintf("CONCURRENT_TEST_%d", idx)
			expected := fmt.Sprintf("value-%d", idx)

			for j := 0; j < 100; j++ {
				result := getEnv(key, "default")
				if result != expected {
					t.Errorf("Concurrent read failed: expected %q, got %q", expected, result)
				}
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGetEnvSecurityScenarios(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		fileContent   string
		expectedValue string
		description   string
	}{
		{
			name:          "handles docker secret format",
			key:           "DOCKER_SECRET",
			fileContent:   "my-secret-value",
			expectedValue: "my-secret-value",
			description:   "Docker secrets often don't have newlines",
		},
		{
			name:          "handles kubernetes secret format",
			key:           "K8S_SECRET",
			fileContent:   "my-secret-value\n",
			expectedValue: "my-secret-value",
			description:   "Kubernetes secrets often have trailing newlines",
		},
		{
			name:          "handles base64 decoded content",
			key:           "BASE64_SECRET",
			fileContent:   "dGVzdC1zZWNyZXQ=",
			expectedValue: "dGVzdC1zZWNyZXQ=",
			description:   "Should handle base64 strings as-is",
		},
		{
			name:          "handles complex passwords",
			key:           "COMPLEX_PASSWORD",
			fileContent:   "p@ssw0rd!#$%^&*()_+-=[]{}|;':\",./<>?\n",
			expectedValue: "p@ssw0rd!#$%^&*()_+-=[]{}|;':\",./<>?",
			description:   "Should preserve special characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			os.Unsetenv(tt.key + "_FILE")
			tempDir := t.TempDir()
			tempFile := filepath.Join(tempDir, "secret")
			err := os.WriteFile(tempFile, []byte(tt.fileContent), 0600)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			os.Setenv(tt.key+"_FILE", tempFile)
			defer os.Unsetenv(tt.key + "_FILE")

			result := getEnv(tt.key, "")

			if result != tt.expectedValue {
				t.Errorf("%s: expected %q, got %q", tt.description, tt.expectedValue, result)
			}
		})
	}
}

func TestGetEnvSecurityChecks(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		fileEnvVar    string
		expectedValue string
		expectWarning bool
	}{
		{
			name:          "rejects relative file paths",
			key:           "TEST_VAR",
			defaultValue:  "default",
			fileEnvVar:    "relative/path/to/secret",
			expectedValue: "default",
			expectWarning: true,
		},
		{
			name:          "rejects paths with parent directory traversal",
			key:           "TEST_VAR",
			defaultValue:  "default",
			fileEnvVar:    "/etc/../etc/passwd",
			expectedValue: "default",
			expectWarning: true,
		},
		{
			name:          "rejects paths with multiple parent directory traversals",
			key:           "TEST_VAR",
			defaultValue:  "default",
			fileEnvVar:    "/run/secrets/../../etc/passwd",
			expectedValue: "default",
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			os.Unsetenv(tt.key + "_FILE")

			if tt.fileEnvVar != "" {
				os.Setenv(tt.key+"_FILE", tt.fileEnvVar)
				defer os.Unsetenv(tt.key + "_FILE")
			}

			result := getEnv(tt.key, tt.defaultValue)

			if result != tt.expectedValue {
				t.Errorf("Expected %q, got %q", tt.expectedValue, result)
			}
		})
	}
}

func TestGetEnvFileSizeLimit(t *testing.T) {
	os.Unsetenv("TEST_VAR")
	os.Unsetenv("TEST_VAR_FILE")

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "large-secret")

	largeContent := make([]byte, 2*maxSecretFileSize)
	for i := range largeContent {
		largeContent[i] = 'a'
	}

	err := os.WriteFile(tempFile, largeContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	os.Setenv("TEST_VAR_FILE", tempFile)
	defer os.Unsetenv("TEST_VAR_FILE")

	result := getEnv("TEST_VAR", "default")

	if result != "default" {
		t.Errorf("Expected default value for large file, got %q", result)
	}
}
