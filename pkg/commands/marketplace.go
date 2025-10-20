package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/flug/mcp-compose/pkg/config"
	"github.com/flug/mcp-compose/pkg/marketplace"
	"github.com/flug/mcp-compose/pkg/models"
	"gopkg.in/yaml.v3"
)

// MarketplaceList lists all available MCP servers from the marketplace
func MarketplaceList() error {
	if err := EnsureInitialized(); err != nil {
		return err
	}

	fmt.Println("Fetching MCP servers from marketplace...")
	cache, err := marketplace.FetchServers(false)
	if err != nil {
		return fmt.Errorf("failed to fetch servers: %w", err)
	}

	fmt.Printf("\n=== MCP Marketplace ===\n")
	fmt.Printf("Last updated: %s\n", cache.LastUpdated)
	fmt.Printf("Total servers: %d\n\n", len(cache.Servers))

	// Group by category
	categories := map[string][]marketplace.MCPServerEntry{
		"reference": {},
		"official":  {},
		"community": {},
	}

	for _, server := range cache.Servers {
		categories[server.Category] = append(categories[server.Category], server)
	}

	// Display reference servers
	if len(categories["reference"]) > 0 {
		fmt.Println("📦 Reference Servers:")
		for _, server := range categories["reference"] {
			fmt.Printf("  • %s - %s\n", server.Name, server.Description)
			if server.PackageName != "" {
				fmt.Printf("    Install: %s %s\n", server.Command, strings.Join(server.Args, " "))
			}
		}
		fmt.Println()
	}

	// Display official servers
	if len(categories["official"]) > 0 {
		fmt.Printf("🎖️  Official Integrations (%d servers):\n", len(categories["official"]))
		for i, server := range categories["official"] {
			if i >= 10 {
				fmt.Printf("  ... and %d more (use 'marketplace search' to find specific servers)\n", len(categories["official"])-10)
				break
			}
			fmt.Printf("  • %s - %s\n", server.Name, server.Description)
		}
		fmt.Println()
	}

	fmt.Println("Use 'mcp-compose marketplace search <query>' to search for specific servers")
	fmt.Println("Use 'mcp-compose marketplace install <name>' to install a server")

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

	// Create MCP server configuration
	serverKey := strings.ToLower(strings.ReplaceAll(found.Name, " ", "-"))

	mcpServer := models.MCPServer{
		Type:    "stdio",
		Command: found.Command,
		Args:    found.Args,
		Scope:   "project",
	}

	// Create YAML structure
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

	fmt.Printf("\n✅ Successfully installed '%s' for project: %s\n", found.Name, projectPath)
	fmt.Println("\nNote: You may need to restart Claude Desktop for changes to take effect.")

	return nil
}
