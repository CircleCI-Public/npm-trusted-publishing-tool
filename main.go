// Command npm-trusted-publishing-tool onboards CircleCI projects to npm trusted
// publishing by orchestrating the circleci and npm CLIs.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
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

// Build metadata, set via -ldflags "-X main.version=... -X main.commit=...".
// GoReleaser populates both for released binaries. For a plain `go build` the
// commit is recovered from the embedded VCS info and version stays "dev".
var (
	version = "dev"
	commit  = ""
)

// buildVersion returns the version with a short commit hash when one is
// available, e.g. "dev (a1b2c3d)" or "0.0.0-main.20260608.b64c21b (b64c21b)".
func buildVersion() string {
	c := commit
	if c == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, s := range info.Settings {
				if s.Key == "vcs.revision" {
					c = s.Value
				}
			}
		}
	}
	if len(c) > 7 {
		c = c[:7]
	}
	if c == "" {
		return version
	}
	return version + " (" + c + ")"
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
