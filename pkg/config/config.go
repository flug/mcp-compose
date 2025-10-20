package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/flug/mcp-compose/pkg/models"
	"github.com/flug/mcp-compose/pkg/platform"
)

// GetClaudeConfigPath returns the path to the Claude configuration file
func GetClaudeConfigPath() (string, error) {
	// Check if we have a saved config
	platformInfo, err := platform.Detect()
	if err != nil {
		return "", fmt.Errorf("error detecting platform: %w", err)
	}

	// Try to read mcp-compose config
	if platformInfo.MCPComposeConfigExists() {
		data, err := os.ReadFile(platformInfo.MCPComposeConfig)
		if err == nil {
			var config models.MCPComposeConfig
			if err := json.Unmarshal(data, &config); err == nil && config.ClaudeConfigPath != "" {
				return config.ClaudeConfigPath, nil
			}
		}
	}

	// Fallback to platform detection
	return platformInfo.ClaudeConfig, nil
}

// ReadClaudeConfig reads and parses the Claude configuration file
func ReadClaudeConfig() (*models.ClaudeConfig, error) {
	configPath, err := GetClaudeConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", configPath, err)
	}

	var config models.ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &config, nil
}

// WriteClaudeConfig writes the Claude configuration to file
func WriteClaudeConfig(config *models.ClaudeConfig) error {
	configPath, err := GetClaudeConfigPath()
	if err != nil {
		return err
	}

	// Read the original file to preserve all fields
	originalData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading original config: %w", err)
	}

	var originalConfig map[string]interface{}
	if err := json.Unmarshal(originalData, &originalConfig); err != nil {
		return fmt.Errorf("error parsing original config: %w", err)
	}

	// Update only the projects field
	configData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	var partialConfig map[string]interface{}
	if err := json.Unmarshal(configData, &partialConfig); err != nil {
		return fmt.Errorf("error unmarshaling partial config: %w", err)
	}

	// Merge the projects field
	if projects, ok := partialConfig["projects"]; ok {
		originalConfig["projects"] = projects
	}

	// Write back with indentation
	jsonData, err := json.MarshalIndent(originalConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("error converting to JSON: %w", err)
	}

	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing to %s: %w", configPath, err)
	}

	return nil
}
