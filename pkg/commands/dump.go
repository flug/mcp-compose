package commands

import (
	"fmt"

	"github.com/flug/mcp-compose/pkg/config"
	"github.com/flug/mcp-compose/pkg/models"
	"gopkg.in/yaml.v3"
)

// Dump reads the Claude config and outputs MCP servers as YAML
func Dump() error {
	// Check if initialized
	if err := EnsureInitialized(); err != nil {
		return err
	}

	claudeConfig, err := config.ReadClaudeConfig()
	if err != nil {
		return err
	}

	// Convert to YAML structure
	yamlConfig := models.YAMLConfig{
		Projects: make(map[string]models.ProjectServers),
	}

	for projectPath, projectConfig := range claudeConfig.Projects {
		// Only include projects that have MCP servers
		if len(projectConfig.MCPServers) > 0 {
			yamlConfig.Projects[projectPath] = models.ProjectServers{
				MCPServers: projectConfig.MCPServers,
			}
		}
	}

	yamlData, err := yaml.Marshal(&yamlConfig)
	if err != nil {
		return fmt.Errorf("error converting to YAML: %w", err)
	}

	fmt.Print(string(yamlData))
	return nil
}
