package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	"github.com/wii/grepom/service"
	svctui "github.com/wii/grepom/service/tui"
)

const svcShellHelper = `gsvc() {
  local dir
  if [ $# -eq 0 ]; then
  dir=$(grepom svc dir 2>/dev/null | head -n 1)
  else
    dir=$(grepom svc dir "$@" 2>/dev/null | head -n 1)
  fi || return
  cd "$dir"
}`

var (
	svcForce       bool
	svcListVerbose bool
	svcLogLines    int
	svcLogFollow   bool
	svcLogOpen     bool
	svcCleanLogs   bool
	svcCleanAll    bool
	svcShell       bool
)

var svcCmd = &cobra.Command{
	Use:     "svc",
	Aliases: []string{},
	Short:   "Manage local development service processes",
	Long: `Start, list, monitor, and stop local development services.

Services run in the background with runtime state stored under the XDG state directory.
Service definitions can be declared in .grepom.yml or passed directly on the command line.`,
	Example: `  grepom svc run -- make dev
  grepom svc run api
  grepom svc run api -- make dev
  grepom svc list
  grepom svc logs -f api
  grepom svc kill api
  grepom svc clean
  grepom svc dir api
  grepom svc tui`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if svcShell {
			fmt.Println(svcShellHelper)
			return nil
		}
		return cmd.Help()
	},
}

var serviceCmd = &cobra.Command{
	Use:        "service",
	Short:      "Alias for svc",
	Long:       svcCmd.Long,
	Example:    svcCmd.Example,
	SilenceUsage: true,
	RunE: svcCmd.RunE,
}

func init() {
	svcCmd.Flags().BoolVar(&svcShell, "shell", false, "print gsvc() shell function for cd to service directories")
	serviceCmd.Flags().BoolVar(&svcShell, "shell", false, "print gsvc() shell function for cd to service directories")

	registerSvcSubcommands(svcCmd)
	registerSvcSubcommands(serviceCmd)
	rootCmd.AddCommand(svcCmd)
	rootCmd.AddCommand(serviceCmd)
}

func registerSvcSubcommands(parent *cobra.Command) {
	run := newSvcRunCmd()
	list := newSvcListCmd()
	status := newSvcStatusCmd()
	logs := newSvcLogsCmd()
	kill := newSvcKillCmd()
	clean := newSvcCleanCmd()
	dir := newSvcDirCmd()
	tui := newSvcTuiCmd()

	for _, c := range []*cobra.Command{run, logs, kill, dir, status} {
		c.ValidArgsFunction = completeSvcNames
	}

	parent.AddCommand(run, list, status, logs, kill, clean, dir, tui)
}

func newSvcRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [name] [-- <command> [args...]]",
		Short: "Start a service in the background",
		Long: `Start a service in the background and record its PID and log path.

Without a name, the current directory name is used.
With only a name, the service definition is loaded from .grepom.yml.
With -- <command>, the command is executed in the service directory.`,
		Example: `  grepom svc run -- make dev
  grepom svc run api
  grepom svc run api -- pnpm dev
  grepom svc run --force -- make dev`,
		Args: cobra.ArbitraryArgs,
		RunE: runSvcRun,
	}
	cmd.Flags().BoolVar(&svcForce, "force", false, "replace an already running service with the same name")
	return cmd
}

func newSvcListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List managed services",
		Example: `  grepom svc list
  grepom svc list -v`,
		RunE:    runSvcList,
	}
	cmd.Flags().BoolVarP(&svcListVerbose, "verbose", "v", false, "show command and log path columns")
	return cmd
}

func newSvcStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status [name]",
		Short:   "Show status for one or all services",
		Args:    cobra.MaximumNArgs(1),
		Example: `  grepom svc status
  grepom svc status api`,
		RunE: runSvcStatus,
	}
}

func newSvcLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logs [name]",
		Short:   "View service logs",
		Args:    cobra.ExactArgs(1),
		Example: `  grepom svc logs api
  grepom svc logs -n 200 api
  grepom svc logs -f api
  grepom svc logs --open api`,
		RunE: runSvcLogs,
	}
	cmd.Flags().IntVarP(&svcLogLines, "lines", "n", 0, "number of trailing log lines to print")
	cmd.Flags().BoolVarP(&svcLogFollow, "follow", "f", false, "follow log output")
	cmd.Flags().BoolVar(&svcLogOpen, "open", false, "open log file in editor or default opener")
	return cmd
}

func newSvcKillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kill [name]",
		Short:   "Stop a running service",
		Args:    cobra.ExactArgs(1),
		Example: `  grepom svc kill api
  grepom svc kill -9 api`,
		RunE: runSvcKill,
	}
	cmd.Flags().Bool("9", false, "force kill (SIGKILL)")
	return cmd
}

func newSvcCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clean",
		Short:   "Remove exited or stale service records",
		Example: `  grepom svc clean
  grepom svc clean --logs
  grepom svc clean --all`,
		RunE: runSvcClean,
	}
	cmd.Flags().BoolVar(&svcCleanLogs, "logs", false, "also delete log files for cleaned services")
	cmd.Flags().BoolVar(&svcCleanAll, "all", false, "clean all non-running services")
	return cmd
}

func newSvcDirCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "dir [name]",
		Short:   "Print service working directory",
		Args:    cobra.ExactArgs(1),
		Example: `  grepom svc dir api
  cd "$(grepom svc dir api)"`,
		RunE: runSvcDir,
	}
}

func newSvcTuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "tui",
		Short:   "Open interactive service management UI",
		Example: `  grepom svc tui`,
		RunE:    runSvcTui,
	}
}

func resolveServiceManager() (*service.Manager, error) {
	configPath := ""
	services := map[string]config.ServiceDef{}

	path, cfg, err := tryLoadSvcConfig()
	if err == nil {
		configPath = path
		if cfg.Services != nil {
			services = cfg.Services
		}
	} else if !config.IsConfigNotFound(err) {
		return nil, err
	}

	return service.NewManager(configPath, services)
}

func tryLoadSvcConfig() (string, *config.Config, error) {
	path, err := resolvedConfigPath()
	if err != nil {
		return "", nil, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", nil, err
	}
	cfg, err := config.Load(abs)
	if err != nil {
		return "", nil, err
	}
	config.ResolveBasePath(cfg, filepath.Dir(abs))
	return abs, cfg, nil
}

func runSvcRun(cmd *cobra.Command, args []string) error {
	mgr, err := resolveServiceManager()
	if err != nil {
		return err
	}

	isConfigured := func(name string) bool {
		_, ok := mgr.Services[name]
		return ok
	}
	name, command, useConfig, err := parseSvcRunArgs(args, isConfigured)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if useConfig {
		def, ok := mgr.Services[name]
		if !ok {
			return fmt.Errorf("service %q is not configured in .grepom.yml", name)
		}
		if def.Command.IsEmpty() {
			return fmt.Errorf("service %q has no command configured", name)
		}
		cwd = config.ResolveServiceCwd(mgr.ConfigDir, def.Cwd)
		command = def.Command
	} else if command.IsEmpty() {
		return fmt.Errorf("command is required; use -- <command> or configure the service in .grepom.yml")
	}

	rec, err := mgr.Run(service.RunOptions{
		Name:    name,
		Cwd:     cwd,
		Command: command,
		Force:   svcForce,
	})
	if err != nil {
		return err
	}

	fmt.Printf("started service %q (pid %d)\n", rec.Name, rec.PID)
	fmt.Printf("log: %s\n", rec.LogPath)
	return nil
}

