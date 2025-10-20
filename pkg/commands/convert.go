package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/flug/mcp-compose/pkg/config"
	"github.com/flug/mcp-compose/pkg/models"
	"gopkg.in/yaml.v3"
)

// Convert reads a JSON file, validates it, converts to YAML, and optionally applies it
func Convert(jsonFile, projectPath string) error {
	// Check if initialized
	if err := EnsureInitialized(); err != nil {
		return err
	}

	// Read JSON file
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	// Try to parse as ConvertInput to detect format
	var input models.ConvertInput
	if err := json.Unmarshal(jsonData, &input); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	// Build the servers map
	servers := make(map[string]models.MCPServer)

	// Check if it's a single server or mcpServers map
	if len(input.MCPServers) > 0 {
		// It's an mcpServers map format
		servers = input.MCPServers
	} else if input.Command != "" || input.URL != "" {
		// It's a single server format - use filename as name
		serverName := strings.TrimSuffix(jsonFile, ".json")
		// Extract just the filename without path
		parts := strings.Split(serverName, "/")
		serverName = parts[len(parts)-1]

		servers[serverName] = models.MCPServer{
			Type:    input.Type,
			Command: input.Command,
			Args:    input.Args,
			Env:     input.Env,
			URL:     input.URL,
			Headers: input.Headers,
			Scope:   input.Scope,
		}
	} else {
		return fmt.Errorf("invalid JSON format: must contain either 'mcpServers' map or server configuration")
	}

	// Validate each server
	for name, server := range servers {
		if err := ValidateMCPServer(name, server); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	fmt.Println("✓ JSON validation successful")
	fmt.Printf("✓ Found %d MCP server(s)\n\n", len(servers))

	// Create a single reader for all user input
	reader := bufio.NewReader(os.Stdin)

	// Check for existing servers in the project
	existingConfig, err := config.ReadClaudeConfig()
	if err == nil && existingConfig.Projects != nil {
		if projectConfig, exists := existingConfig.Projects[projectPath]; exists {
			var conflictingServers []string
			for serverName := range servers {
				if _, exists := projectConfig.MCPServers[serverName]; exists {
					conflictingServers = append(conflictingServers, serverName)
				}
			}

			if len(conflictingServers) > 0 {
				fmt.Printf("⚠ Warning: The following MCP server(s) already exist in project '%s':\n", projectPath)
				for _, name := range conflictingServers {
					fmt.Printf("  - %s\n", name)
				}
				fmt.Print("\nDo you want to overwrite the existing server(s)? [y/N]: ")
				response, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("error reading input: %w", err)
				}

				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("\nOperation cancelled.")
					return nil
				}
				fmt.Println()
			}
		}
	}

	// Build YAML config
	yamlConfig := models.YAMLConfig{
		Projects: map[string]models.ProjectServers{
			projectPath: {
				MCPServers: servers,
			},
		},
	}

	// Convert to YAML
	yamlData, err := yaml.Marshal(&yamlConfig)
	if err != nil {
		return fmt.Errorf("error converting to YAML: %w", err)
	}

	fmt.Println("Converted YAML:")
	fmt.Println("─────────────────────────────────────")
	fmt.Print(string(yamlData))
	fmt.Println("─────────────────────────────────────")

	// Ask if user wants to apply
	fmt.Print("\nDo you want to add this configuration to ~/.claude.json? [y/N]: ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response == "y" || response == "yes" {
		// Create a temporary YAML file
		tmpFile := "/tmp/mcp-convert-temp.yaml"
		if err := os.WriteFile(tmpFile, yamlData, 0644); err != nil {
			return fmt.Errorf("error creating temporary file: %w", err)
		}
		defer func() { _ = os.Remove(tmpFile) }() //nolint:errcheck // Cleanup, error not critical

		// Apply the configuration
		if err := Apply(tmpFile); err != nil {
			return fmt.Errorf("error applying configuration: %w", err)
		}

		fmt.Println("\n✓ Configuration successfully applied!")
	} else {
		fmt.Println("\nConfiguration not applied.")
	}

	return nil
}
