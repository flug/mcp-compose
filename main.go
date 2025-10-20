package main

import (
	"fmt"
	"os"

	"github.com/flug/mcp-compose/pkg/commands"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s init                                              - Initialize mcp-compose configuration\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s dump                                              - Extract MCP servers from Claude config to YAML\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s apply <config.yaml>                               - Update Claude config with MCP servers from YAML\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s convert <config.json> <project-path>              - Convert JSON MCP config to YAML and optionally apply\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s delete <config.yaml> <server-name> <project-path> - Delete MCP server from YAML and optionally from Claude config\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nMarketplace commands:\n")
	fmt.Fprintf(os.Stderr, "  %s marketplace list                                  - List all available MCP servers from marketplace\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s marketplace search <query>                        - Search for MCP servers in marketplace\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s marketplace install <name> <project-path>         - Install an MCP server from marketplace\n", os.Args[0])
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		if err := commands.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "dump":
		if err := commands.Dump(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "apply":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: apply command requires a YAML file argument\n")
			fmt.Fprintf(os.Stderr, "Usage: %s apply <config.yaml>\n", os.Args[0])
			os.Exit(1)
		}
		if err := commands.Apply(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "convert":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Error: convert command requires a JSON file and project path\n")
			fmt.Fprintf(os.Stderr, "Usage: %s convert <config.json> <project-path>\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "\nExample:\n")
			fmt.Fprintf(os.Stderr, "  %s convert server.json /home/user/workspace/my-project\n", os.Args[0])
			os.Exit(1)
		}
		if err := commands.Convert(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "delete":
		if len(os.Args) < 5 {
			fmt.Fprintf(os.Stderr, "Error: delete command requires YAML file, server name, and project path\n")
			fmt.Fprintf(os.Stderr, "Usage: %s delete <config.yaml> <server-name> <project-path>\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "\nExample:\n")
			fmt.Fprintf(os.Stderr, "  %s delete mcp-config.yaml filesystem /home/user/workspace/my-project\n", os.Args[0])
			os.Exit(1)
		}
		if err := commands.Delete(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "marketplace":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: marketplace command requires a subcommand\n")
			fmt.Fprintf(os.Stderr, "Usage:\n")
			fmt.Fprintf(os.Stderr, "  %s marketplace list\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "  %s marketplace search <query>\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "  %s marketplace install <name> <project-path>\n", os.Args[0])
			os.Exit(1)
		}

		subcommand := os.Args[2]
		switch subcommand {
		case "list":
			if err := commands.MarketplaceList(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		case "search":
			if len(os.Args) < 4 {
				fmt.Fprintf(os.Stderr, "Error: search requires a query argument\n")
				fmt.Fprintf(os.Stderr, "Usage: %s marketplace search <query>\n", os.Args[0])
				os.Exit(1)
			}
			if err := commands.MarketplaceSearch(os.Args[3]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		case "install":
			if len(os.Args) < 5 {
				fmt.Fprintf(os.Stderr, "Error: install requires server name and project path\n")
				fmt.Fprintf(os.Stderr, "Usage: %s marketplace install <name> <project-path>\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "\nExample:\n")
				fmt.Fprintf(os.Stderr, "  %s marketplace install filesystem /home/user/workspace/my-project\n", os.Args[0])
				os.Exit(1)
			}
			if err := commands.MarketplaceInstall(os.Args[3], os.Args[4]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown marketplace subcommand '%s'\n", subcommand)
			fmt.Fprintf(os.Stderr, "Available subcommands: list, search, install\n")
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", command)
		fmt.Fprintf(os.Stderr, "Available commands: init, dump, apply, convert, delete, marketplace\n")
		os.Exit(1)
	}
}
