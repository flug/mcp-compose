package marketplace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	githubRepoURL = "https://github.com/modelcontextprotocol/servers.git"
	cacheDir      = ".config/mcp-compose/cache"
	cacheFile     = "marketplace.yaml"
)

// FetchServers downloads and parses the MCP servers repository
func FetchServers(forceRefresh bool) (*MarketplaceCache, error) {
	// Check cache first
	cachePath := getCachePath()
	if !forceRefresh {
		if cache, err := loadCache(cachePath); err == nil {
			return cache, nil
		}
	}

	// Clone or update repository
	repoPath, err := cloneOrUpdateRepo()
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Parse servers from repository
	servers, err := parseServers(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse servers: %w", err)
	}

	// Create cache
	cache := &MarketplaceCache{
		LastUpdated: time.Now().Format(time.RFC3339),
		Servers:     servers,
	}

	// Save cache
	if err := saveCache(cachePath, cache); err != nil {
		return nil, fmt.Errorf("failed to save cache: %w", err)
	}

	return cache, nil
}

// getCachePath returns the full path to the cache file
func getCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir cannot be determined
		home = "."
	}
	return filepath.Join(home, cacheDir, cacheFile)
}

// loadCache loads the marketplace cache from disk
func loadCache(path string) (*MarketplaceCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache MarketplaceCache
	if err := yaml.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// saveCache saves the marketplace cache to disk
func saveCache(path string, cache *MarketplaceCache) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// cloneOrUpdateRepo clones or updates the MCP servers repository
func cloneOrUpdateRepo() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	repoPath := filepath.Join(home, cacheDir, "mcp-servers")

	// Check if repo already exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		// Repository exists, pull latest changes
		cmd := exec.Command("git", "-C", repoPath, "pull", "--depth", "1")
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to update repository: %w", err)
		}
		return repoPath, nil
	}

	// Clone repository
	if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
		return "", err
	}

	cmd := exec.Command("git", "clone", "--depth", "1", githubRepoURL, repoPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return repoPath, nil
}

// parseServers parses MCP servers from the repository
func parseServers(repoPath string) ([]MCPServerEntry, error) {
	var servers []MCPServerEntry

	// Parse reference servers from README
	readmePath := filepath.Join(repoPath, "README.md")
	referenceServers, err := parseREADME(readmePath, "reference")
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference servers: %w", err)
	}
	servers = append(servers, referenceServers...)

	// Parse official third-party servers from README
	officialServers, err := parseREADME(readmePath, "official")
	if err != nil {
		return nil, fmt.Errorf("failed to parse official servers: %w", err)
	}
	servers = append(servers, officialServers...)

	// Enhance reference servers with installation info from src/
	srcPath := filepath.Join(repoPath, "src")
	if err := enhanceWithInstallInfo(servers, srcPath); err != nil {
		return nil, fmt.Errorf("failed to enhance with install info: %w", err)
	}

	return servers, nil
}

// parseREADME parses server entries from the README file
func parseREADME(path string, category string) ([]MCPServerEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var servers []MCPServerEntry
	scanner := bufio.NewScanner(file)
	inSection := false
	sectionHeader := ""

	if category == "reference" {
		sectionHeader = "## 🌟 Reference Servers"
	} else if category == "official" {
		sectionHeader = "### 🎖️ Official Integrations"
	}

	// Regex to match server entries like:
	// - **[Name](url)** - Description
	// - <img ...> **[Name](url)** - Description
	entryRegex := regexp.MustCompile(`\*\*\[([^\]]+)\]\(([^\)]+)\)\*\*\s*[-–]\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if we entered the target section
		if strings.Contains(line, sectionHeader) {
			inSection = true
			continue
		}

		// Check if we left the section (next ## or ### header)
		if inSection && strings.HasPrefix(line, "##") && !strings.Contains(line, sectionHeader) {
			break
		}
		if inSection && category == "official" && strings.HasPrefix(line, "###") && !strings.Contains(line, sectionHeader) {
			break
		}

		// Parse server entry
		if inSection && strings.HasPrefix(strings.TrimSpace(line), "-") {
			matches := entryRegex.FindStringSubmatch(line)
			if len(matches) == 4 {
				servers = append(servers, MCPServerEntry{
					Name:        matches[1],
					URL:         matches[2],
					Description: strings.TrimSpace(matches[3]),
					Category:    category,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

// enhanceWithInstallInfo adds installation information from package files
func enhanceWithInstallInfo(servers []MCPServerEntry, srcPath string) error {
	for i := range servers {
		server := &servers[i]

		// Only process reference servers with local paths
		if server.Category != "reference" || !strings.HasPrefix(server.URL, "src/") {
			continue
		}

		// Extract directory name from URL
		dirName := strings.TrimPrefix(server.URL, "src/")
		serverPath := filepath.Join(srcPath, dirName)

		// Check for package.json (Node.js/TypeScript)
		packageJSONPath := filepath.Join(serverPath, "package.json")
		if _, err := os.Stat(packageJSONPath); err == nil {
			if err := parsePackageJSON(server, packageJSONPath); err == nil {
				continue
			}
		}

		// Check for pyproject.toml (Python)
		pyprojectPath := filepath.Join(serverPath, "pyproject.toml")
		if _, err := os.Stat(pyprojectPath); err == nil {
			if err := parsePyProject(server, pyprojectPath); err == nil {
				continue
			}
		}
	}

	return nil
}

// parsePackageJSON extracts installation info from package.json
func parsePackageJSON(server *MCPServerEntry, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var pkg struct {
		Name string            `json:"name"`
		Bin  map[string]string `json:"bin"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
	}

	if pkg.Name != "" {
		server.PackageManager = "npx"
		server.PackageName = pkg.Name
		server.Command = "npx"
		server.Args = []string{"-y", pkg.Name}
	}

	return nil
}

// parsePyProject extracts installation info from pyproject.toml
func parsePyProject(server *MCPServerEntry, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Simple TOML parsing for name field
	// Look for name = "value" in [project] section
	lines := strings.Split(string(data), "\n")
	inProject := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[project]" {
			inProject = true
			continue
		}

		if inProject && strings.HasPrefix(line, "[") {
			break
		}

		if inProject && strings.HasPrefix(line, "name = ") {
			// Extract name value
			name := strings.TrimPrefix(line, "name = ")
			name = strings.Trim(name, `"`)

			server.PackageManager = "uvx"
			server.PackageName = name
			server.Command = "uvx"
			server.Args = []string{name}
			break
		}
	}

	return nil
}
