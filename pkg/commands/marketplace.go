package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/flug/mcp-compose/pkg/config"
	"github.com/flug/mcp-compose/pkg/marketplace"
	"github.com/flug/mcp-compose/pkg/models"
	"gopkg.in/yaml.v3"
)

// MarketplaceList lists all available MCP servers from the marketplace with interactive selection
func MarketplaceList() error {
	if err := EnsureInitialized(); err != nil {
		return err
	}

	fmt.Println("📦 Fetching MCP servers from marketplace...")
	fmt.Println("   Repository: https://github.com/modelcontextprotocol/servers")
	cache, err := marketplace.FetchServers(false)
	if err != nil {
		return fmt.Errorf("failed to fetch servers: %w", err)
	}

	fmt.Printf("✓ Found %d servers (last updated: %s)\n", len(cache.Servers), cache.LastUpdated)
	fmt.Println("\nLegend: ✓ = Auto-install available  |  ⚠ = Manual install required")

	// Prepare options for multi-select
	var options []string
	serverMap := make(map[string]marketplace.MCPServerEntry)

	for _, server := range cache.Servers {
		// Show all servers, mark which ones have auto-install support
		installStatus := "✓"
		if server.PackageName == "" {
			installStatus = "⚠" // Manual installation required
		}

		label := fmt.Sprintf("[%s] %s %s - %s", server.Category, installStatus, server.Name, server.Description)
		// Limit description length for better display
		if len(label) > 120 {
			label = label[:117] + "..."
		}
		options = append(options, label)
		serverMap[label] = server
	}

	if len(options) == 0 {
		return fmt.Errorf("no servers found in marketplace")
	}

	// Create multi-select prompt
	var selected []string
	prompt := &survey.MultiSelect{
		Message:  "Select MCP servers to install (use space to select, enter to confirm):",
		Options:  options,
		PageSize: 15,
	}

	err = survey.AskOne(prompt, &selected, survey.WithKeepFilter(true))
	if err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if len(selected) == 0 {
		fmt.Println("No servers selected. Exiting.")
		return nil
	}

	// Ask for project path
	var projectPath string
	projectPrompt := &survey.Input{
		Message: "Enter the project path where you want to install these servers:",
		Help:    "Example: /home/user/workspace/my-project",
	}
	err = survey.AskOne(projectPrompt, &projectPath, survey.WithValidator(survey.Required))
	if err != nil {
		return fmt.Errorf("project path input cancelled: %w", err)
	}

	// Install selected servers
	fmt.Printf("\n📥 Installing %d server(s) to project: %s\n\n", len(selected), projectPath)

	successCount := 0
	manualCount := 0

	for i, label := range selected {
		server := serverMap[label]
		fmt.Printf("[%d/%d] %s...\n", i+1, len(selected), server.Name)

		if server.PackageName == "" {
			// Manual installation required
			fmt.Printf("  ⚠ Manual installation required\n")
			fmt.Printf("  → Visit: %s\n", server.URL)
			manualCount++
		} else {
			// Automatic installation
			if err := installServer(&server, projectPath); err != nil {
				fmt.Printf("  ✗ Failed: %v\n", err)
			} else {
				fmt.Printf("  ✓ Installed successfully\n")
				successCount++
			}
		}
	}

	fmt.Printf("\n✅ Installation complete!\n")
	fmt.Printf("   - %d server(s) installed automatically\n", successCount)
	if manualCount > 0 {
		fmt.Printf("   - %d server(s) require manual installation (visit URLs above)\n", manualCount)
	}
	fmt.Println("\nNote: You may need to restart Claude Desktop for changes to take effect.")

	return nil
}

