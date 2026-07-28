package repo

import (
	"net/url"
	"strings"

	"github.com/wii/grepom/config"
)

// URLKind classifies a repo.url value for clone URL construction.
type URLKind int

const (
	// URLRelative is a path relative to the bound resource host (e.g. "org/app.git").
	URLRelative URLKind = iota
	// URLAbsoluteHTTP is a full http(s) URL.
	URLAbsoluteHTTP
	// URLAbsoluteSSH is a full SSH URL (git@ or ssh://).
	URLAbsoluteSSH
)

// ClassifyRepoURL classifies repo.url without binding to a resource.
func ClassifyRepoURL(raw string) URLKind {
	switch {
	case strings.HasPrefix(raw, "https://"), strings.HasPrefix(raw, "http://"):
		return URLAbsoluteHTTP
	case strings.HasPrefix(raw, "git@"), strings.HasPrefix(raw, "ssh://"):
		return URLAbsoluteSSH
	default:
		return URLRelative
	}
}

// ensureGitSuffix appends .git when missing.
func ensureGitSuffix(u string) string {
	if strings.HasSuffix(u, ".git") {
		return u
	}
	return u + ".git"
}

// ResolvedCloneURLs holds CloneURL / SSHURL and preferred protocol for a bound repo.
type ResolvedCloneURLs struct {
	CloneURL    string
	SSHURL      string
	PreferHTTPS bool
}

// ResolveBoundRepoURLs resolves clone URLs for a standalone repo bound to a resource.
// Absolute URLs are used as-is (no secondary concatenation onto resource.url).
func ResolveBoundRepoURLs(repoURL string, res config.Resource) ResolvedCloneURLs {
	switch ClassifyRepoURL(repoURL) {
	case URLAbsoluteHTTP:
		clone := ensureGitSuffix(repoURL)
		path := ExtractRemotePath(repoURL)
		sshURL := ""
		if path != "" {
			if host := hostFromHTTPURL(repoURL); host != "" {
				sshURL = deriveSSHURL(path, host)
			} else if res.URL != "" {
				sshURL = deriveSSHURL(path, res.URL)
			}
		}
		return ResolvedCloneURLs{
			CloneURL:    clone,
			SSHURL:      sshURL,
			PreferHTTPS: true,
		}
	case URLAbsoluteSSH:
		return ResolvedCloneURLs{
			CloneURL:    httpsURLFromSSH(repoURL),
			SSHURL:      repoURL,
			PreferHTTPS: false,
		}
	default:
		path := ExtractRemotePath(repoURL)
		return ResolvedCloneURLs{
			CloneURL:    res.HTTPSURL(path),
			SSHURL:      deriveSSHURL(path, res.URL),
			PreferHTTPS: false,
		}
	}
}

func hostFromHTTPURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Host
}

// httpsURLFromSSH converts git@host:path or ssh://git@host/path to https://host/path.git.
// Returns empty string when conversion is not possible.
func httpsURLFromSSH(raw string) string {
	path := ExtractRemotePath(raw)
	if path == "" {
		return ""
	}
	host := ""
	switch {
	case strings.HasPrefix(raw, "git@"):
		rest := strings.TrimPrefix(raw, "git@")
		if idx := strings.Index(rest, ":"); idx >= 0 {
			host = rest[:idx]
		}
	case strings.HasPrefix(raw, "ssh://"):
		u, err := url.Parse(raw)
		if err == nil {
			host = u.Hostname()
			if u.Port() != "" && u.Port() != "22" {
				host = u.Host
			}
		}
	}
	if host == "" {
		return ""
	}
	return "https://" + host + "/" + path + ".git"
}
