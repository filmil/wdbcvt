// SPDX-License-Identifier: Apache-2.0

// Command wdbdiff compares two waveform databases, ignoring the bytes
// that a database embeds for reasons unrelated to the design.
//
// The corpus method needs two steps, and the first is the one that is
// easy to skip. Simulate one design twice and derive a noise mask from
// the pair; only then compare two different designs through that mask.
// Without the mask a run timestamp gets confidently identified as a
// signal count. See doc/corpus.md.
//
//	# What changes between two runs of the same design.
//	wdbdiff -a run1.wdb -b run2.wdb
//
//	# What one axis of the corpus actually changes.
//	wdbdiff -mask-a run1.wdb -mask-b run2.wdb \
//	        -a t1_bit_one_edge.wdb -b t1_two_bits.wdb
package main

import (
	"flag"
	"fmt"
	"os"

	"git.hdlfactory.com/HDL/wdbcvt/pkg/wdb"
)

func main() {
	a := flag.String("a", "", "the first database (required)")
	b := flag.String("b", "", "the second database (required)")
	maskA := flag.String("mask-a", "", "one design simulated once, to derive the noise mask")
	maskB := flag.String("mask-b", "", "the same design simulated again")
	max := flag.Int("max", 60, "how many differing runs to print")
	flag.Parse()

	if err := run(*a, *b, *maskA, *maskB, *max); err != nil {
		fmt.Fprintf(os.Stderr, "wdbdiff: %v\n", err)
		os.Exit(1)
	}
}

func run(aPath, bPath, maskAPath, maskBPath string, max int) error {
	if aPath == "" || bPath == "" {
		return fmt.Errorf("both -a and -b are required")
	}
	if (maskAPath == "") != (maskBPath == "") {
		return fmt.Errorf("-mask-a and -mask-b go together; a mask needs one design simulated twice")
	}

	av, err := os.ReadFile(aPath)
	if err != nil {
		return err
	}
	bv, err := os.ReadFile(bPath)
	if err != nil {
		return err
	}

	var mask wdb.Mask
	if maskAPath != "" {
		ma, err := os.ReadFile(maskAPath)
		if err != nil {
			return err
		}
		mb, err := os.ReadFile(maskBPath)
		if err != nil {
			return err
		}
		mask = wdb.NoiseMask(ma, mb)
		spans := mask.Spans()
		fmt.Printf("noise mask: %d bytes over %d spans, from %s and %s\n",
			mask.Len(), len(spans), maskAPath, maskBPath)
		for i, s := range spans {
			if i >= max {
				fmt.Printf("  ... %d more spans\n", len(spans)-max)
				break
			}
			fmt.Printf("  %#010x  %d bytes\n", s.Offset, s.Length)
		}
		if mask.Len() == 0 {
			fmt.Println("  the two runs are byte for byte identical, so nothing is masked")
		}
		fmt.Println()
	} else {
		fmt.Println("no mask given: every difference below may be noise rather than design")
		fmt.Println()
	}

	deltas, lenDelta := wdb.Compare(av, bv, mask)

	fmt.Printf("a: %s (%d bytes)\n", aPath, len(av))
	fmt.Printf("b: %s (%d bytes)\n", bPath, len(bv))
	fmt.Printf("size difference: %+d bytes\n", lenDelta)
	fmt.Printf("differing runs outside the mask: %d\n\n", len(deltas))

	for i, d := range deltas {
		if i >= max {
			fmt.Printf("  ... %d more runs\n", len(deltas)-max)
			break
		}
		fmt.Printf("  %s\n", d)
	}
	if len(deltas) == 0 {
		fmt.Println("  none: outside the mask these two databases are identical")
	}
	return nil
}
