package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/provider"
	"github.com/wii/grepom/repo"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start interactive mode with menu-driven UI",
	Long:  "Launch an interactive menu for common operations: init, add resource/group/repo, sync, clone, status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TTY check
		if !isTerminal() {
			return fmt.Errorf("interactive mode requires a TTY")
		}

		mainMenu()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

// isTerminal checks if stdin is connected to a terminal.
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// mainMenu displays the main menu loop.
func mainMenu() {
	for {
		choice := ""
		prompt := &survey.Select{
			Message: "Select an action:",
			Options: []string{
				"Initialize config (init)",
				"Add resource",
				"Add group",
				"Add repo",
				"Sync remote repos",
				"Clone repos",
				"Pull updates",
				"Check status",
				"Exit",
			},
		}
		if err := survey.AskOne(prompt, &choice); err != nil {
			if err == terminal.InterruptErr {
				fmt.Println("\nBye!")
				return
			}
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}

		switch choice {
		case "Initialize config (init)":
			interactiveInit()
		case "Add resource":
			interactiveAddResource()
		case "Add group":
			interactiveAddGroup()
		case "Add repo":
			interactiveAddRepo()
		case "Sync remote repos":
			interactiveSync()
		case "Clone repos":
			interactiveClone()
		case "Pull updates":
			interactivePull()
		case "Check status":
			interactiveStatus()
		case "Exit":
			fmt.Println("Bye!")
			return
		}

		fmt.Println()
	}
}

// --- Interactive init (task 8.1) ---

func interactiveInit() {
	answers := struct {
		ConfigPath string
		Base       string
		AddRes     bool
		Provider   string
		URL        string
		Token      string
		SSHKey     string
	}{}

	// Config file path
	survey.AskOne(&survey.Input{
		Message: "Config file path:",
		Default: ".grepom.yml",
	}, &answers.ConfigPath)

	// Base directory
	survey.AskOne(&survey.Input{
		Message: "Base directory:",
		Default: "~/projects",
	}, &answers.Base)

	// Determine config path
	path := answers.ConfigPath
	if path == "" {
		path = ".grepom.yml"
	}

	// Create config
	if err := config.InitConfig(path, answers.Base); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Printf("Config file created: %s\n", path)

	// Ask if user wants to add a resource
	survey.AskOne(&survey.Confirm{
		Message: "Add first resource?",
		Default: true,
	}, &answers.AddRes)

	if !answers.AddRes {
		return
	}

	// Provider selection
	survey.AskOne(&survey.Select{
		Message: "Provider type:",
		Options: []string{"gitlab", "github", "generic"},
	}, &answers.Provider)

	// URL
	defaultURL := ""
	switch answers.Provider {
	case "gitlab":
		defaultURL = "https://gitlab.com"
	case "github":
		defaultURL = "https://github.com"
	}
	survey.AskOne(&survey.Input{
		Message: "API URL:",
		Default: defaultURL,
	}, &answers.URL)

	// Token
	survey.AskOne(&survey.Input{
		Message: "Token (supports ${ENV_VAR} syntax):",
	}, &answers.Token)

	// SSH key (optional)
	survey.AskOne(&survey.Input{
		Message: "SSH key path (optional, leave empty to skip):",
		Default: "",
	}, &answers.SSHKey)

	// Add resource
	name := answers.Provider
	res := config.Resource{
		Provider: answers.Provider,
		URL:      answers.URL,
		Token:    answers.Token,
		SSHKey:   answers.SSHKey,
	}
	if err := config.AddResource(path, name, res); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Printf("Added resource %s to %s\n", name, path)
}

// --- Interactive add resource (task 8.2) ---

