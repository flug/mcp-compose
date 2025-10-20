package marketplace

// MCPServerEntry represents an MCP server available in the marketplace
type MCPServerEntry struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Category    string `json:"category" yaml:"category"` // "reference", "official", "community"
	URL         string `json:"url" yaml:"url"`
	// Installation info derived from package.json or pyproject.toml
	PackageManager string   `json:"package_manager,omitempty" yaml:"package_manager,omitempty"` // "npm", "pip", "uvx", etc.
	PackageName    string   `json:"package_name,omitempty" yaml:"package_name,omitempty"`
	Command        string   `json:"command,omitempty" yaml:"command,omitempty"`
	Args           []string `json:"args,omitempty" yaml:"args,omitempty"`
}

// MarketplaceCache represents the cached marketplace data
type MarketplaceCache struct {
	LastUpdated string           `json:"last_updated" yaml:"last_updated"`
	Servers     []MCPServerEntry `json:"servers" yaml:"servers"`
}
