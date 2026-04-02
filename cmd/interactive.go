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
			return fmt.Errorf("interactive 模式需要交互式终端")
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
			Message: "请选择操作:",
			Options: []string{
				"初始化配置 (init)",
				"添加资源 (add resource)",
				"添加组 (add group)",
				"添加仓库 (add repo)",
				"同步远程仓库 (sync)",
				"克隆仓库 (clone)",
				"拉取更新 (pull)",
				"查看状态 (status)",
				"退出",
			},
		}
		if err := survey.AskOne(prompt, &choice); err != nil {
			if err == terminal.InterruptErr {
				fmt.Println("\n再见!")
				return
			}
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		switch choice {
		case "初始化配置 (init)":
			interactiveInit()
		case "添加资源 (add resource)":
			interactiveAddResource()
		case "添加组 (add group)":
			interactiveAddGroup()
		case "添加仓库 (add repo)":
			interactiveAddRepo()
		case "同步远程仓库 (sync)":
			interactiveSync()
		case "克隆仓库 (clone)":
			interactiveClone()
		case "拉取更新 (pull)":
			interactivePull()
		case "查看状态 (status)":
			interactiveStatus()
		case "退出":
			fmt.Println("再见!")
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
		Message: "配置文件路径:",
		Default: ".grepom.yml",
	}, &answers.ConfigPath)

	// Base directory
	survey.AskOne(&survey.Input{
		Message: "base 目录:",
		Default: "~/projects",
	}, &answers.Base)

	// Determine config path
	path := answers.ConfigPath
	if path == "" {
		path = ".grepom.yml"
	}

	// Create config
	if err := config.InitConfig(path, answers.Base); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}
	fmt.Printf("创建配置文件: %s\n", path)

	// Ask if user wants to add a resource
	survey.AskOne(&survey.Confirm{
		Message: "是否添加第一个资源?",
		Default: true,
	}, &answers.AddRes)

	if !answers.AddRes {
		return
	}

	// Provider selection
	survey.AskOne(&survey.Select{
		Message: "Provider 类型:",
		Options: []string{"gitlab", "github"},
	}, &answers.Provider)

	// URL
	defaultURL := "https://gitlab.com"
	if answers.Provider == "github" {
		defaultURL = "https://github.com"
	}
	survey.AskOne(&survey.Input{
		Message: "API URL:",
		Default: defaultURL,
	}, &answers.URL)

	// Token
	survey.AskOne(&survey.Input{
		Message: "Token (支持 ${ENV_VAR} 语法):",
	}, &answers.Token)

	// SSH key (optional)
	survey.AskOne(&survey.Input{
		Message: "SSH key 路径 (可选，留空跳过):",
		Default: "",
	}, &answers.SSHKey)

	// Add resource
	name := "gitlab"
	if answers.Provider == "github" {
		name = "github"
	}
	res := config.Resource{
		Provider: answers.Provider,
		URL:      answers.URL,
		Token:    answers.Token,
		SSHKey:   answers.SSHKey,
	}
	if err := config.AddResource(path, name, res); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}
	fmt.Printf("添加资源 %s 到 %s\n", name, path)
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
		Message: "资源名称:",
	}, &answers.Name)
	if answers.Name == "" {
		fmt.Println("已取消")
		return
	}

	survey.AskOne(&survey.Select{
		Message: "Provider 类型:",
		Options: []string{"gitlab", "github"},
	}, &answers.Provider)

	defaultURL := "https://gitlab.com"
	if answers.Provider == "github" {
		defaultURL = "https://github.com"
	}
	survey.AskOne(&survey.Input{
		Message: "API URL:",
		Default: defaultURL,
	}, &answers.URL)

	survey.AskOne(&survey.Input{
		Message: "Token (支持 ${ENV_VAR} 语法):",
	}, &answers.Token)

	survey.AskOne(&survey.Input{
		Message: "SSH key 路径 (可选，留空跳过):",
	}, &answers.SSHKey)

	// Confirmation
	fmt.Printf("\n资源: %s\n  provider: %s\n  url: %s\n  token: %s\n  ssh_key: %s\n",
		answers.Name, answers.Provider, answers.URL, maskToken(answers.Token), answers.SSHKey)
	survey.AskOne(&survey.Confirm{
		Message: "确认添加?",
		Default: true,
	}, &answers.Confirm)

	if !answers.Confirm {
		fmt.Println("已取消")
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
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}
	fmt.Printf("添加资源 %s 到 %s\n", answers.Name, path)
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
		fmt.Fprintf(os.Stderr, "无法加载配置: %v\n", err)
		return
	}

	if len(cfg.Resources) == 0 {
		fmt.Println("请先添加资源")
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
		Message: "组名称:",
	}, &answers.Name)
	if answers.Name == "" {
		fmt.Println("已取消")
		return
	}

	survey.AskOne(&survey.Select{
		Message: "关联资源:",
		Options: resourceNames,
	}, &answers.Resource)

	survey.AskOne(&survey.Input{
		Message: "远程路径 (如 my-org/frontend):",
	}, &answers.Path)

	survey.AskOne(&survey.Input{
		Message: "本地路径:",
		Default: "./" + answers.Name,
	}, &answers.LocalPath)

	// Check if resource is gitlab for recursive option
	res := cfg.Resources[answers.Resource]
	if res.Provider == "gitlab" {
		survey.AskOne(&survey.Confirm{
			Message: "是否递归 (recursive)?",
			Default: false,
		}, &answers.Recursive)
	}

	// Optional SSH key
	survey.AskOne(&survey.Input{
		Message: "SSH key 路径 (可选，覆盖资源默认，留空跳过):",
	}, &answers.SSHKey)

	// Optional token
	survey.AskOne(&survey.Input{
		Message: "Token (可选，覆盖资源默认，支持 ${ENV_VAR}，留空跳过):",
	}, &answers.Token)

	// Confirmation
	fmt.Printf("\n组: %s\n  resource: %s\n  path: %s\n  local: %s\n  recursive: %v\n",
		answers.Name, answers.Resource, answers.Path, answers.LocalPath, answers.Recursive)
	if answers.SSHKey != "" {
		fmt.Printf("  ssh_key: %s\n", answers.SSHKey)
	}
	if answers.Token != "" {
		fmt.Printf("  token: %s\n", maskToken(answers.Token))
	}
	survey.AskOne(&survey.Confirm{
		Message: "确认添加?",
		Default: true,
	}, &answers.Confirm)

	if !answers.Confirm {
		fmt.Println("已取消")
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
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}
	fmt.Printf("添加组 %s 到 %s\n", answers.Name, path)
}

