package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	mcpInstallScope string
	mcpInstallPrint bool
	mcpInstallAll   bool
)

var mcpInstallCmd = &cobra.Command{
	Use:   "install [agent]",
	Short: "Install the grepom MCP server into an agent's config",
	Long: `Write the grepom MCP server (grepom mcp serve) into a target agent's
configuration file. Existing config and other MCP servers are preserved; a
.bak backup is created before the file is modified.

Supported agents: claude-code, claude-desktop, cursor, codex, zcode, kimi, pi.

Examples:
  grepom mcp install cursor
  grepom mcp install cursor --scope project
  grepom mcp install codex --print
  grepom mcp install --all`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if mcpInstallAll {
			return grepomInstallAll(mcpInstallScope, mcpInstallPrint, cmd.OutOrStdout())
		}
		if len(args) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), grepomFormatAgentList())
			return nil
		}
		target, ok := grepomFindAgent(args[0])
		if !ok {
			return fmt.Errorf("unknown agent %q; supported: %s", args[0], strings.Join(grepomSupportedAgentIDs(), ", "))
		}
		return grepomInstallInto(target, mcpInstallScope, mcpInstallPrint, cmd.OutOrStdout())
	},
}

type grepomServerEntryJSON struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

func grepomInstallInto(t mcpAgentTarget, scope string, printOnly bool, out interface{ Write([]byte) (int, error) }) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}
	if scope == "" {
		scope = "user"
	}
	if scope != "user" && scope != "project" {
		return fmt.Errorf("invalid scope %q: must be \"user\" or \"project\"", scope)
	}
	cfgPath := t.configPath(home, scope)
	spec := defaultGrepomServerSpec(!printOnly)

	switch t.format {
	case mcpFormatJSON:
		return grepomInstallJSON(t, cfgPath, spec, printOnly, out)
	case mcpFormatTOML:
		return grepomInstallTOML(t, cfgPath, spec, printOnly, out)
	default:
		return fmt.Errorf("unsupported format for agent %q", t.id)
	}
}

func grepomInstallAll(scope string, printOnly bool, out interface{ Write([]byte) (int, error) }) error {
	var errs []string
	for _, t := range grepomSupportedAgents() {
		if err := grepomInstallInto(t, scope, printOnly, out); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", t.id, err))
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func grepomInstallJSON(t mcpAgentTarget, cfgPath string, spec mcpServerSpec, printOnly bool, out interface{ Write([]byte) (int, error) }) error {
	root := map[string]any{}
	if existing, err := os.ReadFile(cfgPath); err == nil && len(existing) > 0 {
		if err := json.Unmarshal(existing, &root); err != nil {
			return fmt.Errorf("parse %s: %w", cfgPath, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", cfgPath, err)
	}

	servers, _ := root[t.jsonServersKey].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers["grepom"] = grepomServerEntryJSON{Command: spec.Command, Args: spec.Args}
	root[t.jsonServersKey] = servers

	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if printOnly {
		entry, _ := json.MarshalIndent(map[string]any{
			t.jsonServersKey: map[string]any{"grepom": servers["grepom"]},
		}, "", "  ")
		fmt.Fprintf(out, "# %s — add to %s\n%s\n", t.name, cfgPath, entry)
		return nil
	}

	if err := grepomWriteWithBackup(cfgPath, data); err != nil {
		return err
	}
	fmt.Fprintf(out, "Installed grepom MCP server into %s\n  %s\n", t.name, cfgPath)
	if t.note != "" {
		fmt.Fprintf(out, "  %s\n", t.note)
	}
	return nil
}

func grepomInstallTOML(t mcpAgentTarget, cfgPath string, spec mcpServerSpec, printOnly bool, out interface{ Write([]byte) (int, error) }) error {
	block := grepomRenderCodexBlock(spec)
	if printOnly {
		fmt.Fprintf(out, "# %s — add to %s\n%s", t.name, cfgPath, block)
		return nil
	}

	existing, _ := os.ReadFile(cfgPath)
	merged := grepomUpsertTomlServer(string(existing), block)
	if err := grepomWriteWithBackup(cfgPath, []byte(merged)); err != nil {
		return err
	}
	fmt.Fprintf(out, "Installed grepom MCP server into %s\n  %s\n", t.name, cfgPath)
	if t.note != "" {
		fmt.Fprintf(out, "  %s\n", t.note)
	}
	return nil
}

func grepomRenderCodexBlock(spec mcpServerSpec) string {
	var b strings.Builder
	b.WriteString("[mcp_servers.grepom]\n")
	fmt.Fprintf(&b, "command = %q\n", spec.Command)
	fmt.Fprintf(&b, "args = [\"%s\"]\n", strings.Join(spec.Args, "\", \""))
	return b.String()
}

func grepomUpsertTomlServer(src, newBlock string) string {
	tableHeader := "[mcp_servers.grepom]"

	lines := strings.Split(src, "\n")
	var out []string
	inBlock := false
	replaced := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isHeader := strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") && !strings.HasPrefix(trimmed, "[[")
		if isHeader {
			if trimmed == tableHeader || strings.HasPrefix(trimmed, "[mcp_servers.grepom.") {
				inBlock = true
				if !replaced {
					out = append(out, strings.TrimRight(newBlock, "\n"))
					replaced = true
				}
				continue
			}
			inBlock = false
		}
		if inBlock {
			continue
		}
		out = append(out, line)
	}

	result := strings.Join(out, "\n")
	if !replaced {
		if result != "" && !strings.HasSuffix(result, "\n\n") {
			if strings.HasSuffix(result, "\n") {
				result += "\n"
			} else {
				result += "\n\n"
			}
		}
		result += newBlock
	}
	result = strings.TrimRight(result, "\n") + "\n"
	return result
}

func grepomWriteWithBackup(path string, data []byte) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
	}
	if existing, err := os.ReadFile(path); err == nil && len(existing) > 0 {
		if err := os.WriteFile(path+".bak", existing, 0o600); err != nil {
			return fmt.Errorf("write backup: %w", err)
		}
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func init() {
	mcpCmd.AddCommand(mcpInstallCmd)
	mcpInstallCmd.Flags().StringVar(&mcpInstallScope, "scope", "user", "config scope: user or project")
	mcpInstallCmd.Flags().BoolVar(&mcpInstallPrint, "print", false, "print the config snippet instead of writing")
	mcpInstallCmd.Flags().BoolVar(&mcpInstallAll, "all", false, "install into every supported agent")
}
