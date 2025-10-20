package models

// MCPComposeConfig represents the mcp-compose configuration
type MCPComposeConfig struct {
	Version          string `json:"version"`
	ClaudeConfigPath string `json:"claude_config_path"`
	DefaultProject   string `json:"default_project,omitempty"`
	AutoBackup       bool   `json:"auto_backup"`
	Platform         string `json:"platform"`
	User             string `json:"user"`
}
