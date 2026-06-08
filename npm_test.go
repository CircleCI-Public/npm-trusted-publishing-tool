package main

import (
	"strings"
	"testing"
)

func TestTrustConfigArgs(t *testing.T) {
	base := trustConfig{
		OrgID:       "org-1",
		ProjectID:   "proj-1",
		PipelineDef: "def-1",
		VCSOrigin:   "github.com/acme/widget",
	}

	t.Run("required flags, no package", func(t *testing.T) {
		got := strings.Join(base.args(), " ")
		want := "trust circleci --org-id org-1 --project-id proj-1 " +
			"--pipeline-definition-id def-1 --vcs-origin github.com/acme/widget --yes"
		if got != want {
			t.Errorf("args = %q, want %q", got, want)
		}
	})

	t.Run("package is positional before flags", func(t *testing.T) {
		c := base
		c.Package = "@acme/widget"
		args := c.args()
		if args[0] != "trust" || args[1] != "circleci" || args[2] != "@acme/widget" {
			t.Errorf("package not placed positionally: %v", args[:3])
		}
	})

	t.Run("contexts, both permissions, dry-run", func(t *testing.T) {
		c := base
		c.ContextIDs = []string{"ctx-a", "ctx-b"}
		c.AllowPublish = true
		c.AllowStage = true
		c.DryRun = true
		got := strings.Join(c.args(), " ")
		for _, want := range []string{
			"--context-id ctx-a", "--context-id ctx-b",
			"--allow-publish", "--allow-stage-publish", "--dry-run",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("args missing %q; got %q", want, got)
			}
		}
	})

	t.Run("omitted permissions are absent", func(t *testing.T) {
		got := strings.Join(base.args(), " ")
		for _, absent := range []string{"--allow-publish", "--allow-stage-publish", "--dry-run", "--context-id"} {
			if strings.Contains(got, absent) {
				t.Errorf("args unexpectedly contains %q; got %q", absent, got)
			}
		}
	})

	t.Run("json is not passed (would break web 2FA)", func(t *testing.T) {
		if strings.Contains(strings.Join(base.args(), " "), "--json") {
			t.Error("args must not contain --json; it requires capturing stdout, which disables npm's web 2FA")
		}
	})
}
