package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol (MCP) integration",
	Long: `Expose grepom's multi-repo management capabilities to local AI agents over MCP (stdio),
or install the MCP server into an agent's configuration.

Typical flow:
  grepom mcp install cursor   # write grepom into the agent's config
  # restart the agent; the grepom_* tools are now available`,
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the grepom MCP server over stdio",
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := newGrepomMCPServer()
		return srv.Run(cmd.Context(), &mcp.StdioTransport{})
	},
}

var mcpListToolsCmd = &cobra.Command{
	Use:   "list-tools",
	Short: "List the MCP tools exposed by grepom",
	RunE: func(cmd *cobra.Command, args []string) error {
		catalogue := grepomToolCatalogue()
		data, err := json.MarshalIndent(catalogue, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

func newGrepomMCPServer() *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{Name: "grepom", Version: Version}, nil)
	registerGrepomTools(srv)
	return srv
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpServeCmd)
	mcpCmd.AddCommand(mcpListToolsCmd)
}
