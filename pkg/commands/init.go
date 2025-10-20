package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/flug/mcp-compose/pkg/models"
	"github.com/flug/mcp-compose/pkg/platform"
)

const version = "1.0.0"

// Init initializes mcp-compose configuration
func Init() error {
	fmt.Println("🚀 Initializing MCP Compose...")
	fmt.Println()

	// Detect platform and user info
	platformInfo, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("error detecting platform: %w", err)
	}

	// Display platform information
	fmt.Printf("Platform Information:\n")
	fmt.Printf("  OS:           %s (%s)\n", platformInfo.GetOSName(), platformInfo.Arch)
	fmt.Printf("  User:         %s\n", platformInfo.Username)
	fmt.Printf("  Home:         %s\n", platformInfo.HomeDir)
	fmt.Println()

	// Display Claude configuration paths
	fmt.Printf("Claude Desktop Configuration:\n")
	fmt.Printf("  Config Dir:   %s\n", platformInfo.ClaudeConfigDir)
	fmt.Printf("  Config File:  %s\n", platformInfo.ClaudeConfig)

	if platformInfo.ClaudeConfigExists() {
		fmt.Printf("  Status:       ✓ Found\n")
	} else {
		fmt.Printf("  Status:       ⚠ Not found (will be created when needed)\n")
	}
	fmt.Println()

	// Check if mcp-compose config already exists
	if platformInfo.MCPComposeConfigExists() {
		fmt.Printf("MCP Compose configuration already exists at:\n")
		fmt.Printf("  %s\n", platformInfo.MCPComposeConfig)
		fmt.Println()

		// Read existing config
		data, err := os.ReadFile(platformInfo.MCPComposeConfig)
		if err == nil {
			var existingConfig models.MCPComposeConfig
			if err := json.Unmarshal(data, &existingConfig); err == nil {
				fmt.Printf("Current configuration:\n")
				fmt.Printf("  Version:      %s\n", existingConfig.Version)
				fmt.Printf("  Auto Backup:  %v\n", existingConfig.AutoBackup)
				if existingConfig.DefaultProject != "" {
					fmt.Printf("  Default Project: %s\n", existingConfig.DefaultProject)
				}
				fmt.Println()
			}
		}

		fmt.Print("Do you want to reinitialize? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("\nConfiguration unchanged.")
			return nil
		}
		fmt.Println()
	}

	// Create mcp-compose config
	config := models.MCPComposeConfig{
		Version:          version,
		ClaudeConfigPath: platformInfo.ClaudeConfig,
		AutoBackup:       true,
		Platform:         platformInfo.GetOSName(),
		User:             platformInfo.Username,
	}

	// Create config directory if needed
	if err := platformInfo.CreateMCPComposeConfigDir(); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Write config file
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(platformInfo.MCPComposeConfig, configData, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	fmt.Printf("✓ Configuration saved to:\n")
	fmt.Printf("  %s\n", platformInfo.MCPComposeConfig)
	fmt.Println()

	// Show configuration
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Version:          %s\n", config.Version)
	fmt.Printf("  Claude Config:    %s\n", config.ClaudeConfigPath)
	fmt.Printf("  Auto Backup:      %v\n", config.AutoBackup)
	fmt.Printf("  Platform:         %s\n", config.Platform)
	fmt.Printf("  User:             %s\n", config.User)
	fmt.Println()

	fmt.Println("✓ MCP Compose initialized successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Run 'mcp-compose dump' to export your current MCP servers")
	fmt.Println("  • Run 'mcp-compose --help' to see all available commands")

	return nil
}
