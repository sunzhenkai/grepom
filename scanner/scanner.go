package scanner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/semgroup"
	"github.com/spf13/viper"
	"github.com/zricethezav/gitleaks/v8/cmd/scm"
	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/report"
	"github.com/zricethezav/gitleaks/v8/sources"
)

// Options 配置 Scanner 的行为。
type Options struct {
	// GitleaksConfigPath 是自定义 gitleaks.toml 配置文件路径。
	// 为空时使用 gitleaks 默认规则集。
	GitleaksConfigPath string
	// MaxTargetMegaBytes 跳过超过此大小（MB）的文件。0 表示不限制。
	MaxTargetMegaBytes int
}

// Scanner 封装 gitleaks 检测引擎，提供对仓库进行敏感信息扫描的能力。
type Scanner struct {
	opts Options
}

// NewScanner 创建一个新的 Scanner 实例。
func NewScanner(opts Options) *Scanner {
	return &Scanner{opts: opts}
}

// loadConfig 加载 gitleaks 配置。
// 如果指定了自定义配置路径则加载该文件，否则使用默认规则集。
func (s *Scanner) loadConfig() (config.Config, error) {
	if s.opts.GitleaksConfigPath != "" {
		viper.SetConfigFile(s.opts.GitleaksConfigPath)
		viper.SetConfigType("toml")
		if err := viper.ReadInConfig(); err != nil {
			return config.Config{}, fmt.Errorf("failed to read gitleaks config: %w", err)
		}
		var vc config.ViperConfig
		if err := viper.Unmarshal(&vc); err != nil {
			return config.Config{}, fmt.Errorf("failed to parse gitleaks config: %w", err)
		}
		cfg, err := vc.Translate()
		if err != nil {
			return config.Config{}, fmt.Errorf("failed to translate gitleaks config: %w", err)
		}
		return cfg, nil
	}

	// 使用默认配置
	detector, err := detect.NewDetectorDefaultConfig()
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to load default gitleaks config: %w", err)
	}
	return detector.Config, nil
}

// newDetectorForRepo 为指定仓库创建一个 Detector 实例。
// 它会加载配置、注入 .gitignore 排除规则、加载 .gitleaksignore。
func (s *Scanner) newDetectorForRepo(ctx context.Context, repoPath string) (*detect.Detector, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}

	// 注入 .gitignore 排除规则到全局 allowlist
	gitignorePatterns := parseGitignore(repoPath)
	if len(gitignorePatterns) > 0 {
		var pathRegexps []*regexp.Regexp
		for _, p := range gitignorePatterns {
			if re, err := regexp.Compile(p); err == nil {
				pathRegexps = append(pathRegexps, re)
			}
		}
		if len(pathRegexps) > 0 {
			cfg.Allowlists = append(cfg.Allowlists, &config.Allowlist{
				Paths:          pathRegexps,
				MatchCondition: config.AllowlistMatchOr,
			})
		}
	}

	det := detect.NewDetectorContext(ctx, cfg)
	det.Redact = 0 // 不让 gitleaks redact，由 grepom 自行控制脱敏

	if s.opts.MaxTargetMegaBytes > 0 {
		det.MaxTargetMegaBytes = s.opts.MaxTargetMegaBytes
	}

	// 加载 .gitleaksignore
	gitleaksignorePath := filepath.Join(repoPath, ".gitleaksignore")
	if _, err := os.Stat(gitleaksignorePath); err == nil {
		if err := det.AddGitleaksIgnore(gitleaksignorePath); err != nil {
			// 非致命错误，仅记录
			fmt.Fprintf(os.Stderr, "warning: failed to read .gitleaksignore: %v\n", err)
		}
	}

	return det, nil
}

