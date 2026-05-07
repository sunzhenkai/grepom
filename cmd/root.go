package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
)

var (
	configFile string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "grepom",
	Short: "Git Repository Orchestrator & Manager",
	Long: `A CLI tool for managing multiple git repositories across GitLab groups and GitHub organizations.

Use YAML configuration files to define resources (authentication), groups (remote paths),
and standalone repos. Grepom discovers and manages repositories automatically.`,
	Example: `  grepom -c work.yml list              # List repos from a specific config
  grepom clone --group frontend        # Clone all repos in a group
  grepom status                        # Check status of all cloned repos
  grepom pull web-app                  # Pull updates for a single repo`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.EnablePrefixMatching = true
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "path to config file (default: .grepom.yml in current or parent directory)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}

func loadConfig() (string, *config.Config, error) {
	config.SetVerbose(verbose)
	path, err := config.FindConfig(configFile)
	if err != nil {
		return "", nil, err
	}
	// 确保返回绝对路径，因为 FindConfig 在当前目录找到时返回 ".grepom.yml"
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", nil, err
	}
	cfg, err := config.Load(absPath)
	if err != nil {
		return "", nil, err
	}
	// 将相对 base 解析为绝对路径（相对于配置文件所在目录）
	config.ResolveBasePath(cfg, filepath.Dir(absPath))
	return absPath, cfg, nil
}

func resolvedConfigPath() (string, error) {
	return config.FindConfig(configFile)
}

// defaultConfigPath 返回默认的配置文件路径（用于创建新配置）。
// 当用户未通过 -c 指定路径时，返回 ".grepom.yml"。
func defaultConfigPath() string {
	if configFile != "" {
		return configFile
	}
	return ".grepom.yml"
}
