# MCP Compose

[![Test](https://github.com/flug/mcp-compose/workflows/Test/badge.svg)](https://github.com/flug/mcp-compose/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/flug/mcp-compose)](https://goreportcard.com/report/github.com/flug/mcp-compose)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/flug/mcp-compose)](https://go.dev/)

A CLI tool to manage MCP (Model Context Protocol) servers configuration for Claude Desktop via YAML files.

> **⚠️ EXPERIMENTAL - Use with Caution**
>
> This tool is in active development and should be considered experimental. While it includes safety features and validation, please:
> - **Backup your Claude configuration** before using this tool
> - Test on non-critical projects first
> - Report any issues on GitHub
> - The tool modifies `~/.claude.json` or platform-specific Claude Desktop configuration files
>
> **Always keep a backup of your configuration files!**

## Installation

```bash
go install github.com/flug/mcp-compose@latest
```

Or build from source:

```bash
git clone https://github.com/flug/mcp-compose
cd mcp-compose
go build
```

## Quick Start

1. **Initialize mcp-compose:**
   ```bash
   mcp-compose init
   ```
   This will detect your platform, locate your Claude Desktop configuration, and set up mcp-compose.

2. **Browse and install MCP servers from the marketplace:**
   ```bash
   # List all available servers
   mcp-compose marketplace list

   # Search for specific servers
   mcp-compose marketplace search filesystem

   # Install a server
   mcp-compose marketplace install filesystem /path/to/your/project
   ```

3. **Or manage servers via YAML:**
   ```bash
   # Export your current MCP servers
   mcp-compose dump > mcp-servers.yaml

   # Edit the YAML file
   vim mcp-servers.yaml

   # Apply changes
   mcp-compose apply mcp-servers.yaml
   ```

## Usage

### Init Command

Initialize mcp-compose configuration. This detects your operating system, user, and Claude Desktop configuration paths.

```bash
mcp-compose init
```

The init command:
- Detects your platform (Windows, macOS, or Linux)
- Locates your Claude Desktop configuration file
- Creates mcp-compose configuration file
- Displays system information

**Platform-specific paths:**
- **Linux**: `~/.config/Claude/claude_desktop_config.json` or `~/.claude.json`
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

### Dump Command

Extract MCP servers from Claude Desktop configuration and output as YAML:

```bash
mcp-compose dump
```

Save to a file:

```bash
mcp-compose dump > mcp-config.yaml
```

### Apply Command

Update `~/.claude.json` with MCP servers from a YAML file:

```bash
mcp-compose apply mcp-config.yaml
```

### Convert Command

Convert a JSON MCP server configuration to YAML format, validate it, and optionally apply it to `~/.claude.json`:

```bash
mcp-compose convert server.json /path/to/project
```

The convert command supports two JSON formats:

1. **Single server format:**
```json
{
  "type": "stdio",
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
  "env": {
    "DEBUG": "true"
  },
  "scope": "project"
}
```

2. **Multiple servers (mcpServers map):**
```json
{
  "mcpServers": {
    "filesystem": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"]
    },
    "github": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "ghp_example"
      }
    }
  }
}
```

The command will:
1. Validate the JSON configuration
2. Convert it to YAML format
3. Display the converted configuration
4. Ask if you want to apply it to `~/.claude.json`

### Delete Command

Delete an MCP server from a YAML configuration file and optionally from `~/.claude.json`:

```bash
mcp-compose delete mcp-config.yaml server-name /path/to/project
```

Example:
```bash
mcp-compose delete mcp-config.yaml filesystem /home/user/workspace/my-project
```

The command will:
1. Display the server configuration to be deleted
2. Ask if you want to delete it from `~/.claude.json` (if it exists there)
3. Ask for confirmation before deleting from the YAML file
4. Remove the server from the YAML file
5. Optionally remove it from `~/.claude.json`

## Marketplace Commands

The marketplace commands allow you to browse and install MCP servers from the official [Model Context Protocol servers repository](https://github.com/modelcontextprotocol/servers).

### Marketplace List

List all available MCP servers from the marketplace:

```bash
mcp-compose marketplace list
```

This command will:
- Fetch the latest list of MCP servers from the official repository
- Cache the results locally for faster subsequent access
- Display reference servers and official integrations
- Show installation commands for each server

### Marketplace Search

Search for MCP servers by name or description:

```bash
mcp-compose marketplace search <query>
```

Examples:
```bash
# Search for filesystem-related servers
mcp-compose marketplace search filesystem

# Search for Git-related servers
mcp-compose marketplace search git

# Search for database servers
mcp-compose marketplace search database
```

The search is case-insensitive and matches against both server names and descriptions.

### Marketplace Install

Install an MCP server from the marketplace:

```bash
mcp-compose marketplace install <server-name> <project-path>
```

Example:
```bash
mcp-compose marketplace install filesystem /home/user/workspace/my-project
```

The install command will:
1. Search for the server in the marketplace
2. Display server information and installation details
3. Generate the MCP server configuration
4. Ask for confirmation before applying changes
5. Add the server to your Claude Desktop configuration for the specified project

**Note:** You may need to restart Claude Desktop for changes to take effect.

## YAML Configuration Format

The configuration is organized by project path. Each project can have its own MCP servers.

```yaml
projects:
  /path/to/project:
    mcpServers:
      server-name:
        type: "stdio"      # or "http"
        command: "command-to-run"
        args:
          - "arg1"
          - "arg2"
        env:
          ENV_VAR: "value"
        scope: "project"   # optional: "local", "user", or "project"

      http-server-name:
        type: "http"
        url: "https://api.example.com/mcp"
        headers:
          Authorization: "Bearer token"
```

### Example

```yaml
projects:
  /home/user/workspace/my-project:
    mcpServers:
      filesystem:
        type: stdio
        command: npx
        args:
          - "-y"
          - "@modelcontextprotocol/server-filesystem"
          - "/path/to/allowed/directory"

      github:
        type: stdio
        command: npx
        args:
          - "-y"
          - "@modelcontextprotocol/server-github"
        env:
          GITHUB_TOKEN: "your-token-here"

      custom-api:
        type: http
        url: "https://api.example.com/mcp"
        headers:
          Authorization: "Bearer your-token"
```

## Workflows

### Installing from Marketplace

1. Search for available servers:
   ```bash
   mcp-compose marketplace search <query>
   ```

2. Install the desired server:
   ```bash
   mcp-compose marketplace install <server-name> /path/to/your/project
   ```

3. Restart Claude Desktop to apply changes

### Managing Existing Configuration

1. Export current configuration:
   ```bash
   mcp-compose dump > mcp-config.yaml
   ```

2. Edit the YAML file with your preferred editor

3. Apply changes:
   ```bash
   mcp-compose apply mcp-config.yaml
   ```

### Adding New MCP Server from JSON

1. Get or create a JSON configuration file for your MCP server

2. Convert and validate:
   ```bash
   mcp-compose convert server.json /path/to/your/project
   ```

3. Review the converted YAML and confirm to apply

### Removing MCP Server

1. Export current configuration if you don't have a YAML file:
   ```bash
   mcp-compose dump > mcp-config.yaml
   ```

2. Delete the server:
   ```bash
   mcp-compose delete mcp-config.yaml server-name /path/to/project
   ```

3. Optionally confirm deletion from `~/.claude.json`

## Project Structure

```
mcp-compose/
├── main.go              # CLI entry point
├── pkg/
│   ├── models/          # Data structures
│   │   └── types.go
│   ├── config/          # Config file operations
│   │   └── config.go
│   └── commands/        # CLI commands
│       ├── dump.go      # Dump command
│       ├── apply.go     # Apply command
│       ├── convert.go   # Convert command
│       ├── delete.go    # Delete command
│       └── validate.go  # Validation logic
```

## Development

Build:
```bash
go build
```

Run:
```bash
go run main.go dump
go run main.go apply config.yaml
go run main.go convert server.json /path/to/project
go run main.go delete config.yaml server-name /path/to/project
```

Test:
```bash
go test ./...
```

## Configuration Fields

### MCP Server Types

- **stdio**: Command-based MCP server (requires `command`, optional `args` and `env`)
- **http**: HTTP-based MCP server (requires `url`, optional `headers`)

### Scope Values

- **local**: Server is local to the current machine
- **user**: Server is available to the current user
- **project**: Server is specific to the project

## Notes

### Safety Features

- The tool preserves all other fields in Claude config, only modifying the `mcpServers` section of each project
- Interactive prompts prevent accidental modifications
- Validation checks configuration before applying changes
- Init command required before first use

### Best Practices

- **Always backup your configuration** before making changes:
  ```bash
  # Linux/macOS
  cp ~/.claude.json ~/.claude.json.backup
  # or if using standard location
  cp ~/.config/Claude/claude_desktop_config.json ~/.config/Claude/claude_desktop_config.json.backup

  # Windows (PowerShell)
  Copy-Item "$env:APPDATA\Claude\claude_desktop_config.json" "$env:APPDATA\Claude\claude_desktop_config.json.backup"
  ```
- Test changes on a non-critical project first
- Use `dump` to export current configuration before making changes
- Review YAML output from `convert` before confirming

### Technical Details

- MCP servers are organized per project in Claude Desktop
- Both `stdio` and `http` type MCP servers are supported
- Configuration stored in user's home directory, not with the binary
- Multi-platform support: Windows, macOS, Linux
