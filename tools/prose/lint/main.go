// SPDX-License-Identifier: Apache-2.0

// Checks the documents named on the command line against the writing
// rules, and exits non zero when one of them breaks a rule:
//
//	bazel run //tools/prose/lint -- docs/format.md
//
// //docs:prose_test runs it over every document in the repository, so
// a break fails `bazel test //...` rather than waiting for a reader.
package main

import (
	"fmt"
	"os"

	"git.hdlfactory.com/HDL/wdbcvt/tools/prose"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "prose lint: name the documents to check")
		os.Exit(2)
	}
	n := 0
	for _, f := range os.Args[1:] {
		b, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, "prose lint:", err)
			os.Exit(1)
		}
		for _, h := range prose.Check(f, string(b)) {
			fmt.Println(h)
			n++
		}
	}
	if n > 0 {
		fmt.Fprintf(os.Stderr, "\n%d line(s) break a writing rule; see AGENTS.md\n", n)
		os.Exit(1)
	}
}
