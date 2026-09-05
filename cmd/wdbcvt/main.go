// SPDX-License-Identifier: Apache-2.0

// Command wdbcvt inspects Vivado xsim waveform databases.
//
// The .wdb container format is undocumented. What the tool knows about
// it was found by an automated assistant running experiments on the
// corpus under //hdl/corpus, on files written by Vivado 2025.2, and is
// recorded in //docs/format.md. Two modes:
//
//	bazel run //cmd/wdbcvt -- -in $PWD/bazel-bin/hdl/counter/sim.wdb
//	bazel run //cmd/wdbcvt -- -dump -in $PWD/bazel-bin/hdl/counter/sim.wdb
//	bazel run //cmd/wdbcvt -- -in $PWD/bazel-bin/hdl/counter/sim.wdb -fst out.fst
//	bazel run //cmd/wdbcvt -- -in $PWD/bazel-bin/hdl/counter/sim.wdb -sqlite out.db
//
// Without -dump the tool probes: it reports the measurements a decoder
// has to be built on. With -dump it decodes every structure it knows
// and prints them, ending with each object's values over time. With
// -fst it converts the database into an FST waveform file, which
// GTKWave, Surfer and nvc read; see //docs/fst-output.md. With -sqlite
// it writes the signals and their changes into an SQLite file in the
// schema go-vcd-parser writes from a VCD; see //docs/sqlite-output.md.
package main

import (
	"flag"
	"fmt"
	"os"

	"git.hdlfactory.com/HDL/wdbcvt/pkg/fstout"
	"git.hdlfactory.com/HDL/wdbcvt/pkg/sqlout"
	"git.hdlfactory.com/HDL/wdbcvt/pkg/wdb"
)

// runFST converts the database at in into an FST file at out.
func runFST(in, out string) error {
	return convert(in, out, fstout.Write)
}

// runSQLite converts the database at in into an SQLite file at out.
func runSQLite(in, out string) error {
	return convert(in, out, sqlout.Write)
}

// convert reads the database and hands it to one of the writers.
func convert(in, out string, write func(*wdb.File, string) error) error {
	if in == "" {
		return fmt.Errorf("-in is required")
	}
	f, err := wdb.ReadFile(in)
	if err != nil {
		return err
	}
	if err := write(f, out); err != nil {
		return fmt.Errorf("%s: %w", out, err)
	}
	return nil
}

func main() {
	in := flag.String("in", "", "the .wdb file to inspect (required)")
	dump := flag.Bool("dump", false, "decode the file and print every known structure")
	out := flag.String("fst", "", "convert the file into an FST waveform file at this path")
	sqliteOut := flag.String("sqlite", "", "convert the file into an SQLite database at this path")
	maxStrings := flag.Int("max_strings", 40, "how many printable runs to print")
	maxBlocks := flag.Int("max_blocks", 16, "how many entropy blocks to print")
	flag.Usage = usage
	flag.Parse()

	var err error
	if *out != "" && *sqliteOut != "" {
		err = fmt.Errorf("give one of -fst and -sqlite, not both")
	} else if *out != "" {
		err = runFST(*in, *out)
	} else if *sqliteOut != "" {
		err = runSQLite(*in, *sqliteOut)
	} else if *dump {
		err = runDump(*in)
	} else {
		err = run(*in, *maxStrings, *maxBlocks)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "wdbcvt: %v\n", err)
		os.Exit(1)
	}
}

// provenance is printed by --help, so that a user who finds the binary
// without the repository still learns how its format knowledge was
// obtained. Keep it in step with docs/provenance.md.
const provenance = `wdbcvt reads Vivado xsim waveform databases (.wdb).

The .wdb format is undocumented. What this tool knows about it was
worked out by an AI agent running experiments on files that Vivado
2025.2 wrote for small designs, and reading the bytes that came out.
No AMD documentation, source or binary was used. The decoding is
checked against a truth file written per design from its VHDL or
Verilog source, and every claim is scoped to Vivado 2025.2.
It is not a specification. Where a wrong answer would be silent and
expensive, open the database in Vivado instead.
`

func usage() {
	fmt.Fprint(flag.CommandLine.Output(), provenance)
	fmt.Fprintln(flag.CommandLine.Output(), "\nFlags:")
	flag.PrintDefaults()
}

func runDump(in string) error {
	if in == "" {
		return fmt.Errorf("-in is required")
	}
	f, err := wdb.ReadFile(in)
	if err != nil {
		return err
	}
	return f.Dump(os.Stdout)
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
