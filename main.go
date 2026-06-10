// Command npm-trusted-publishing-tool onboards CircleCI projects to npm trusted
// publishing by orchestrating the circleci and npm CLIs.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

const usage = `npm trusted publishing onboarding

Configures npm trusted publishing for one or more CircleCI projects via
` + "`npm trust circleci`" + `.

Usage:
  npm-trusted-publishing-tool [flags]

Flags:
  --projects <file>  List of projects to process, one per line as
                     "package-name,project-slug" (e.g. @acme/widget,gh/acme/widget).
                     Lines starting with # are ignored.
                     If omitted, runs once in the current directory (npm reads the
                     package from ./package.json; circleci infers the project from
                     the git remote).
  --dry-run          Pass --dry-run to npm trust; makes no changes.
  --version          Show version and exit.
  -h, --help         Show this help.

Notes:
  Requires the circleci and npm CLIs on PATH and an authenticated circleci.
  Projects are grouped by organization and processed org-by-org. Contexts are
  organization-scoped: you choose contexts per org and may reuse them for the
  rest of that org; the choice is re-prompted when the org changes.
`

// version is the release tag, set via -ldflags "-X main.version=...". GoReleaser
// sets it for released binaries; it stays "dev" otherwise. The commit hash and
// dirty-tree flag come from the binary's embedded VCS info (debug.ReadBuildInfo),
// the same way the circleci CLI does it, so no separate -X main.commit is needed.
var version = "dev"

// buildVersion returns "<version> (<commit>)", e.g. "1.0.42 (a1b2c3d)", with a
// "-dirty" marker when built from an uncommitted tree. The commit is appended
// unless the version string already contains it.
func buildVersion() string {
	var commit string
	var modified bool
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				commit = s.Value
			case "vcs.modified":
				modified = s.Value == "true"
			}
		}
	}
	if len(commit) > 7 {
		commit = commit[:7]
	}
	if commit == "" || (strings.Contains(version, commit) && !modified) {
		return version
	}
	if modified {
		commit += "-dirty"
	}
	return version + " (" + commit + ")"
}

func main() {
	var projectsFile string
	var dryRun, help, showVersion bool
	flag.StringVar(&projectsFile, "projects", "", "path to projects list file")
	flag.BoolVar(&dryRun, "dry-run", false, "pass --dry-run to npm trust")
	flag.BoolVar(&showVersion, "version", false, "show version and exit")
	flag.BoolVar(&help, "help", false, "show help")
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	if showVersion {
		fmt.Println(buildVersion())
		return
	}

	if help {
		fmt.Print(usage)
		return
	}

	os.Exit(run(projectsFile, dryRun))
}
