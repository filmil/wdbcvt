// SPDX-License-Identifier: Apache-2.0

// Command bundle collects WDB databases, converted FST waveforms,
// and HDL sources into a release archive.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"git.hdlfactory.com/HDL/wdbcvt/tools/waveforms"
)

func main() {
	ws := os.Getenv("BUILD_WORKSPACE_DIRECTORY")
	defaultSrc := "."
	defaultBin := "bazel-bin"
	if ws != "" {
		defaultSrc = ws
		defaultBin = filepath.Join(ws, "bazel-bin")
	}

	binDir := flag.String("bin", defaultBin, "path to bazel-bin directory")
	srcDir := flag.String("src", defaultSrc, "path to repository root directory")
	outPath := flag.String("out", "dist/release/wdbcvt-waveforms.tar.gz", "output tar.gz archive path")
	prefix := flag.String("prefix", "wdbcvt-waveforms/", "archive internal prefix")
	flag.Parse()

	out := *outPath
	if ws != "" && !filepath.IsAbs(out) {
		out = filepath.Join(ws, out)
	}

	opts := waveforms.Options{
		BinDir:  *binDir,
		SrcDir:  *srcDir,
		OutPath: out,
		Prefix:  *prefix,
	}

	if err := waveforms.Bundle(opts); err != nil {
		fmt.Fprintf(os.Stderr, "bundle: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s\n", out)
}