// --- Interactive add repo (task 8.4) ---

func interactiveAddRepo() {
	// First ask: standalone or group repo
	repoType := ""
	survey.AskOne(&survey.Select{
		Message: "添加仓库类型:",
		Options: []string{"独立仓库", "添加到组"},
	}, &repoType)

	switch repoType {
	case "独立仓库":
		interactiveAddStandaloneRepo()
	case "添加到组":
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
		fmt.Fprintf(os.Stderr, "无法加载配置: %v\n", err)
		return
	}

	if len(cfg.Resources) == 0 {
		fmt.Println("请先添加资源")
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
		Message: "仓库名称:",
	}, &answers.Name)
	if answers.Name == "" {
		fmt.Println("已取消")
		return
	}

	survey.AskOne(&survey.Select{
		Message: "关联资源:",
		Options: resourceNames,
	}, &answers.Resource)

	survey.AskOne(&survey.Input{
		Message: "Clone URL:",
	}, &answers.URL)
	if answers.URL == "" {
		fmt.Println("已取消")
		return
	}

	survey.AskOne(&survey.Input{
		Message: "本地路径:",
		Default: "./" + answers.Name,
	}, &answers.LocalPath)

	// Optional SSH key
	survey.AskOne(&survey.Input{
		Message: "SSH key 路径 (可选，覆盖资源默认，留空跳过):",
	}, &answers.SSHKey)

	// Optional token
	survey.AskOne(&survey.Input{
		Message: "Token (可选，覆盖资源默认，支持 ${ENV_VAR}，留空跳过):",
	}, &answers.Token)

	// Confirmation
	fmt.Printf("\n仓库: %s\n  resource: %s\n  url: %s\n  local: %s\n",
		answers.Name, answers.Resource, answers.URL, answers.LocalPath)
	if answers.SSHKey != "" {
		fmt.Printf("  ssh_key: %s\n", answers.SSHKey)
	}
	if answers.Token != "" {
		fmt.Printf("  token: %s\n", maskToken(answers.Token))
	}
	survey.AskOne(&survey.Confirm{
		Message: "确认添加?",
		Default: true,
	}, &answers.Confirm)

	if !answers.Confirm {
		fmt.Println("已取消")
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
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}
	fmt.Printf("添加仓库 %s 到 %s\n", answers.Name, path)
}

