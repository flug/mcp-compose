package commands

import (
	"fmt"
	"os"

	"github.com/flug/mcp-compose/pkg/config"
	"github.com/flug/mcp-compose/pkg/models"
	"gopkg.in/yaml.v3"
)

// Apply reads a YAML file and updates the Claude config
func Apply(yamlFile string) error {
	// Check if initialized
	if err := EnsureInitialized(); err != nil {
		return err
	}

	yamlData, err := os.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %w", err)
	}

	var yamlConfig models.YAMLConfig
	if err := yaml.Unmarshal(yamlData, &yamlConfig); err != nil {
		return fmt.Errorf("error parsing YAML: %w", err)
	}

	// Read existing Claude config
	claudeConfig, err := config.ReadClaudeConfig()
	if err != nil {
		return err
	}

	// Initialize projects map if nil
	if claudeConfig.Projects == nil {
		claudeConfig.Projects = make(map[string]models.ProjectConfig)
	}

	// Update MCP servers for each project
	for projectPath, projectServers := range yamlConfig.Projects {
		// Get or create project config
		projectConfig, exists := claudeConfig.Projects[projectPath]
		if !exists {
			projectConfig = models.ProjectConfig{
				MCPServers: make(map[string]models.MCPServer),
			}
		}

		// Initialize MCPServers map if nil
		if projectConfig.MCPServers == nil {
			projectConfig.MCPServers = make(map[string]models.MCPServer)
		}

		// Update MCP servers
		projectConfig.MCPServers = projectServers.MCPServers

		claudeConfig.Projects[projectPath] = projectConfig
	}

	// Write back to file
	if err := config.WriteClaudeConfig(claudeConfig); err != nil {
		return err
	}

	if configPath, err := config.GetClaudeConfigPath(); err == nil {
		fmt.Printf("Successfully updated %s\n", configPath)
	}
	fmt.Printf("Updated %d project(s)\n", len(yamlConfig.Projects))

	return nil
}
