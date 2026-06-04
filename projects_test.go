package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "projects.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadProjects(t *testing.T) {
	content := "# a comment\n" +
		"@acme/widget,gh/acme/widget\n" +
		"\n" +
		"  @acme/gadget , gh/acme/gadget  \n"
	got, err := loadProjects(writeTemp(t, content))
	if err != nil {
		t.Fatalf("loadProjects() error = %v", err)
	}
	want := []projectEntry{
		{Package: "@acme/widget", Slug: "gh/acme/widget"},
		{Package: "@acme/gadget", Slug: "gh/acme/gadget"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestLoadProjectsErrors(t *testing.T) {
	tests := []struct {
		name, content string
	}{
		{"missing comma", "badline-no-comma\n"},
		{"empty package", ",gh/acme/widget\n"},
		{"empty slug", "@acme/widget,\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := loadProjects(writeTemp(t, tt.content)); err == nil {
				t.Errorf("loadProjects(%q) expected error, got nil", tt.content)
			}
		})
	}
}

func TestLoadProjectsMissingFile(t *testing.T) {
	if _, err := loadProjects(filepath.Join(t.TempDir(), "nope.txt")); err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// resolveProjects with no path yields a single empty entry (single-dir mode).
func TestResolveProjectsSingleDir(t *testing.T) {
	entries, err := resolveProjects("")
	if err != nil {
		t.Fatalf("resolveProjects(\"\") error = %v", err)
	}
	if len(entries) != 1 || entries[0] != (projectEntry{}) {
		t.Errorf("got entries=%+v, want one empty entry in single-dir mode", entries)
	}
}

// An empty (comments-only) file also falls back to single-dir mode.
func TestResolveProjectsEmptyFile(t *testing.T) {
	entries, err := resolveProjects(writeTemp(t, "# only comments\n\n"))
	if err != nil {
		t.Fatalf("resolveProjects() error = %v", err)
	}
	if len(entries) != 1 || entries[0] != (projectEntry{}) {
		t.Errorf("got entries=%+v, want single-dir fallback", entries)
	}
}

func TestOrgFromSlug(t *testing.T) {
	tests := []struct{ in, want string }{
		{"gh/acme/widget", "gh/acme"},
		{"bitbucket/acme/gadget", "bitbucket/acme"},
		{"gh/acme", "gh/acme"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := orgFromSlug(tt.in); got != tt.want {
			t.Errorf("orgFromSlug(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// groupByOrg clusters by org while preserving first-appearance order of orgs and
// of entries within each org, even when orgs are interleaved in the input.
func TestGroupByOrg(t *testing.T) {
	entries := []projectEntry{
		{Package: "a", Slug: "gh/acme/a"},
		{Package: "b", Slug: "gh/beta/b"},
		{Package: "c", Slug: "gh/acme/c"},
	}
	groups := groupByOrg(entries)
	if len(groups) != 2 {
		t.Fatalf("got %d groups, want 2: %+v", len(groups), groups)
	}
	if groups[0].org != "gh/acme" || len(groups[0].entries) != 2 ||
		groups[0].entries[0].Package != "a" || groups[0].entries[1].Package != "c" {
		t.Errorf("group 0 = %+v, want gh/acme with [a c]", groups[0])
	}
	if groups[1].org != "gh/beta" || len(groups[1].entries) != 1 || groups[1].entries[0].Package != "b" {
		t.Errorf("group 1 = %+v, want gh/beta with [b]", groups[1])
	}
}
