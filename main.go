// Command npm-trusted-publishing-tool onboards CircleCI projects to npm trusted
// publishing by orchestrating the circleci and npm CLIs.
package main

import (
	"flag"
	"fmt"
	"os"
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

// version is set at build time via -ldflags "-X main.version=...".
// GoReleaser populates it for released binaries; it stays "dev" otherwise.
var version = "dev"

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
		fmt.Println(version)
		return
	}

	if help {
		fmt.Print(usage)
		return
	}

	os.Exit(run(projectsFile, dryRun))
}