func interactiveAddResource() {
	answers := struct {
		Name     string
		Provider string
		URL      string
		Token    string
		SSHKey   string
		Confirm  bool
	}{}

	survey.AskOne(&survey.Input{
		Message: "Resource name:",
	}, &answers.Name)
	if answers.Name == "" {
		fmt.Println("Cancelled")
		return
	}

	survey.AskOne(&survey.Select{
		Message: "Provider type:",
		Options: []string{"gitlab", "github", "generic"},
	}, &answers.Provider)

	defaultURL := ""
	switch answers.Provider {
	case "gitlab":
		defaultURL = "https://gitlab.com"
	case "github":
		defaultURL = "https://github.com"
	}
	survey.AskOne(&survey.Input{
		Message: "API URL:",
		Default: defaultURL,
	}, &answers.URL)

	survey.AskOne(&survey.Input{
		Message: "Token (supports ${ENV_VAR} syntax):",
	}, &answers.Token)

	survey.AskOne(&survey.Input{
		Message: "SSH key path (optional, leave empty to skip):",
	}, &answers.SSHKey)

	// Confirmation
	fmt.Printf("\nResource: %s\n  provider: %s\n  url: %s\n  token: %s\n  ssh_key: %s\n",
		answers.Name, answers.Provider, answers.URL, maskToken(answers.Token), answers.SSHKey)
	survey.AskOne(&survey.Confirm{
		Message: "Confirm add?",
		Default: true,
	}, &answers.Confirm)

	if !answers.Confirm {
		fmt.Println("Cancelled")
		return
	}

	path, err := resolvedConfigPath()
	if err != nil {
		path = configFile
	}

	res := config.Resource{
		Provider: answers.Provider,
		URL:      answers.URL,
		Token:    answers.Token,
		SSHKey:   answers.SSHKey,
	}
	if err := config.AddResource(path, answers.Name, res); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Printf("Added resource %s to %s\n", answers.Name, path)
}

// --- Interactive add group (task 8.3) ---

func interactiveAddGroup() {
	// Need config to list resources
	path, err := resolvedConfigPath()
	if err != nil {
		path = configFile
	}

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return
	}

	if len(cfg.Resources) == 0 {
		fmt.Println("No resources configured. Please add a resource first.")
		return
	}

	// Build resource list for selection
	var resourceNames []string
	for name := range cfg.Resources {
		resourceNames = append(resourceNames, name)
	}

	answers := struct {
		Name      string
		Resource  string
		Path      string
		LocalPath string
		Recursive bool
		SSHKey    string
		Token     string
		Confirm   bool
	}{}

	survey.AskOne(&survey.Input{
		Message: "Group name:",
	}, &answers.Name)
	if answers.Name == "" {
		fmt.Println("Cancelled")
		return
	}

	survey.AskOne(&survey.Select{
		Message: "Linked resource:",
		Options: resourceNames,
	}, &answers.Resource)

	survey.AskOne(&survey.Input{
		Message: "Remote path (e.g. my-org/frontend):",
	}, &answers.Path)

	survey.AskOne(&survey.Input{
		Message: "Local path:",
		Default: "./" + answers.Name,
	}, &answers.LocalPath)

	// Check if resource is gitlab for recursive option
	res := cfg.Resources[answers.Resource]
	if res.Provider == "gitlab" {
		survey.AskOne(&survey.Confirm{
			Message: "Recursive?",
			Default: false,
		}, &answers.Recursive)
	}

	// Optional SSH key
	survey.AskOne(&survey.Input{
		Message: "SSH key path (optional, overrides resource default, leave empty to skip):",
	}, &answers.SSHKey)

	// Optional token
	survey.AskOne(&survey.Input{
		Message: "Token (optional, overrides resource default, supports ${ENV_VAR}, leave empty to skip):",
	}, &answers.Token)

	// Confirmation
	fmt.Printf("\nGroup: %s\n  resource: %s\n  path: %s\n  local: %s\n  recursive: %v\n",
		answers.Name, answers.Resource, answers.Path, answers.LocalPath, answers.Recursive)
	if answers.SSHKey != "" {
		fmt.Printf("  ssh_key: %s\n", answers.SSHKey)
	}
	if answers.Token != "" {
		fmt.Printf("  token: %s\n", maskToken(answers.Token))
	}
	survey.AskOne(&survey.Confirm{
		Message: "Confirm add?",
		Default: true,
	}, &answers.Confirm)

	if !answers.Confirm {
		fmt.Println("Cancelled")
		return
	}

	group := config.Group{
		Name:      answers.Name,
		Resource:  answers.Resource,
		Path:      answers.Path,
		LocalPath: answers.LocalPath,
		Recursive: answers.Recursive,
		SSHKey:    answers.SSHKey,
		Token:     answers.Token,
	}
	if err := config.AddGroup(path, group); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Printf("Added group %s to %s\n", answers.Name, path)
}

// --- Interactive add repo (task 8.4) ---

func interactiveAddRepo() {
	// First ask: standalone or group repo
	repoType := ""
	survey.AskOne(&survey.Select{
		Message: "Repo type:",
		Options: []string{"Standalone repo", "Add to group"},
	}, &repoType)

	switch repoType {
	case "Standalone repo":
		interactiveAddStandaloneRepo()
	case "Add to group":
		interactiveAddGroupRepo()
	}
}