func interactiveAddGroupRepo() {
	path, err := resolvedConfigPath()
	if err != nil {
		path = configFile
	}

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法加载配置: %v\n", err)
		return
	}

	if len(cfg.Groups) == 0 {
		fmt.Println("请先添加组")
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
		Message: "目标组:",
		Options: groupNames,
	}, &answers.Group)

	survey.AskOne(&survey.Input{
		Message: "仓库名称:",
	}, &answers.Name)
	if answers.Name == "" {
		fmt.Println("已取消")
		return
	}

	survey.AskOne(&survey.Input{
		Message: "Clone URL:",
	}, &answers.URL)
	if answers.URL == "" {
		fmt.Println("已取消")
		return
	}

	survey.AskOne(&survey.Input{
		Message: "远程路径 (组内路径，如 my-org/frontend/repo-name):",
	}, &answers.Path)

	// Confirmation
	fmt.Printf("\n组内仓库: %s → %s\n  url: %s\n  path: %s\n",
		answers.Name, answers.Group, answers.URL, answers.Path)
	survey.AskOne(&survey.Confirm{
		Message: "确认添加?",
		Default: true,
	}, &answers.Confirm)

	if !answers.Confirm {
		fmt.Println("已取消")
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
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}
	fmt.Printf("添加仓库 %s 到组 %s\n", answers.Name, answers.Group)
}

// --- Interactive sync (task 8.5) ---

func interactiveSync() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	configPath, err := resolvedConfigPath()
	if err != nil {
		configPath = configFile
	}

	// Determine scope
	scope := ""
	scopeOptions := []string{"全部", "按组", "按资源"}
	survey.AskOne(&survey.Select{
		Message: "同步范围:",
		Options: scopeOptions,
	}, &scope)

	var groupsToProcess []config.Group

	switch scope {
	case "全部":
		groupsToProcess = cfg.Groups
	case "按组":
		if len(cfg.Groups) == 0 {
			fmt.Println("没有配置组")
			return
		}
		var groupNames []string
		for _, g := range cfg.Groups {
			groupNames = append(groupNames, g.Name)
		}
		selectedGroup := ""
		survey.AskOne(&survey.Select{
			Message: "选择组:",
			Options: groupNames,
		}, &selectedGroup)
		for _, g := range cfg.Groups {
			if g.Name == selectedGroup {
				groupsToProcess = append(groupsToProcess, g)
			}
		}
	case "按资源":
		if len(cfg.Resources) == 0 {
			fmt.Println("没有配置资源")
			return
		}
		var resourceNames []string
		for name := range cfg.Resources {
			resourceNames = append(resourceNames, name)
		}
		selectedResource := ""
		survey.AskOne(&survey.Select{
			Message: "选择资源:",
			Options: resourceNames,
		}, &selectedResource)
		for _, g := range cfg.Groups {
			if g.Resource == selectedResource {
				groupsToProcess = append(groupsToProcess, g)
			}
		}
	}

	if len(groupsToProcess) == 0 {
		fmt.Println("没有需要同步的组")
		return
	}

	var totalRepos, totalNewRepos int

	for _, g := range groupsToProcess {
		res, ok := cfg.Resources[g.Resource]
		if !ok {
			fmt.Fprintf(os.Stderr, "错误: 组 %q: 资源 %q 未找到\n", g.Name, g.Resource)
			continue
		}

		p, err := provider.Get(res.Provider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 组 %q: %v\n", g.Name, err)
			continue
		}

		config.Verbose("同步组 %q (资源: %s, 路径: %s)", g.Name, g.Resource, g.Path)

		params := provider.ListReposParams{
			ServerURL: res.URL,
			Token:     res.Token,
		}

		if res.Provider == "github" {
			params.Orgs = []string{g.Path}
		} else {
			params.Groups = []provider.GroupQuery{{Path: g.Path, Recursive: g.Recursive}}
		}

		repos, err := p.ListRepos(context.Background(), params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 组 %q: %v\n", g.Name, err)
			continue
		}

		totalRepos += len(repos)
		fmt.Printf("组 %q: 发现 %d 个仓库\n", g.Name, len(repos))

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
				fmt.Fprintf(os.Stderr, "错误: 组 %q: %v\n", g.Name, err)
			} else if added > 0 {
				totalNewRepos += added
				fmt.Printf("  添加 %d 个新仓库到组 %q\n", added, g.Name)
			}
		}
	}

	fmt.Printf("\n同步完成: 发现 %d 个仓库, 新增 %d 个\n", totalRepos, totalNewRepos)
}