// MarketplaceSearch searches for MCP servers in the marketplace
func MarketplaceSearch(query string) error {
	if err := EnsureInitialized(); err != nil {
		return err
	}

	if query == "" {
		return fmt.Errorf("search query is required")
	}

	cache, err := marketplace.FetchServers(false)
	if err != nil {
		return fmt.Errorf("failed to fetch servers: %w", err)
	}

	query = strings.ToLower(query)
	var matches []marketplace.MCPServerEntry

	for _, server := range cache.Servers {
		nameLower := strings.ToLower(server.Name)
		descLower := strings.ToLower(server.Description)

		if strings.Contains(nameLower, query) || strings.Contains(descLower, query) {
			matches = append(matches, server)
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No servers found matching '%s'\n", query)
		return nil
	}

	fmt.Printf("\n=== Search Results for '%s' ===\n", query)
	fmt.Printf("Found %d server(s):\n\n", len(matches))

	for _, server := range matches {
		fmt.Printf("📦 %s\n", server.Name)
		fmt.Printf("   Description: %s\n", server.Description)
		fmt.Printf("   Category: %s\n", server.Category)
		if server.PackageName != "" {
			fmt.Printf("   Install: %s %s\n", server.Command, strings.Join(server.Args, " "))
		}
		fmt.Printf("   URL: %s\n", server.URL)
		fmt.Println()
	}

	return nil
}

// installServer is a helper function to install a single MCP server
func installServer(server *marketplace.MCPServerEntry, projectPath string) error {
	if server.PackageName == "" {
		return fmt.Errorf("no installation information available for '%s'. Please visit %s for manual installation instructions", server.Name, server.URL)
	}

	// Create MCP server configuration
	serverKey := strings.ToLower(strings.ReplaceAll(server.Name, " ", "-"))

	mcpServer := models.MCPServer{
		Type:    "stdio",
		Command: server.Command,
		Args:    server.Args,
		Scope:   "project",
	}

	// Apply configuration directly
	claudeConfig, err := config.ReadClaudeConfig()
	if err != nil {
		return fmt.Errorf("failed to read Claude config: %w", err)
	}

	// Initialize projects map if nil
	if claudeConfig.Projects == nil {
		claudeConfig.Projects = make(map[string]models.ProjectConfig)
	}

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

	// Add the new server
	projectConfig.MCPServers[serverKey] = mcpServer
	claudeConfig.Projects[projectPath] = projectConfig

	// Write back to file
	if err := config.WriteClaudeConfig(claudeConfig); err != nil {
		return fmt.Errorf("failed to write Claude config: %w", err)
	}

	return nil
}

// MarketplaceInstall installs an MCP server from the marketplace
func MarketplaceInstall(serverName, projectPath string) error {
	if err := EnsureInitialized(); err != nil {
		return err
	}

	if serverName == "" {
		return fmt.Errorf("server name is required")
	}

	if projectPath == "" {
		return fmt.Errorf("project path is required")
	}

	// Fetch servers
	cache, err := marketplace.FetchServers(false)
	if err != nil {
		return fmt.Errorf("failed to fetch servers: %w", err)
	}

	// Find server by name (case-insensitive)
	var found *marketplace.MCPServerEntry
	serverNameLower := strings.ToLower(serverName)
	for i, server := range cache.Servers {
		if strings.ToLower(server.Name) == serverNameLower {
			found = &cache.Servers[i]
			break
		}
	}

	if found == nil {
		return fmt.Errorf("server '%s' not found in marketplace", serverName)
	}

	// Display server info
	fmt.Printf("\n=== Installing MCP Server ===\n")
	fmt.Printf("Name: %s\n", found.Name)
	fmt.Printf("Description: %s\n", found.Description)
	fmt.Printf("Category: %s\n", found.Category)
	fmt.Printf("URL: %s\n", found.URL)

	if found.PackageName == "" {
		return fmt.Errorf("no installation information available for '%s'. Please visit %s for manual installation instructions", found.Name, found.URL)
	}

	fmt.Printf("\nInstallation command: %s %s\n", found.Command, strings.Join(found.Args, " "))
	fmt.Printf("Project: %s\n\n", projectPath)

	// Create MCP server configuration for preview
	serverKey := strings.ToLower(strings.ReplaceAll(found.Name, " ", "-"))

	mcpServer := models.MCPServer{
		Type:    "stdio",
		Command: found.Command,
		Args:    found.Args,
		Scope:   "project",
	}

	// Create YAML structure for preview
	yamlConfig := models.YAMLConfig{
		Projects: map[string]models.ProjectServers{
			projectPath: {
				MCPServers: map[string]models.MCPServer{
					serverKey: mcpServer,
				},
			},
		},
	}

	// Display configuration
	yamlData, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	fmt.Println("Generated configuration:")
	fmt.Println("---")
	fmt.Println(string(yamlData))
	fmt.Println("---")

	// Ask for confirmation
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nDo you want to add this server to your Claude configuration? [y/N]: ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Installation cancelled.")
		return nil
	}

	// Install using helper function
	if err := installServer(found, projectPath); err != nil {
		return err
	}

	fmt.Printf("\n✅ Successfully installed '%s' for project: %s\n", found.Name, projectPath)
	fmt.Println("\nNote: You may need to restart Claude Desktop for changes to take effect.")

	return nil
}
