package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/selfupdate"
)

var (
	updateVersion    string
	updateInstallDir string
	updateForce      bool
	updateRepo       string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Download and install the latest grepom release",
	Long: `Query GitHub Releases for the latest stable version and install the
matching binary for the current platform.

By default, grepom is installed to the directory containing the current
executable. Use --install-dir to override the destination.`,
	Example: `  grepom update
  grepom update --version v0.1.7
  grepom update --install-dir ~/.local/bin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := selfupdate.Update(selfupdate.Options{
			Repo:       updateRepo,
			Version:    updateVersion,
			InstallDir: updateInstallDir,
			Current:    Version,
			Force:      updateForce,
			Out:        cmd.OutOrStdout(),
		})
		if err != nil {
			return err
		}
		if result.AlreadyUpToDate {
			return nil
		}
		if result.InstalledTo != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Restart your shell or run %q to use the new version.\n", result.InstalledTo)
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateVersion, "version", "latest", "release tag to install (default: latest stable)")
	updateCmd.Flags().StringVar(&updateInstallDir, "install-dir", "", "installation directory (default: directory of current executable)")
	updateCmd.Flags().BoolVar(&updateForce, "force", false, "install even if the current version matches")
	updateCmd.Flags().StringVar(&updateRepo, "repo", selfupdate.DefaultRepo, "GitHub repository in owner/name form")
	rootCmd.AddCommand(updateCmd)
}