// --- Interactive clone (task 8.6) ---

func interactiveClone() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	// Determine scope
	scope := ""
	scopeOptions := []string{"全部", "按组", "按资源"}
	survey.AskOne(&survey.Select{
		Message: "克隆范围:",
		Options: scopeOptions,
	}, &scope)

	filter := repo.Filter{}

	switch scope {
	case "按组":
		if len(cfg.Groups) == 0 {
			fmt.Println("没有配置组")
			return
		}
		var groupNames []string
		for _, g := range cfg.Groups {
			groupNames = append(groupNames, g.Name)
		}
		selectedGroup := ""
		survey.AskOne(&survey.Select{
			Message: "选择组:",
			Options: groupNames,
		}, &selectedGroup)
		filter.Group = selectedGroup
	case "按资源":
		if len(cfg.Resources) == 0 {
			fmt.Println("没有配置资源")
			return
		}
		var resourceNames []string
		for name := range cfg.Resources {
			resourceNames = append(resourceNames, name)
		}
		selectedResource := ""
		survey.AskOne(&survey.Select{
			Message: "选择资源:",
			Options: resourceNames,
		}, &selectedResource)
		filter.Resource = selectedResource
	}

	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	if len(repos) == 0 {
		fmt.Println("没有需要克隆的仓库")
		return
	}

	// Ask for concurrency
	concurrency := 4
	survey.AskOne(&survey.Select{
		Message: "并行度:",
		Options: []string{"1 (顺序)", "2", "4 (默认)", "8"},
	}, &scope)
	switch scope {
	case "1 (顺序)":
		concurrency = 1
	case "2":
		concurrency = 2
	case "4 (默认)":
		concurrency = 4
	case "8":
		concurrency = 8
	}

	// Filter out already-cloned repos
	var toClone []gitpkg.CloneTask
	for _, r := range repos {
		fullPath := repo.FullPath(cfg.Base, r)
		if gitpkg.IsCloned(fullPath) {
			fmt.Printf("跳过 %s (已克隆)\n", r.Path)
			continue
		}
		toClone = append(toClone, gitpkg.CloneTask{
			Repo:     r,
			FullPath: fullPath,
		})
	}

	if len(toClone) == 0 {
		fmt.Println("所有仓库已克隆")
		return
	}

	if concurrency > 1 && len(toClone) > 1 {
		// Parallel clone
		progress := NewProgressRenderer("cloning", len(toClone))
		defer progress.Done()

		results := gitpkg.CloneAll(concurrency, toClone)
		progress.Update(len(results))
		PrintCloneSummary(results, nil)
	} else {
		// Sequential clone
		for _, task := range toClone {
			fmt.Printf("克隆 %s...\n", task.Repo.Path)
			opts := gitpkg.CloneOptions{
				Token:          task.Repo.Token,
				Provider:       task.Repo.Provider,
				SSHKey:         task.Repo.SSHKey,
				HasGroupToken:  task.Repo.HasGroupToken,
				HasGroupSSHKey: task.Repo.HasGroupSSHKey,
			}
			if err := gitpkg.Clone(task.FullPath, task.Repo.SSHURL, task.Repo.CloneURL, opts); err != nil {
				fmt.Fprintf(os.Stderr, "克隆 %s 失败: %v\n", task.Repo.Path, err)
				continue
			}
			fmt.Printf("  %s 完成\n", task.Repo.Name)
		}
	}
}

// --- Interactive pull ---

