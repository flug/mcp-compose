package models

// MCPServer represents an MCP server configuration
type MCPServer struct {
	Type    string            `json:"type,omitempty" yaml:"type,omitempty"`
	Command string            `json:"command,omitempty" yaml:"command,omitempty"`
	Args    []string          `json:"args,omitempty" yaml:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	URL     string            `json:"url,omitempty" yaml:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Scope   string            `json:"scope,omitempty" yaml:"scope,omitempty"` // "local", "user", "project"
}

// ProjectConfig represents configuration for a single project
type ProjectConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers" yaml:"mcpServers"`
}

// ClaudeConfig represents the complete ~/.claude.json structure
type ClaudeConfig struct {
	Projects map[string]ProjectConfig `json:"projects"`
}

// YAMLConfig represents the YAML configuration for MCP servers
// This is organized by project path
type YAMLConfig struct {
	Projects map[string]ProjectServers `yaml:"projects"`
}

// ProjectServers contains the MCP servers for a project
type ProjectServers struct {
	MCPServers map[string]MCPServer `yaml:"mcpServers"`
}

// ConvertInput represents input JSON for convert command
// Can be a single server or a map of servers
type ConvertInput struct {
	// For single server format
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Scope   string            `json:"scope,omitempty"`

	// For mcpServers map format
	MCPServers map[string]MCPServer `json:"mcpServers,omitempty"`
}