// ScanDir 扫描指定路径的工作区文件，返回发现的所有敏感信息。
func (s *Scanner) ScanDir(ctx context.Context, repoPath string) ([]Finding, error) {
	det, err := s.newDetectorForRepo(ctx, repoPath)
	if err != nil {
		return nil, err
	}

	sema := semgroup.NewGroup(ctx, 40)

	files := &sources.Files{
		Path:   repoPath,
		Sema:   sema,
		Config: &det.Config,
	}

	findings, err := det.DetectSource(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return convertFindings(findings, ""), nil
}

// ScanGitHistory 扫描指定仓库的 git 历史（所有分支的所有提交），返回发现的所有敏感信息。
func (s *Scanner) ScanGitHistory(ctx context.Context, repoPath string) ([]Finding, error) {
	det, err := s.newDetectorForRepo(ctx, repoPath)
	if err != nil {
		return nil, err
	}

	gitCmd, err := sources.NewGitLogCmdContext(ctx, repoPath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to start git log command: %w", err)
	}

	gitSource := &sources.Git{
		Cmd:    gitCmd,
		Config: &det.Config,
		Remote: sources.NewRemoteInfoContext(ctx, scm.NoPlatform, repoPath),
		Sema:   semgroup.NewGroup(ctx, 40),
	}

	findings, err := det.DetectSource(ctx, gitSource)
	if err != nil {
		return nil, fmt.Errorf("failed to scan git history: %w", err)
	}

	return convertFindings(findings, ""), nil
}

// convertFindings 将 gitleaks 的 report.Finding 列表转换为 grepom 的 scanner.Finding 列表。
func convertFindings(findings []report.Finding, repoName string) []Finding {
	result := make([]Finding, 0, len(findings))
	for _, f := range findings {
		result = append(result, Finding{
			Repo:        repoName,
			File:        f.File,
			Line:        f.StartLine,
			RuleID:      f.RuleID,
			Description: f.Description,
			Secret:      f.Secret,
			Severity:    severityFromTags(f.Tags),
		})
	}
	return result
}

// severityFromTags 根据规则的 tags 推断严重程度。
// gitleaks 默认规则带有 key、token、file 等标签，
// 我们根据标签中的关键词映射严重程度。
func severityFromTags(tags []string) Severity {
	for _, tag := range tags {
		t := strings.ToLower(tag)
		switch {
		case strings.Contains(t, "key"):
			return SeverityCritical
		case strings.Contains(t, "token"):
			return SeverityHigh
		case strings.Contains(t, "secret"):
			return SeverityHigh
		case strings.Contains(t, "password"):
			return SeverityHigh
		case strings.Contains(t, "credential"):
			return SeverityHigh
		}
	}
	return SeverityMedium
}

// parseGitignore 读取仓库根目录下的 .gitignore 文件，
// 将每行转换为可用于 gitleaks allowlist 的正则表达式模式。
// 如果文件不存在或为空，返回 nil。
func parseGitignore(repoPath string) []string {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		return nil
	}

	var patterns []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 转换 .gitignore 模式为正则
		re := gitignoreToRegex(line)
		if re != "" {
			patterns = append(patterns, re)
		}
	}
	return patterns
}

// gitignoreToRegex 将单个 .gitignore 模式转换为正则表达式。
// 这是一个简化实现，覆盖最常见的 .gitignore 模式。
func gitignoreToRegex(pattern string) string {
	// 去掉前导的 !
	if strings.HasPrefix(pattern, "!") {
		return "" // 忽略 "取反" 模式，不做特殊处理
	}

	p := pattern

	// 目录模式：以 / 结尾
	isDir := strings.HasSuffix(p, "/")
	if isDir {
		p = strings.TrimSuffix(p, "/")
	}

	// 处理前导 /
	if strings.HasPrefix(p, "/") {
		p = "^" + regexp.QuoteMeta(p[1:])
	} else {
		// 没有前导 / 的模式可以匹配任意层级
		p = "(.*/)?" + regexp.QuoteMeta(p)
	}

	// 处理 ** 通配符（先替换 ** 再替换 *）
	p = strings.ReplaceAll(p, regexp.QuoteMeta("**"), ".*")
	p = strings.ReplaceAll(p, regexp.QuoteMeta("*"), "[^/]*")
	p = strings.ReplaceAll(p, regexp.QuoteMeta("?"), "[^/]")

	if isDir {
		p = p + "/.*"
	} else {
		p = p + "($|/.*)"
	}

	// 尝试编译验证
	if _, err := regexp.Compile(p); err != nil {
		return ""
	}
	return p
}