func interactiveAddStandaloneRepo() {
	path, err := resolvedConfigPath()
	if err != nil {
		path = configFile
	}

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return
	}

	if len(cfg.Resources) == 0 {
		fmt.Println("No resources configured. Please add a resource first.")
		return
	}

	var resourceNames []string
	for name := range cfg.Resources {
		resourceNames = append(resourceNames, name)
	}

	answers := struct {
		Name      string
		Resource  string
		URL       string
		LocalPath string
		SSHKey    string
		Token     string
		Confirm   bool
	}{}

	survey.AskOne(&survey.Input{
		Message: "Repo name:",
	}, &answers.Name)
	if answers.Name == "" {
		fmt.Println("Cancelled")
		return
	}

	survey.AskOne(&survey.Select{
		Message: "Linked resource:",
		Options: resourceNames,
	}, &answers.Resource)

	survey.AskOne(&survey.Input{
		Message: "Clone URL:",
	}, &answers.URL)
	if answers.URL == "" {
		fmt.Println("Cancelled")
		return
	}

	survey.AskOne(&survey.Input{
		Message: "Local path:",
		Default: "./" + answers.Name,
	}, &answers.LocalPath)

	// Optional SSH key
	survey.AskOne(&survey.Input{
		Message: "SSH key path (optional, overrides resource default, leave empty to skip):",
	}, &answers.SSHKey)

	// Optional token
	survey.AskOne(&survey.Input{
		Message: "Token (optional, overrides resource default, supports ${ENV_VAR}, leave empty to skip):",
	}, &answers.Token)

	// Confirmation
	fmt.Printf("\nRepo: %s\n  resource: %s\n  url: %s\n  local: %s\n",
		answers.Name, answers.Resource, answers.URL, answers.LocalPath)
	if answers.SSHKey != "" {
		fmt.Printf("  ssh_key: %s\n", answers.SSHKey)
	}
	if answers.Token != "" {
		fmt.Printf("  token: %s\n", maskToken(answers.Token))
	}
	survey.AskOne(&survey.Confirm{
		Message: "Confirm add?",
		Default: true,
	}, &answers.Confirm)

	if !answers.Confirm {
		fmt.Println("Cancelled")
		return
	}

	repo := config.Repo{
		Name:      answers.Name,
		Resource:  answers.Resource,
		URL:       answers.URL,
		LocalPath: answers.LocalPath,
		SSHKey:    answers.SSHKey,
		Token:     answers.Token,
	}
	if err := config.AddRepo(path, repo); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Printf("Added repo %s to %s\n", answers.Name, path)
}

func interactiveAddGroupRepo() {
	path, err := resolvedConfigPath()
	if err != nil {
		path = configFile
	}

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return
	}

	if len(cfg.Groups) == 0 {
		fmt.Println("No groups configured. Please add a group first.")
		return
	}

	var groupNames []string
	for _, g := range cfg.Groups {
		groupNames = append(groupNames, g.Name)
	}

	answers := struct {
		Group   string
		Name    string
		URL     string
		Path    string
		Confirm bool
	}{}

	survey.AskOne(&survey.Select{
		Message: "Target group:",
		Options: groupNames,
	}, &answers.Group)

	survey.AskOne(&survey.Input{
		Message: "Repo name:",
	}, &answers.Name)
	if answers.Name == "" {
		fmt.Println("Cancelled")
		return
	}

	survey.AskOne(&survey.Input{
		Message: "Clone URL:",
	}, &answers.URL)
	if answers.URL == "" {
		fmt.Println("Cancelled")
		return
	}

	survey.AskOne(&survey.Input{
		Message: "Remote path (in-group path, e.g. my-org/frontend/repo-name):",
	}, &answers.Path)

	// Confirmation
	fmt.Printf("\nRepo in group: %s -> %s\n  url: %s\n  path: %s\n",
		answers.Name, answers.Group, answers.URL, answers.Path)
	survey.AskOne(&survey.Confirm{
		Message: "Confirm add?",
		Default: true,
	}, &answers.Confirm)

	if !answers.Confirm {
		fmt.Println("Cancelled")
		return
	}

	repo := config.GroupRepo{
		Name: answers.Name,
		URL:  answers.URL,
		Path: answers.Path,
	}
	if repo.Path == "" {
		repo.Path = answers.Name
	}
	if err := config.AddGroupRepo(path, answers.Group, repo); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Printf("Added repo %s to group %s\n", answers.Name, answers.Group)
}

