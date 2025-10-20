package commands

import (
	"fmt"
	"strings"

	"github.com/flug/mcp-compose/pkg/models"
)

// ValidateMCPServer validates an MCP server configuration
func ValidateMCPServer(name string, server models.MCPServer) error {
	if name == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	// Validate type
	if server.Type != "" && server.Type != "stdio" && server.Type != "http" {
		return fmt.Errorf("invalid type '%s' for server '%s': must be 'stdio' or 'http'", server.Type, name)
	}

	// Validate stdio type
	if server.Type == "stdio" || (server.Type == "" && server.Command != "") {
		if server.Command == "" {
			return fmt.Errorf("server '%s': command is required for stdio type", name)
		}
		if server.URL != "" {
			return fmt.Errorf("server '%s': stdio type cannot have URL field", name)
		}
		if len(server.Headers) > 0 {
			return fmt.Errorf("server '%s': stdio type cannot have headers", name)
		}
	}

	// Validate http type
	if server.Type == "http" {
		if server.URL == "" {
			return fmt.Errorf("server '%s': URL is required for http type", name)
		}
		if !strings.HasPrefix(server.URL, "http://") && !strings.HasPrefix(server.URL, "https://") {
			return fmt.Errorf("server '%s': URL must start with http:// or https://", name)
		}
		if server.Command != "" {
			return fmt.Errorf("server '%s': http type cannot have command", name)
		}
		if len(server.Args) > 0 {
			return fmt.Errorf("server '%s': http type cannot have args", name)
		}
	}

	// Validate scope
	if server.Scope != "" {
		validScopes := []string{"local", "user", "project"}
		valid := false
		for _, s := range validScopes {
			if server.Scope == s {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid scope '%s' for server '%s': must be 'local', 'user', or 'project'", server.Scope, name)
		}
	}

	return nil
}

// ValidateYAMLConfig validates a YAML configuration
func ValidateYAMLConfig(config *models.YAMLConfig) error {
	if len(config.Projects) == 0 {
		return fmt.Errorf("configuration must have at least one project")
	}

	for projectPath, projectConfig := range config.Projects {
		if projectPath == "" {
			return fmt.Errorf("project path cannot be empty")
		}

		if len(projectConfig.MCPServers) == 0 {
			return fmt.Errorf("project '%s': must have at least one MCP server", projectPath)
		}

		for serverName, server := range projectConfig.MCPServers {
			if err := ValidateMCPServer(serverName, server); err != nil {
				return fmt.Errorf("project '%s': %w", projectPath, err)
			}
		}
	}

	return nil
}