func interactivePull() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	// Determine scope
	scope := ""
	scopeOptions := []string{"全部", "按组", "按资源"}
	survey.AskOne(&survey.Select{
		Message: "拉取范围:",
		Options: scopeOptions,
	}, &scope)

	filter := repo.Filter{}

	switch scope {
	case "按组":
		if len(cfg.Groups) == 0 {
			fmt.Println("没有配置组")
			return
		}
		var groupNames []string
		for _, g := range cfg.Groups {
			groupNames = append(groupNames, g.Name)
		}
		selectedGroup := ""
		survey.AskOne(&survey.Select{
			Message: "选择组:",
			Options: groupNames,
		}, &selectedGroup)
		filter.Group = selectedGroup
	case "按资源":
		if len(cfg.Resources) == 0 {
			fmt.Println("没有配置资源")
			return
		}
		var resourceNames []string
		for name := range cfg.Resources {
			resourceNames = append(resourceNames, name)
		}
		selectedResource := ""
		survey.AskOne(&survey.Select{
			Message: "选择资源:",
			Options: resourceNames,
		}, &selectedResource)
		filter.Resource = selectedResource
	}

	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	if len(repos) == 0 {
		fmt.Println("没有找到仓库")
		return
	}

	// Ask for concurrency
	concurrency := 4
	survey.AskOne(&survey.Select{
		Message: "并行度:",
		Options: []string{"1 (顺序)", "2", "4 (默认)", "8"},
	}, &scope)
	switch scope {
	case "1 (顺序)":
		concurrency = 1
	case "2":
		concurrency = 2
	case "4 (默认)":
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
			fmt.Printf("跳过 %s (未克隆)\n", r.Path)
			skipped++
			continue
		}
		ok, reason := gitpkg.CheckPullSafety(fullPath)
		if !ok {
			fmt.Printf("跳过 %s (%s)\n", r.Path, reason)
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
			fmt.Printf("没有需要拉取的仓库: %d 个被跳过\n", skipped)
		} else {
			fmt.Println("没有需要拉取的仓库")
		}
		return
	}

	if concurrency > 1 && len(toPull) > 1 {
		progress := NewProgressRenderer("pulling", len(toPull))
		defer progress.Done()

		results := gitpkg.PullAll(concurrency, toPull)
		progress.Update(len(results))
		PrintPullSummary(results, skipped, nil)
	} else {
		for _, task := range toPull {
			fmt.Printf("拉取 %s...\n", task.Repo.Path)
			if err := gitpkg.Pull(task.FullPath); err != nil {
				fmt.Fprintf(os.Stderr, "拉取 %s 失败: %v\n", task.Repo.Path, err)
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
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	// Determine scope
	scope := ""
	scopeOptions := []string{"全部", "按组", "按资源"}
	survey.AskOne(&survey.Select{
		Message: "查看范围:",
		Options: scopeOptions,
	}, &scope)

	filter := repo.Filter{}

	switch scope {
	case "按组":
		if len(cfg.Groups) == 0 {
			fmt.Println("没有配置组")
			return
		}
		var groupNames []string
		for _, g := range cfg.Groups {
			groupNames = append(groupNames, g.Name)
		}
		selectedGroup := ""
		survey.AskOne(&survey.Select{
			Message: "选择组:",
			Options: groupNames,
		}, &selectedGroup)
		filter.Group = selectedGroup
	case "按资源":
		if len(cfg.Resources) == 0 {
			fmt.Println("没有配置资源")
			return
		}
		var resourceNames []string
		for name := range cfg.Resources {
			resourceNames = append(resourceNames, name)
		}
		selectedResource := ""
		survey.AskOne(&survey.Select{
			Message: "选择资源:",
			Options: resourceNames,
		}, &selectedResource)
		filter.Resource = selectedResource
	}

	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	if len(repos) == 0 {
		fmt.Println("没有找到仓库")
		return
	}

	for _, r := range repos {
		fullPath := repo.FullPath(cfg.Base, r)

		st := gitpkg.GetStatus(fullPath)

		if !st.Cloned {
			fmt.Printf("%s: 未克隆\n", r.Path)
			continue
		}

		if st.NotARepo {
			fmt.Printf("%s: 不是 git 仓库\n", r.Path)
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
