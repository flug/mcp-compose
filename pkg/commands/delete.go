package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/flug/mcp-compose/pkg/config"
	"github.com/flug/mcp-compose/pkg/models"
	"gopkg.in/yaml.v3"
)

// Delete removes MCP servers from a YAML file and optionally from ~/.claude.json
func Delete(yamlFile, serverName, projectPath string) error {
	// Check if initialized
	if err := EnsureInitialized(); err != nil {
		return err
	}

	// Read YAML file
	yamlData, err := os.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %w", err)
	}

	var yamlConfig models.YAMLConfig
	if err := yaml.Unmarshal(yamlData, &yamlConfig); err != nil {
		return fmt.Errorf("error parsing YAML: %w", err)
	}

	// Check if project exists in YAML
	projectConfig, projectExists := yamlConfig.Projects[projectPath]
	if !projectExists {
		return fmt.Errorf("project '%s' not found in YAML file", projectPath)
	}

	// Check if server exists in project
	if _, serverExists := projectConfig.MCPServers[serverName]; !serverExists {
		return fmt.Errorf("MCP server '%s' not found in project '%s'", serverName, projectPath)
	}

	// Show what will be deleted
	fmt.Printf("MCP server to delete:\n")
	fmt.Printf("  Project: %s\n", projectPath)
	fmt.Printf("  Server:  %s\n\n", serverName)

	// Display server configuration
	serverConfig := projectConfig.MCPServers[serverName]
	serverYAML, _ := yaml.Marshal(map[string]models.MCPServer{serverName: serverConfig})
	fmt.Println("Configuration:")
	fmt.Println("─────────────────────────────────────")
	fmt.Print(string(serverYAML))
	fmt.Println("─────────────────────────────────────")

	// Create reader for user input
	reader := bufio.NewReader(os.Stdin)

	// Check if server exists in ~/.claude.json
	claudeConfig, err := config.ReadClaudeConfig()
	existsInClaudeJson := false
	if err == nil && claudeConfig.Projects != nil {
		if claudeProject, exists := claudeConfig.Projects[projectPath]; exists {
			if _, exists := claudeProject.MCPServers[serverName]; exists {
				existsInClaudeJson = true
			}
		}
	}

	deleteFromClaudeJson := false
	if existsInClaudeJson {
		fmt.Printf("\n⚠ This server also exists in ~/.claude.json\n")
		fmt.Print("Do you want to delete it from ~/.claude.json as well? [y/N]: ")
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		deleteFromClaudeJson = (response == "y" || response == "yes")
	}

	// Confirm deletion from YAML
	fmt.Print("\nDo you want to delete this server from the YAML file? [y/N]: ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		fmt.Println("\nOperation cancelled.")
		return nil
	}

	// Delete from YAML
	delete(projectConfig.MCPServers, serverName)

	// If project has no more servers, remove the project
	if len(projectConfig.MCPServers) == 0 {
		delete(yamlConfig.Projects, projectPath)
		fmt.Printf("\n✓ Server '%s' deleted from YAML\n", serverName)
		fmt.Printf("✓ Project '%s' removed from YAML (no more servers)\n", projectPath)
	} else {
		yamlConfig.Projects[projectPath] = projectConfig
		fmt.Printf("\n✓ Server '%s' deleted from YAML\n", serverName)
	}

	// Write updated YAML back to file
	updatedYAML, err := yaml.Marshal(&yamlConfig)
	if err != nil {
		return fmt.Errorf("error marshaling updated YAML: %w", err)
	}

	if err := os.WriteFile(yamlFile, updatedYAML, 0644); err != nil {
		return fmt.Errorf("error writing YAML file: %w", err)
	}

	// Delete from ~/.claude.json if requested
	if deleteFromClaudeJson {
		if claudeConfig.Projects != nil {
			if claudeProject, exists := claudeConfig.Projects[projectPath]; exists {
				delete(claudeProject.MCPServers, serverName)
				claudeConfig.Projects[projectPath] = claudeProject

				if err := config.WriteClaudeConfig(claudeConfig); err != nil {
					return fmt.Errorf("error updating ~/.claude.json: %w", err)
				}

				configPath, _ := config.GetClaudeConfigPath()
				fmt.Printf("✓ Server '%s' deleted from %s\n", serverName, configPath)
			}
		}
	}

	fmt.Println("\n✓ Deletion completed successfully!")

	return nil
}
