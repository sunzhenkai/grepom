package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type mcpAgentFormat int

const (
	mcpFormatJSON mcpAgentFormat = iota
	mcpFormatTOML
)

type mcpServerSpec struct {
	Command string
	Args    []string
}

func defaultGrepomServerSpec(resolve bool) mcpServerSpec {
	cmd := "grepom"
	if resolve {
		if abs, err := grepomAbsExecutable(); err == nil && abs != "" {
			cmd = abs
		}
	}
	return mcpServerSpec{Command: cmd, Args: []string{"mcp", "serve"}}
}

func grepomAbsExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	if abs, err := filepath.Abs(exe); err == nil {
		exe = abs
	}
	return exe, nil
}

type mcpAgentTarget struct {
	id             string
	name           string
	format         mcpAgentFormat
	configPath     func(home, scope string) string
	jsonServersKey string
	tomlTableName  string
	note           string
}

func grepomSupportedAgents() []mcpAgentTarget {
	return []mcpAgentTarget{
		{
			id:             "claude-code",
			name:           "Claude Code",
			format:         mcpFormatJSON,
			configPath:     func(home, _ string) string { return filepath.Join(home, ".claude.json") },
			jsonServersKey: "mcpServers",
			note:           "Restart Claude Code for the server to load.",
		},
		{
			id:             "claude-desktop",
			name:           "Claude Desktop",
			format:         mcpFormatJSON,
			configPath:     grepomClaudeDesktopPath,
			jsonServersKey: "mcpServers",
			note:           "Quit and reopen Claude Desktop to load the server.",
		},
		{
			id:     "cursor",
			name:   "Cursor",
			format: mcpFormatJSON,
			configPath: func(home, scope string) string {
				if scope == "project" {
					return ".cursor/mcp.json"
				}
				return filepath.Join(home, ".cursor", "mcp.json")
			},
			jsonServersKey: "mcpServers",
			note:           "Restart Cursor (or reload the window) for the server to load.",
		},
		{
			id:            "codex",
			name:          "Codex (OpenAI)",
			format:        mcpFormatTOML,
			configPath:    func(home, _ string) string { return filepath.Join(home, ".codex", "config.toml") },
			tomlTableName: "mcp_servers",
			note:          "Restart Codex for the server to load.",
		},
		{
			id:             "zcode",
			name:           "ZCode",
			format:         mcpFormatJSON,
			configPath:     func(home, _ string) string { return filepath.Join(home, ".zcode", "config.json") },
			jsonServersKey: "mcpServers",
			note:           "Restart ZCode for the server to load.",
		},
		{
			id:             "kimi",
			name:           "Kimi CLI",
			format:         mcpFormatJSON,
			configPath:     func(home, _ string) string { return filepath.Join(home, ".kimi", "mcp.json") },
			jsonServersKey: "mcpServers",
			note:           "Restart Kimi CLI for the server to load.",
		},
		{
			id:             "pi",
			name:           "PI",
			format:         mcpFormatJSON,
			configPath:     func(home, _ string) string { return filepath.Join(home, ".pi", "config.json") },
			jsonServersKey: "mcpServers",
			note:           "Restart PI for the server to load.",
		},
	}
}

func grepomClaudeDesktopPath(home, _ string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json")
	default:
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
	}
}

func grepomFindAgent(id string) (mcpAgentTarget, bool) {
	lower := strings.ToLower(id)
	for _, a := range grepomSupportedAgents() {
		if strings.ToLower(a.id) == lower {
			return a, true
		}
	}
	return mcpAgentTarget{}, false
}

func grepomSupportedAgentIDs() []string {
	agents := grepomSupportedAgents()
	ids := make([]string, len(agents))
	for i, a := range agents {
		ids[i] = a.id
	}
	return ids
}

func grepomFormatAgentList() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Supported agents:\n")
	for _, a := range grepomSupportedAgents() {
		fmt.Fprintf(&b, "  %-16s %s\n", a.id, a.name)
	}
	fmt.Fprintf(&b, "\nRun: grepom mcp install <agent> [--scope user|project] [--print]")
	return b.String()
}