// --- Interactive sync (task 8.5) ---

func interactiveSync() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	configPath, err := resolvedConfigPath()
	if err != nil {
		configPath = configFile
	}

	// Determine scope
	scope := ""
	scopeOptions := []string{"All", "By group", "By resource"}
	survey.AskOne(&survey.Select{
		Message: "Sync scope:",
		Options: scopeOptions,
	}, &scope)

	var groupsToProcess []config.Group

	switch scope {
	case "All":
		groupsToProcess = cfg.Groups
	case "By group":
		if len(cfg.Groups) == 0 {
			fmt.Println("No groups configured")
			return
		}
		var groupNames []string
		for _, g := range cfg.Groups {
			groupNames = append(groupNames, g.Name)
		}
		selectedGroup := ""
		survey.AskOne(&survey.Select{
			Message: "Select group:",
			Options: groupNames,
		}, &selectedGroup)
		for _, g := range cfg.Groups {
			if g.Name == selectedGroup {
				groupsToProcess = append(groupsToProcess, g)
			}
		}
	case "By resource":
		if len(cfg.Resources) == 0 {
			fmt.Println("No resources configured")
			return
		}
		var resourceNames []string
		for name := range cfg.Resources {
			resourceNames = append(resourceNames, name)
		}
		selectedResource := ""
		survey.AskOne(&survey.Select{
			Message: "Select resource:",
			Options: resourceNames,
		}, &selectedResource)
		for _, g := range cfg.Groups {
			if g.Resource == selectedResource {
				groupsToProcess = append(groupsToProcess, g)
			}
		}
	}

	if len(groupsToProcess) == 0 {
		fmt.Println("No groups to sync")
		return
	}

	var totalRepos, totalNewRepos int

	for _, g := range groupsToProcess {
		res, ok := cfg.Resources[g.Resource]
		if !ok {
			fmt.Fprintf(os.Stderr, "error: group %q: resource %q not found\n", g.Name, g.Resource)
			continue
		}

		p, err := provider.Get(res.Provider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: group %q: %v\n", g.Name, err)
			continue
		}

		config.Verbose("syncing group %q (resource: %s, path: %s)", g.Name, g.Resource, g.Path)

		resolvedToken, err := res.ResolvedToken()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: group %q (resource %q): %v\n", g.Name, g.Resource, err)
			continue
		}

		params := provider.ListReposParams{
			ServerURL: res.URL,
			Token:     resolvedToken,
		}

		if res.Provider == "github" {
			params.Orgs = []string{g.Path}
		} else {
			params.Groups = []provider.GroupQuery{{Path: g.Path, Recursive: g.Recursive}}
		}

		repos, err := p.ListRepos(context.Background(), params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: group %q: %v\n", g.Name, err)
			continue
		}

		totalRepos += len(repos)
		fmt.Printf("group %q: found %d repos\n", g.Name, len(repos))

		var newGroupRepos []config.GroupRepo
		for _, r := range repos {
			newGroupRepos = append(newGroupRepos, config.GroupRepo{
				Name: r.Name,
				URL:  r.CloneURL,
				Path: r.Path,
			})
		}

		if len(newGroupRepos) > 0 {
			added, err := config.SyncGroupRepos(configPath, g.Name, newGroupRepos)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: group %q: %v\n", g.Name, err)
			} else if added > 0 {
				totalNewRepos += added
				fmt.Printf("  added %d new repos to group %q\n", added, g.Name)
			}
		}
	}

	fmt.Printf("\nSync complete: %d repos discovered, %d new\n", totalRepos, totalNewRepos)
}

// --- Interactive clone (task 8.6) ---

