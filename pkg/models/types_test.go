package models

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMCPServerJSONMarshaling(t *testing.T) {
	server := MCPServer{
		Type:    "stdio",
		Command: "npx",
		Args:    []string{"-y", "test-package"},
		Env: map[string]string{
			"DEBUG": "true",
		},
		Scope: "project",
	}

	// Marshal to JSON
	data, err := json.Marshal(server)
	if err != nil {
		t.Fatalf("Failed to marshal MCPServer: %v", err)
	}

	// Unmarshal back
	var unmarshaled MCPServer
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal MCPServer: %v", err)
	}

	// Verify fields
	if unmarshaled.Type != server.Type {
		t.Errorf("Type mismatch: got %v, want %v", unmarshaled.Type, server.Type)
	}
	if unmarshaled.Command != server.Command {
		t.Errorf("Command mismatch: got %v, want %v", unmarshaled.Command, server.Command)
	}
	if len(unmarshaled.Args) != len(server.Args) {
		t.Errorf("Args length mismatch: got %v, want %v", len(unmarshaled.Args), len(server.Args))
	}
	if unmarshaled.Scope != server.Scope {
		t.Errorf("Scope mismatch: got %v, want %v", unmarshaled.Scope, server.Scope)
	}
}

func TestMCPServerYAMLMarshaling(t *testing.T) {
	server := MCPServer{
		Type:    "http",
		URL:     "https://api.example.com/mcp",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(server)
	if err != nil {
		t.Fatalf("Failed to marshal MCPServer: %v", err)
	}

	// Unmarshal back
	var unmarshaled MCPServer
	if err := yaml.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal MCPServer: %v", err)
	}

	// Verify fields
	if unmarshaled.Type != server.Type {
		t.Errorf("Type mismatch: got %v, want %v", unmarshaled.Type, server.Type)
	}
	if unmarshaled.URL != server.URL {
		t.Errorf("URL mismatch: got %v, want %v", unmarshaled.URL, server.URL)
	}
	if len(unmarshaled.Headers) != len(server.Headers) {
		t.Errorf("Headers length mismatch: got %v, want %v", len(unmarshaled.Headers), len(server.Headers))
	}
}

func TestConvertInputJSONFormat(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
	}{
		{
			name: "single server format",
			jsonData: `{
				"type": "stdio",
				"command": "npx",
				"args": ["-y", "test"]
			}`,
			wantErr: false,
		},
		{
			name: "mcpServers map format",
			jsonData: `{
				"mcpServers": {
					"server1": {
						"type": "stdio",
						"command": "npx"
					}
				}
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input ConvertInput
			err := json.Unmarshal([]byte(tt.jsonData), &input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYAMLConfigStructure(t *testing.T) {
	config := YAMLConfig{
		Projects: map[string]ProjectServers{
			"/test/project": {
				MCPServers: map[string]MCPServer{
					"server1": {
						Type:    "stdio",
						Command: "npx",
					},
					"server2": {
						Type: "http",
						URL:  "https://api.example.com",
					},
				},
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal YAMLConfig: %v", err)
	}

	// Unmarshal back
	var unmarshaled YAMLConfig
	if err := yaml.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal YAMLConfig: %v", err)
	}

	// Verify structure
	if len(unmarshaled.Projects) != 1 {
		t.Errorf("Projects count mismatch: got %v, want 1", len(unmarshaled.Projects))
	}

	project, exists := unmarshaled.Projects["/test/project"]
	if !exists {
		t.Error("Project '/test/project' not found after unmarshal")
	}

	if len(project.MCPServers) != 2 {
		t.Errorf("MCPServers count mismatch: got %v, want 2", len(project.MCPServers))
	}
}
