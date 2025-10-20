package platform

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// Info contains platform and user information
type Info struct {
	OS              string
	Arch            string
	Username        string
	HomeDir         string
	ClaudeConfigDir string
	ClaudeConfig    string
	MCPComposeConfig string
}

// Detect returns platform and user information
func Detect() (*Info, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %w", err)
	}

	info := &Info{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Username: currentUser.Username,
		HomeDir:  homeDir,
	}

	// Determine Claude config paths based on platform
	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\Claude\claude_desktop_config.json
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		info.ClaudeConfigDir = filepath.Join(appData, "Claude")
		info.ClaudeConfig = filepath.Join(info.ClaudeConfigDir, "claude_desktop_config.json")

		// mcp-compose config in same directory
		info.MCPComposeConfig = filepath.Join(info.ClaudeConfigDir, "mcp-compose.json")

	case "darwin":
		// macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
		info.ClaudeConfigDir = filepath.Join(homeDir, "Library", "Application Support", "Claude")
		info.ClaudeConfig = filepath.Join(info.ClaudeConfigDir, "claude_desktop_config.json")

		// mcp-compose config
		info.MCPComposeConfig = filepath.Join(homeDir, ".config", "mcp-compose", "config.json")

	case "linux":
		// Linux: ~/.config/Claude/claude_desktop_config.json or ~/.claude.json (fallback)
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(homeDir, ".config")
		}

		info.ClaudeConfigDir = filepath.Join(configDir, "Claude")
		claudeConfigPath := filepath.Join(info.ClaudeConfigDir, "claude_desktop_config.json")

		// Check if the standard config exists, otherwise use ~/.claude.json
		if _, err := os.Stat(claudeConfigPath); os.IsNotExist(err) {
			fallbackPath := filepath.Join(homeDir, ".claude.json")
			if _, err := os.Stat(fallbackPath); err == nil {
				info.ClaudeConfig = fallbackPath
				info.ClaudeConfigDir = homeDir
			} else {
				// Use standard path even if it doesn't exist yet
				info.ClaudeConfig = claudeConfigPath
			}
		} else {
			info.ClaudeConfig = claudeConfigPath
		}

		// mcp-compose config
		info.MCPComposeConfig = filepath.Join(configDir, "mcp-compose", "config.json")

	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return info, nil
}

// GetOSName returns a friendly OS name
func (i *Info) GetOSName() string {
	switch i.OS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return i.OS
	}
}

// ClaudeConfigExists checks if Claude config file exists
func (i *Info) ClaudeConfigExists() bool {
	_, err := os.Stat(i.ClaudeConfig)
	return err == nil
}

// MCPComposeConfigExists checks if mcp-compose config file exists
func (i *Info) MCPComposeConfigExists() bool {
	_, err := os.Stat(i.MCPComposeConfig)
	return err == nil
}

// CreateMCPComposeConfigDir creates the mcp-compose config directory if it doesn't exist
func (i *Info) CreateMCPComposeConfigDir() error {
	dir := filepath.Dir(i.MCPComposeConfig)
	return os.MkdirAll(dir, 0755)
}