func interactiveClone() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	// Determine scope
	scope := ""
	scopeOptions := []string{"All", "By group", "By resource"}
	survey.AskOne(&survey.Select{
		Message: "Clone scope:",
		Options: scopeOptions,
	}, &scope)

	filter := repo.Filter{}

	switch scope {
	case "By group":
		if len(cfg.Groups) == 0 {
			fmt.Println("No groups configured")
			return
		}
		var groupNames []string
		for _, g := range cfg.Groups {
			groupNames = append(groupNames, g.Name)
		}
		selectedGroup := ""
		survey.AskOne(&survey.Select{
			Message: "Select group:",
			Options: groupNames,
		}, &selectedGroup)
		filter.Group = selectedGroup
	case "By resource":
		if len(cfg.Resources) == 0 {
			fmt.Println("No resources configured")
			return
		}
		var resourceNames []string
		for name := range cfg.Resources {
			resourceNames = append(resourceNames, name)
		}
		selectedResource := ""
		survey.AskOne(&survey.Select{
			Message: "Select resource:",
			Options: resourceNames,
		}, &selectedResource)
		filter.Resource = selectedResource
	}

	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	if len(repos) == 0 {
		fmt.Println("No repos to clone")
		return
	}

	// Ask for concurrency
	concurrency := 4
	survey.AskOne(&survey.Select{
		Message: "Concurrency:",
		Options: []string{"1 (sequential)", "2", "4 (default)", "8"},
	}, &scope)
	switch scope {
	case "1 (sequential)":
		concurrency = 1
	case "2":
		concurrency = 2
	case "4 (default)":
		concurrency = 4
	case "8":
		concurrency = 8
	}

	// Filter out already-cloned repos
	var toClone []gitpkg.CloneTask
	for _, r := range repos {
		fullPath := repo.FullPath(cfg.Base, r)
		if gitpkg.IsCloned(fullPath) {
			fmt.Printf("skipped %s (already cloned)\n", r.Path)
			continue
		}
		toClone = append(toClone, gitpkg.CloneTask{
			Repo:     r,
			FullPath: fullPath,
		})
	}

	if len(toClone) == 0 {
		fmt.Println("All repos already cloned")
		return
	}

	if concurrency > 1 && len(toClone) > 1 {
		// Parallel clone
		progress := NewProgressRenderer("cloning", len(toClone))
		defer progress.Done()

		results := gitpkg.CloneAll(concurrency, toClone, func(event gitpkg.ProgressEvent) {
			progress.Handle(event)
		})
		PrintCloneSummary(results, nil)
	} else {
		// Sequential clone
		for _, task := range toClone {
			fmt.Printf("Cloning %s...\n", task.Repo.Path)
			opts := gitpkg.CloneOptions{
				Token:          task.Repo.Token,
				Provider:       task.Repo.Provider,
				SSHKey:         task.Repo.SSHKey,
				HasGroupToken:  task.Repo.HasGroupToken,
				HasGroupSSHKey: task.Repo.HasGroupSSHKey,
			}
			if err := gitpkg.Clone(task.FullPath, task.Repo.SSHURL, task.Repo.CloneURL, opts); err != nil {
				fmt.Fprintf(os.Stderr, "clone %s failed: %v\n", task.Repo.Path, err)
				continue
			}
			fmt.Printf("  %s done\n", task.Repo.Name)
		}
	}
}

// --- Interactive pull ---

