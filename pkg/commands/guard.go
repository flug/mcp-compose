package commands

import (
	"fmt"

	"github.com/flug/mcp-compose/pkg/platform"
)

// EnsureInitialized checks if mcp-compose has been initialized
// Returns an error with helpful message if not initialized
func EnsureInitialized() error {
	platformInfo, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("error detecting platform: %w", err)
	}

	if !platformInfo.MCPComposeConfigExists() {
		return fmt.Errorf(`mcp-compose is not initialized.

Please run the following command first:
  mcp-compose init

This will:
  • Detect your platform and user information
  • Locate your Claude Desktop configuration
  • Set up mcp-compose for use`)
	}

	return nil
}
