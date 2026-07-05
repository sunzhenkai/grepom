package provider

import "testing"

func TestIsDeletionScheduled(t *testing.T) {
	tests := []struct {
		name              string
		repoName          string
		pathWithNamespace string
		want              bool
	}{
		{
			name:              "name 含 deletion_scheduled 标记",
			repoName:          "creative-matching-deletion_scheduled-499",
			pathWithNamespace: "dsp/creative-matching-deletion_scheduled-499",
			want:              true,
		},
		{
			name:              "仅 pathWithNamespace 含标记（组被删除）",
			repoName:          "some-repo",
			pathWithNamespace: "dsp-services-deletion_scheduled-452/some-repo",
			want:              true,
		},
		{
			name:              "name 与 path 均含标记",
			repoName:          "ranker-deletion_scheduled-497",
			pathWithNamespace: "dsp-services-deletion_scheduled-452/ranker-deletion_scheduled-497",
			want:              true,
		},
		{
			name:              "正常代码库",
			repoName:          "grepom",
			pathWithNamespace: "wii/solo/grepom",
			want:              false,
		},
		{
			name:              "name 仅部分匹配（不含完整标记）",
			repoName:          "deletion-handler",
			pathWithNamespace: "org/deletion-handler",
			want:              false,
		},
		{
			name:              "空值",
			repoName:          "",
			pathWithNamespace: "",
			want:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDeletionScheduled(tt.repoName, tt.pathWithNamespace); got != tt.want {
				t.Errorf("IsDeletionScheduled(%q, %q) = %v, want %v",
					tt.repoName, tt.pathWithNamespace, got, tt.want)
			}
		})
	}
}
