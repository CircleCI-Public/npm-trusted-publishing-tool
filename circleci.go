package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Project is the subset of `circleci project get --json` we use.
type Project struct {
	ID               string `json:"id"`
	Slug             string `json:"slug"`
	Name             string `json:"name"`
	OrganizationID   string `json:"organization_id"`
	OrganizationSlug string `json:"organization_slug"`
	VCSURL           string `json:"vcs_url"`
}

// PipelineDefinition is the subset of `circleci pipeline list --json` we use.
type PipelineDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Context is the subset of `circleci context list --json` we use.
type Context struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// runCircleCI runs a circleci subcommand and returns its stdout. Telemetry and
// informational output go to stderr, so capturing stdout alone yields clean JSON.
func runCircleCI(args ...string) ([]byte, error) {
	cmd := exec.Command("circleci", args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	cmd.Env = append(os.Environ(), "CIRCLECI_CLI_TELEMETRY_OPTOUT=1")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("circleci %s: %s", strings.Join(args, " "), circleciError(errb.Bytes(), out.Bytes(), err))
	}
	return out.Bytes(), nil
}

// circleciError extracts a concise message from circleci's structured error
// output (on stderr or stdout), falling back to the raw process error.
func circleciError(stderr, stdout []byte, runErr error) string {
	for _, b := range [][]byte{stderr, stdout} {
		var e struct {
			Message     string   `json:"message"`
			Suggestions []string `json:"suggestions"`
		}
		if json.Unmarshal(b, &e) == nil && e.Message != "" {
			if len(e.Suggestions) > 0 {
				return e.Message + " (" + strings.Join(e.Suggestions, "; ") + ")"
			}
			return e.Message
		}
	}
	if s := strings.TrimSpace(string(stderr)); s != "" {
		return s
	}
	return runErr.Error()
}

// getProject resolves a project. An empty slug lets circleci infer it from the
// current git remote.
func getProject(slug string) (Project, error) {
	args := []string{"project", "get", "-q", "--json"}
	if slug != "" {
		args = append(args, "--project", slug)
	}
	out, err := runCircleCI(args...)
	if err != nil {
		return Project{}, err
	}
	var p Project
	if err := json.Unmarshal(out, &p); err != nil {
		return Project{}, fmt.Errorf("parsing project: %w", err)
	}
	return p, nil
}

// listPipelineDefinitions lists pipeline definitions for a project slug.
func listPipelineDefinitions(slug string) ([]PipelineDefinition, error) {
	args := []string{"pipeline", "list", "-q", "--json"}
	if slug != "" {
		args = append(args, "--project", slug)
	}
	out, err := runCircleCI(args...)
	if err != nil {
		return nil, err
	}
	var defs []PipelineDefinition
	if err := json.Unmarshal(out, &defs); err != nil {
		return nil, fmt.Errorf("parsing pipeline definitions: %w", err)
	}
	return defs, nil
}

// listContexts lists contexts for an organization slug.
func listContexts(orgSlug string) ([]Context, error) {
	args := []string{"context", "list", "-q", "--json"}
	if orgSlug != "" {
		args = append(args, "--org", orgSlug)
	}
	out, err := runCircleCI(args...)
	if err != nil {
		return nil, err
	}
	var ctxs []Context
	if err := json.Unmarshal(out, &ctxs); err != nil {
		return nil, fmt.Errorf("parsing contexts: %w", err)
	}
	return ctxs, nil
}

// vcsOrigin returns vcs_url with the scheme stripped, e.g.
// https://github.com/acme/widget -> github.com/acme/widget. This host-style
// origin (not the gh/owner/repo slug) matches the format in CircleCI's OIDC
// token claim, which is what npm trust validates --vcs-origin against.
func vcsOrigin(vcsURL string) string {
	s := strings.TrimPrefix(vcsURL, "https://")
	s = strings.TrimPrefix(s, "http://")
	return strings.TrimSuffix(s, "/")
}
