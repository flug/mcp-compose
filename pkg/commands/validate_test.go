package commands

import (
	"testing"

	"github.com/flug/mcp-compose/pkg/models"
)

func TestValidateMCPServer(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		server     models.MCPServer
		wantErr    bool
	}{
		{
			name:       "valid stdio server",
			serverName: "test-server",
			server: models.MCPServer{
				Type:    "stdio",
				Command: "npx",
				Args:    []string{"-y", "test-package"},
			},
			wantErr: false,
		},
		{
			name:       "valid http server",
			serverName: "test-api",
			server: models.MCPServer{
				Type: "http",
				URL:  "https://api.example.com/mcp",
			},
			wantErr: false,
		},
		{
			name:       "invalid type",
			serverName: "test-server",
			server: models.MCPServer{
				Type:    "invalid",
				Command: "npx",
			},
			wantErr: true,
		},
		{
			name:       "stdio without command",
			serverName: "test-server",
			server: models.MCPServer{
				Type: "stdio",
			},
			wantErr: true,
		},
		{
			name:       "http without URL",
			serverName: "test-server",
			server: models.MCPServer{
				Type: "http",
			},
			wantErr: true,
		},
		{
			name:       "invalid scope",
			serverName: "test-server",
			server: models.MCPServer{
				Type:    "stdio",
				Command: "npx",
				Scope:   "invalid-scope",
			},
			wantErr: true,
		},
		{
			name:       "valid scope",
			serverName: "test-server",
			server: models.MCPServer{
				Type:    "stdio",
				Command: "npx",
				Scope:   "project",
			},
			wantErr: false,
		},
		{
			name:       "stdio with URL should fail",
			serverName: "test-server",
			server: models.MCPServer{
				Type:    "stdio",
				Command: "npx",
				URL:     "https://example.com",
			},
			wantErr: true,
		},
		{
			name:       "http with command should fail",
			serverName: "test-server",
			server: models.MCPServer{
				Type:    "http",
				URL:     "https://example.com",
				Command: "npx",
			},
			wantErr: true,
		},
		{
			name:       "empty server name",
			serverName: "",
			server: models.MCPServer{
				Type:    "stdio",
				Command: "npx",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMCPServer(tt.serverName, tt.server)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMCPServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateYAMLConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *models.YAMLConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &models.YAMLConfig{
				Projects: map[string]models.ProjectServers{
					"/test/project": {
						MCPServers: map[string]models.MCPServer{
							"server1": {
								Type:    "stdio",
								Command: "npx",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty projects",
			config: &models.YAMLConfig{
				Projects: map[string]models.ProjectServers{},
			},
			wantErr: true,
		},
		{
			name: "project with no servers",
			config: &models.YAMLConfig{
				Projects: map[string]models.ProjectServers{
					"/test/project": {
						MCPServers: map[string]models.MCPServer{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid server in config",
			config: &models.YAMLConfig{
				Projects: map[string]models.ProjectServers{
					"/test/project": {
						MCPServers: map[string]models.MCPServer{
							"server1": {
								Type: "invalid",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateYAMLConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateYAMLConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