func interactivePull() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	// Determine scope
	scope := ""
	scopeOptions := []string{"All", "By group", "By resource"}
	survey.AskOne(&survey.Select{
		Message: "Pull scope:",
		Options: scopeOptions,
	}, &scope)

	filter := repo.Filter{}

	switch scope {
	case "By group":
		if len(cfg.Groups) == 0 {
			fmt.Println("No groups configured")
			return
		}
		var groupNames []string
		for _, g := range cfg.Groups {
			groupNames = append(groupNames, g.Name)
		}
		selectedGroup := ""
		survey.AskOne(&survey.Select{
			Message: "Select group:",
			Options: groupNames,
		}, &selectedGroup)
		filter.Group = selectedGroup
	case "By resource":
		if len(cfg.Resources) == 0 {
			fmt.Println("No resources configured")
			return
		}
		var resourceNames []string
		for name := range cfg.Resources {
			resourceNames = append(resourceNames, name)
		}
		selectedResource := ""
		survey.AskOne(&survey.Select{
			Message: "Select resource:",
			Options: resourceNames,
		}, &selectedResource)
		filter.Resource = selectedResource
	}

	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	if len(repos) == 0 {
		fmt.Println("No repos found")
		return
	}

	// Ask for concurrency
	concurrency := 4
	survey.AskOne(&survey.Select{
		Message: "Concurrency:",
		Options: []string{"1 (sequential)", "2", "4 (default)", "8"},
	}, &scope)
	switch scope {
	case "1 (sequential)":
		concurrency = 1
	case "2":
		concurrency = 2
	case "4 (default)":
		concurrency = 4
	case "8":
		concurrency = 8
	}

	// Safety check: determine which repos are eligible
	var toPull []gitpkg.PullTask
	skipped := 0

	for _, r := range repos {
		fullPath := repo.FullPath(cfg.Base, r)
		if !gitpkg.IsCloned(fullPath) {
			fmt.Printf("skipped %s (not cloned)\n", r.Path)
			skipped++
			continue
		}
		ok, reason := gitpkg.CheckPullSafety(fullPath)
		if !ok {
			fmt.Printf("skipped %s (%s)\n", r.Path, reason)
			skipped++
			continue
		}
		toPull = append(toPull, gitpkg.PullTask{
			Repo:     r,
			FullPath: fullPath,
		})
	}

	if len(toPull) == 0 {
		if skipped > 0 {
			fmt.Printf("No repos to pull: %d skipped\n", skipped)
		} else {
			fmt.Println("No repos to pull")
		}
		return
	}

	if concurrency > 1 && len(toPull) > 1 {
		progress := NewProgressRenderer("pulling", len(toPull))
		defer progress.Done()

		results := gitpkg.PullAll(concurrency, toPull, func(event gitpkg.ProgressEvent) {
			progress.Handle(event)
		})
		PrintPullSummary(results, skipped, nil)
	} else {
		for _, task := range toPull {
			fmt.Printf("Pulling %s...\n", task.Repo.Path)
			if err := gitpkg.Pull(task.FullPath); err != nil {
				fmt.Fprintf(os.Stderr, "pull %s failed: %v\n", task.Repo.Path, err)
				continue
			}
		}
		results := make([]gitpkg.PullResult, 0, len(toPull))
		for _, task := range toPull {
			results = append(results, gitpkg.PullResult{
				Repo:     task.Repo,
				FullPath: task.FullPath,
			})
		}
		PrintPullSummary(results, skipped, nil)
	}
}

// --- Interactive status (task 8.7) ---

func interactiveStatus() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	// Determine scope
	scope := ""
	scopeOptions := []string{"All", "By group", "By resource"}
	survey.AskOne(&survey.Select{
		Message: "Status scope:",
		Options: scopeOptions,
	}, &scope)

	filter := repo.Filter{}

	switch scope {
	case "By group":
		if len(cfg.Groups) == 0 {
			fmt.Println("No groups configured")
			return
		}
		var groupNames []string
		for _, g := range cfg.Groups {
			groupNames = append(groupNames, g.Name)
		}
		selectedGroup := ""
		survey.AskOne(&survey.Select{
			Message: "Select group:",
			Options: groupNames,
		}, &selectedGroup)
		filter.Group = selectedGroup
	case "By resource":
		if len(cfg.Resources) == 0 {
			fmt.Println("No resources configured")
			return
		}
		var resourceNames []string
		for name := range cfg.Resources {
			resourceNames = append(resourceNames, name)
		}
		selectedResource := ""
		survey.AskOne(&survey.Select{
			Message: "Select resource:",
			Options: resourceNames,
		}, &selectedResource)
		filter.Resource = selectedResource
	}

	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	if len(repos) == 0 {
		fmt.Println("No repos found")
		return
	}

	for _, r := range repos {
		fullPath := repo.FullPath(cfg.Base, r)

		st := gitpkg.GetStatus(fullPath)

		if !st.Cloned {
			fmt.Printf("%s: not cloned\n", r.Path)
			continue
		}

		if st.NotARepo {
			fmt.Printf("%s: not a git repo\n", r.Path)
			continue
		}

		parts := []string{st.Branch}

		if st.Clean {
			parts = append(parts, "clean")
		} else {
			parts = append(parts, fmt.Sprintf("dirty (%d files)", st.Dirty))
		}

		if st.Ahead > 0 {
			parts = append(parts, fmt.Sprintf("ahead %d", st.Ahead))
		}
		if st.Behind > 0 {
			parts = append(parts, fmt.Sprintf("behind %d", st.Behind))
		}

		fmt.Printf("%s: %s\n", r.Path, strings.Join(parts, ", "))
	}
}

// --- helpers ---

// maskToken masks a token for display, showing only first/last few chars.
func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "****"
	}
	if strings.HasPrefix(token, "${") && strings.HasSuffix(token, "}") {
		return token // show env var syntax as-is
	}
	return token[:4] + "****" + token[len(token)-4:]
}
