package repo

import (
	"testing"

	"github.com/wii/grepom/provider"
)

func TestApplyFilter_ByName(t *testing.T) {
	repos := []provider.Repo{
		{Name: "web-app", Path: "org/frontend/web-app", Provider: "gitlab"},
		{Name: "api-server", Path: "org/backend/api-server", Provider: "gitlab"},
	}

	result := ApplyFilter(repos, Filter{Name: "web-app"})
	if len(result) != 1 || result[0].Name != "web-app" {
		t.Error("name filter failed")
	}
}

func TestApplyFilter_ByGroup(t *testing.T) {
	repos := []provider.Repo{
		{Name: "web-app", Path: "org/frontend/web-app", Provider: "gitlab"},
		{Name: "api-server", Path: "org/backend/api-server", Provider: "gitlab"},
	}

	result := ApplyFilter(repos, Filter{Group: "org/frontend"})
	if len(result) != 1 || result[0].Name != "web-app" {
		t.Error("group filter failed")
	}
}

func TestApplyFilter_ByProvider(t *testing.T) {
	repos := []provider.Repo{
		{Name: "web-app", Path: "org/frontend/web-app", Provider: "gitlab"},
		{Name: "api-server", Path: "org/api-server", Provider: "github"},
	}

	result := ApplyFilter(repos, Filter{Provider: "github"})
	if len(result) != 1 || result[0].Name != "api-server" {
		t.Error("provider filter failed")
	}
}

func TestApplyFilter_NoFilter(t *testing.T) {
	repos := []provider.Repo{
		{Name: "web-app", Path: "org/frontend/web-app", Provider: "gitlab"},
		{Name: "api-server", Path: "org/api-server", Provider: "github"},
	}

	result := ApplyFilter(repos, Filter{})
	if len(result) != 2 {
		t.Error("no filter should return all repos")
	}
}

func TestApplyFilter_Combined(t *testing.T) {
	repos := []provider.Repo{
		{Name: "web-app", Path: "org/frontend/web-app", Provider: "gitlab"},
		{Name: "api-server", Path: "org/frontend/api-server", Provider: "github"},
	}

	result := ApplyFilter(repos, Filter{Group: "org/frontend", Provider: "gitlab"})
	if len(result) != 1 || result[0].Name != "web-app" {
		t.Error("combined filter failed")
	}
}

func TestFullPath(t *testing.T) {
	r := provider.Repo{Path: "org/frontend/web-app"}
	result := FullPath("/home/user/projects", r)
	expected := "/home/user/projects/org/frontend/web-app"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
