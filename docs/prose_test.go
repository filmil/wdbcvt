// SPDX-License-Identifier: Apache-2.0

// Holds every document of this repository to the writing rules of
// AGENTS.md. A rule that can be checked mechanically is checked here,
// so that breaking one fails `bazel test //...`.
package docs_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"git.hdlfactory.com/HDL/wdbcvt/tools/prose"
)

// documents are the markdown files the rules apply to: the pages under
// docs/, the README and the instruction file.
func documents(t *testing.T) []string {
	t.Helper()
	root := filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"))
	var out []string
	for _, pat := range []string{"docs/*.md", "docs/format/*.md", "*.md"} {
		m, err := filepath.Glob(filepath.Join(root, pat))
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, m...)
	}
	// AGENTS.md is the one file; CLAUDE.md and GEMINI.md are symlinks
	// to it, and checking it three times only repeats the failures.
	seen := map[string]bool{}
	var real []string
	for _, p := range out {
		r, err := filepath.EvalSymlinks(p)
		if err != nil {
			r = p
		}
		if seen[r] {
			continue
		}
		seen[r] = true
		real = append(real, p)
	}
	out = real
	sort.Strings(out)
	if len(out) == 0 {
		t.Fatal("no documents in the test's runfiles")
	}
	return out
}

func TestWritingRules(t *testing.T) {
	root := filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"))
	for _, p := range documents(t) {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		name := strings.TrimPrefix(p, root+"/")
		for _, h := range prose.Check(name, string(b)) {
			t.Errorf("%v", h)
		}
	}
}
