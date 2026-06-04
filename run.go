package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
)

// runState carries choices remembered across iterations. Publish permissions are
// asked once for the whole run; context choices are scoped to one org and reset
// at each org boundary by the run loop.
type runState struct {
	publishAsked bool
	allowPublish bool
	allowStage   bool

	contextsDecided  bool
	reuseContexts    bool
	sharedContextIDs []string
}

type result struct {
	label  string
	ok     bool
	detail string
}

// run processes every project and returns the process exit code.
func run(projectsFile string, dryRun bool) int {
	entries, err := resolveProjects(projectsFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	groups := groupByOrg(entries)
	state := &runState{}
	var results []result
	total := len(entries)
	multi := total > 1
	n := 0

	for _, g := range groups {
		// Context choices are per-org: reset at each org boundary.
		state.contextsDecided = false
		state.reuseContexts = false
		state.sharedContextIDs = nil

		for _, e := range g.entries {
			n++
			label := e.Slug
			if label == "" {
				label = "current directory"
			}
			if multi {
				fmt.Printf("\n=== [%d/%d] %s ===\n", n, total, label)
			}

			detail, err := process(e, len(g.entries), state, dryRun)
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					fmt.Fprintln(os.Stderr, "aborted")
					return 130
				}
				fmt.Fprintf(os.Stderr, "  failed: %v\n", err)
				results = append(results, result{label: label, ok: false, detail: err.Error()})
				continue
			}
			results = append(results, result{label: label, ok: true, detail: detail})
		}
	}

	printSummary(results, dryRun)
	for _, r := range results {
		if !r.ok {
			return 1
		}
	}
	return 0
}

// resolveProjects returns the entries to process. With no file (or an empty
// file) it returns a single empty entry: single-directory mode.
func resolveProjects(path string) ([]projectEntry, error) {
	if path == "" {
		return []projectEntry{{}}, nil
	}
	entries, err := loadProjects(path)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return []projectEntry{{}}, nil
	}
	return entries, nil
}

// process handles a single project end to end. orgCount is the number of
// projects in the current org, used to decide whether to offer context reuse.
func process(e projectEntry, orgCount int, state *runState, dryRun bool) (string, error) {
	proj, err := getProject(e.Slug)
	if err != nil {
		return "", err
	}

	defs, err := listPipelineDefinitions(proj.Slug)
	if err != nil {
		return "", err
	}
	def, err := selectPipelineDefinition(defs)
	if err != nil {
		return "", err
	}

	contextIDs, err := resolveContexts(proj, orgCount, state)
	if err != nil {
		return "", err
	}

	if !state.publishAsked {
		ap, as, err := selectPublishMode()
		if err != nil {
			return "", err
		}
		state.allowPublish, state.allowStage, state.publishAsked = ap, as, true
	}

	cfg := trustConfig{
		Package:      e.Package, // empty in single-dir mode -> npm uses package.json
		OrgID:        proj.OrganizationID,
		ProjectID:    proj.ID,
		PipelineDef:  def.ID,
		VCSOrigin:    vcsOrigin(proj.VCSURL),
		ContextIDs:   contextIDs,
		AllowPublish: state.allowPublish,
		AllowStage:   state.allowStage,
		DryRun:       dryRun,
	}

	// npm writes its output (incl. 2FA prompt and results) straight to the
	// terminal; success is judged by exit code.
	if err := runNpmTrust(cfg); err != nil {
		return "", fmt.Errorf("npm trust: %w", err)
	}
	return "", nil
}

// resolveContexts returns the context IDs for a project, honoring the
// reuse-for-this-org decision. State is reset by the caller at each org boundary.
func resolveContexts(proj Project, orgCount int, state *runState) ([]string, error) {
	if state.contextsDecided && state.reuseContexts {
		return state.sharedContextIDs, nil
	}

	ctxs, err := listContexts(proj.OrganizationSlug)
	if err != nil {
		return nil, err
	}
	ids, err := selectContexts(ctxs)
	if err != nil {
		return nil, err
	}

	// On the first project of a multi-project org, offer to reuse this selection
	// for the rest of the org.
	if orgCount > 1 && !state.contextsDecided && len(ctxs) > 0 {
		reuse, err := confirm("Use these contexts for all remaining projects in this org?")
		if err != nil {
			return nil, err
		}
		state.contextsDecided = true
		state.reuseContexts = reuse
		if reuse {
			state.sharedContextIDs = ids
		}
	}
	return ids, nil
}

func printSummary(results []result, dryRun bool) {
	fmt.Println("\nSummary:")
	if dryRun {
		fmt.Println("(dry run — no changes were made)")
	}
	for _, r := range results {
		status := "ok"
		if !r.ok {
			status = "FAILED"
		}
		line := fmt.Sprintf("  [%s] %s", status, r.label)
		if r.detail != "" {
			line += " — " + r.detail
		}
		fmt.Println(line)
	}
}
