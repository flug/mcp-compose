package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flug/mcp-compose/pkg/platform"
)

func TestEnsureInitialized(t *testing.T) {
	// Save original state
	platformInfo, err := platform.Detect()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}

	originalConfigPath := platformInfo.MCPComposeConfig
	backupPath := originalConfigPath + ".backup"

	// Backup existing config if it exists
	if platformInfo.MCPComposeConfigExists() {
		data, err := os.ReadFile(originalConfigPath)
		if err == nil {
			_ = os.WriteFile(backupPath, data, 0644)
			defer func() {
				_ = os.WriteFile(originalConfigPath, data, 0644)
				_ = os.Remove(backupPath)
			}()
		}
	}

	tests := []struct {
		name         string
		configExists bool
		wantErr      bool
		errContains  string
	}{
		{
			name:         "config exists - should pass",
			configExists: true,
			wantErr:      false,
		},
		{
			name:         "config does not exist - should fail",
			configExists: false,
			wantErr:      true,
			errContains:  "not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: remove or create config based on test case
			if tt.configExists {
				// Create a minimal config
				if err := os.MkdirAll(filepath.Dir(originalConfigPath), 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				if err := os.WriteFile(originalConfigPath, []byte(`{"version":"1.0.0"}`), 0644); err != nil {
					t.Fatalf("Failed to create config: %v", err)
				}
			} else {
				// Remove config
				_ = os.Remove(originalConfigPath)
			}

			// Test
			err := EnsureInitialized()

			// Verify
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureInitialized() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestCommandsRequireInit(t *testing.T) {
	// This test verifies that commands properly check initialization
	platformInfo, err := platform.Detect()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}

	// Backup and remove config
	originalConfigPath := platformInfo.MCPComposeConfig
	var backupData []byte
	hadConfig := false

	if platformInfo.MCPComposeConfigExists() {
		backupData, _ = os.ReadFile(originalConfigPath)
		hadConfig = true
		_ = os.Remove(originalConfigPath)
		defer func() {
			if hadConfig {
				_ = os.WriteFile(originalConfigPath, backupData, 0644)
			}
		}()
	}

	// Test that each command fails without initialization
	commands := []struct {
		name string
		fn   func() error
	}{
		{
			name: "dump",
			fn:   func() error { return Dump() },
		},
		{
			name: "apply",
			fn:   func() error { return Apply("nonexistent.yaml") },
		},
		{
			name: "convert",
			fn:   func() error { return Convert("nonexistent.json", "/test") },
		},
		{
			name: "delete",
			fn:   func() error { return Delete("nonexistent.yaml", "server", "/test") },
		},
	}

	for _, cmd := range commands {
		t.Run(cmd.name, func(t *testing.T) {
			err := cmd.fn()
			if err == nil {
				t.Errorf("%s should fail without initialization", cmd.name)
			}
			if !contains(err.Error(), "not initialized") {
				t.Errorf("%s error should mention initialization, got: %v", cmd.name, err)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
