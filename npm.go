package main

import (
	"os"
	"os/exec"
)

// trustConfig holds everything needed for one `npm trust circleci` call.
type trustConfig struct {
	Package      string // optional; empty => npm reads ./package.json
	OrgID        string
	ProjectID    string
	PipelineDef  string
	VCSOrigin    string
	ContextIDs   []string
	AllowPublish bool
	AllowStage   bool
	DryRun       bool
}

func (c trustConfig) args() []string {
	args := []string{"trust", "circleci"}
	if c.Package != "" {
		args = append(args, c.Package)
	}
	args = append(args,
		"--org-id", c.OrgID,
		"--project-id", c.ProjectID,
		"--pipeline-definition-id", c.PipelineDef,
		"--vcs-origin", c.VCSOrigin,
	)
	for _, id := range c.ContextIDs {
		args = append(args, "--context-id", id)
	}
	if c.AllowPublish {
		args = append(args, "--allow-publish")
	}
	if c.AllowStage {
		args = append(args, "--allow-stage-publish")
	}
	if c.DryRun {
		args = append(args, "--dry-run")
	}
	return append(args, "--yes")
}

// runNpmTrust runs `npm trust circleci`, inheriting all three std streams.
//
// npm's web-based 2FA only runs when both stdin and stdout are TTYs (see
// npm/cli lib/utils/auth.js otplease), so we must NOT capture stdout — npm
// writes its prompts, the auth URL, and results straight to the terminal. The
// first call in a run prompts for 2FA via the browser; npm offers a "skip 2FA
// for 5 minutes" option there, after which later calls proceed without it.
// Success is determined by exit code.
func runNpmTrust(c trustConfig) error {
	cmd := exec.Command("npm", c.args()...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