func parseSvcRunArgs(args []string, isConfigured func(string) bool) (string, config.ServiceCommand, bool, error) {
	dash := -1
	for i, a := range args {
		if a == "--" {
			dash = i
			break
		}
	}

	if dash >= 0 {
		before := args[:dash]
		cmdArgs := args[dash+1:]
		if len(cmdArgs) == 0 {
			return "", config.ServiceCommand{}, false, fmt.Errorf("missing command after --")
		}
		name := defaultServiceName()
		if len(before) == 1 {
			name = before[0]
		} else if len(before) > 1 {
			return "", config.ServiceCommand{}, false, fmt.Errorf("unexpected arguments before --")
		}
		return name, config.ServiceCommand{Args: cmdArgs}, false, nil
	}

	switch len(args) {
	case 0:
		return "", config.ServiceCommand{}, false, fmt.Errorf("command is required; use -- <command>")
	case 1:
		return args[0], config.ServiceCommand{}, true, nil
	default:
		if isConfigured != nil && isConfigured(args[0]) {
			return args[0], config.ServiceCommand{Args: args[1:]}, false, nil
		}
		return defaultServiceName(), config.ServiceCommand{Args: args}, false, nil
	}
}

func defaultServiceName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "service"
	}
	return filepath.Base(cwd)
}

func runSvcList(cmd *cobra.Command, args []string) error {
	mgr, err := resolveServiceManager()
	if err != nil {
		return err
	}
	entries, err := mgr.List()
	if err != nil {
		return err
	}
	return printSvcTable(os.Stdout, entries, svcListVerbose)
}

func runSvcStatus(cmd *cobra.Command, args []string) error {
	mgr, err := resolveServiceManager()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return runSvcList(cmd, args)
	}
	entry, err := mgr.Status(args[0])
	if err != nil {
		return err
	}
	return printSvcTable(os.Stdout, []service.Entry{*entry}, true)
}

func printSvcTable(w io.Writer, entries []service.Entry, verbose bool) error {
	if len(entries) == 0 {
		fmt.Fprintln(w, "No services found.")
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	if verbose {
		fmt.Fprintln(tw, "NAME\tSTATUS\tPID\tPATH\tCOMMAND\tLOG")
	} else {
		fmt.Fprintln(tw, "NAME\tSTATUS\tPID\tPATH")
	}
	for _, e := range entries {
		pid := "-"
		if e.Record.PID > 0 {
			pid = fmt.Sprintf("%d", e.Record.PID)
		}
		path := shortenHomePath(e.Record.Cwd)
		if verbose {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
				e.Record.Name, e.Status, pid, path, e.Record.Command, e.Record.LogPath)
		} else {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
				e.Record.Name, e.Status, pid, path)
		}
	}
	return tw.Flush()
}

func shortenHomePath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == home {
		return "~"
	}
	prefix := home + string(filepath.Separator)
	if strings.HasPrefix(path, prefix) {
		return "~" + path[len(home):]
	}
	return path
}

func runSvcLogs(cmd *cobra.Command, args []string) error {
	mgr, err := resolveServiceManager()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if svcLogFollow {
		ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	}
	return mgr.Logs(ctx, args[0], svcLogLines, svcLogFollow, svcLogOpen, os.Stdout)
}

func runSvcKill(cmd *cobra.Command, args []string) error {
	mgr, err := resolveServiceManager()
	if err != nil {
		return err
	}
	force, _ := cmd.Flags().GetBool("9")
	if err := mgr.Kill(args[0], force); err != nil {
		return err
	}
	fmt.Printf("stopped service %q\n", args[0])
	return nil
}

func runSvcClean(cmd *cobra.Command, args []string) error {
	mgr, err := resolveServiceManager()
	if err != nil {
		return err
	}
	removed, err := mgr.Clean(service.CleanOptions{
		RemoveLogs: svcCleanLogs,
		All:        svcCleanAll,
	})
	if err != nil {
		return err
	}
	fmt.Printf("cleaned %d service record(s)\n", removed)
	return nil
}

func runSvcDir(cmd *cobra.Command, args []string) error {
	mgr, err := resolveServiceManager()
	if err != nil {
		return err
	}
	dir, err := mgr.Dir(args[0])
	if err != nil {
		return err
	}
	fmt.Println(dir)
	return nil
}

func runSvcTui(cmd *cobra.Command, args []string) error {
	if err := svctui.EnsureTTY(); err != nil {
		return err
	}
	mgr, err := resolveServiceManager()
	if err != nil {
		return err
	}
	return svctui.Run(context.Background(), mgr)
}
