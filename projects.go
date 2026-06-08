package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// projectEntry is one line from the --projects file.
type projectEntry struct {
	Package string // npm package name
	Slug    string // CircleCI project slug, e.g. gh/acme/widget
}

// orgGroup is a set of entries that share an organization, in input order.
type orgGroup struct {
	org     string
	entries []projectEntry
}

// orgFromSlug derives the org slug from a project slug, e.g.
// gh/acme/widget -> gh/acme. Empty or single-segment slugs are returned as-is.
func orgFromSlug(slug string) string {
	parts := strings.Split(slug, "/")
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return slug
}

// groupByOrg clusters entries by org, preserving the first-appearance order of
// both orgs and the entries within each org. Processing org-by-org lets context
// selections be reused safely within an org and re-prompted when the org changes.
func groupByOrg(entries []projectEntry) []orgGroup {
	var groups []orgGroup
	idx := make(map[string]int)
	for _, e := range entries {
		org := orgFromSlug(e.Slug)
		i, ok := idx[org]
		if !ok {
			i = len(groups)
			idx[org] = i
			groups = append(groups, orgGroup{org: org})
		}
		groups[i].entries = append(groups[i].entries, e)
	}
	return groups
}

// loadProjects reads a --projects file. Each non-empty, non-comment line is
// "package-name,project-slug". Returns the parsed entries.
func loadProjects(path string) ([]projectEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []projectEntry
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		parts := strings.SplitN(text, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: expected \"package-name,project-slug\", got %q", line, text)
		}
		pkg := strings.TrimSpace(parts[0])
		slug := strings.TrimSpace(parts[1])
		if pkg == "" || slug == "" {
			return nil, fmt.Errorf("line %d: package and slug must both be set, got %q", line, text)
		}
		entries = append(entries, projectEntry{Package: pkg, Slug: slug})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
