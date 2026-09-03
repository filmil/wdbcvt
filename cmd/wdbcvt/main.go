// SPDX-License-Identifier: Apache-2.0

// Command wdbcvt inspects Vivado xsim waveform databases.
//
// The .wdb container format is undocumented, so the tool starts as a
// probe: it reports the measurements that a decoder has to be built on.
// Point it at the file produced by //hdl/counter:sim.
//
//	bazel build //hdl/counter:sim
//	bazel run //cmd/wdbcvt -- -in $PWD/bazel-bin/hdl/counter/sim.wdb
package main

import (
	"flag"
	"fmt"
	"os"

	"git.hdlfactory.com/HDL/wdbcvt/pkg/wdb"
)

func main() {
	in := flag.String("in", "", "the .wdb file to inspect (required)")
	maxStrings := flag.Int("max_strings", 40, "how many printable runs to print")
	maxBlocks := flag.Int("max_blocks", 16, "how many entropy blocks to print")
	flag.Parse()

	if err := run(*in, *maxStrings, *maxBlocks); err != nil {
		fmt.Fprintf(os.Stderr, "wdbcvt: %v\n", err)
		os.Exit(1)
	}
}

func run(in string, maxStrings, maxBlocks int) error {
	if in == "" {
		return fmt.Errorf("-in is required")
	}
	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return err
	}

	rep, err := wdb.Probe(f, st.Size())
	if err != nil {
		return err
	}

	fmt.Printf("file:          %s\n", in)
	fmt.Printf("size:          %d bytes\n", rep.Size)
	fmt.Printf("mean entropy:  %.3f bits/byte\n", rep.MeanEntropy())
	fmt.Printf("\nheader:\n%s\n", rep.HeaderDump())

	fmt.Printf("entropy by %d-byte block (first %d):\n", len(rep.Header), maxBlocks)
	for i, b := range rep.Blocks {
		if i >= maxBlocks {
			fmt.Printf("  ... %d more blocks\n", len(rep.Blocks)-maxBlocks)
			break
		}
		fmt.Printf("  %#010x  %5d B  %.3f\n", b.Offset, b.Length, b.Entropy)
	}

	fmt.Printf("\nlongest printable runs (%d of %d):\n",
		min(maxStrings, len(rep.Strings)), len(rep.Strings))
	for _, s := range rep.TopStrings(maxStrings) {
		fmt.Printf("  %#010x  %q\n", s.Offset, s.Text)
	}
	return nil
}
