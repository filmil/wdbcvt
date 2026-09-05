// SPDX-License-Identifier: Apache-2.0

// Searches a waveform database for a byte pattern, in the file as it
// lies and in the inflated value pages, and prints where it is found.
// A value that is written nowhere is what several of the open questions
// turn on, and this is how that is checked:
//
//	bazel run //tools/pagegrep -- -pat ZQXJ "$PWD/bazel-bin/hdl/corpus/t68_str_lit4____/sim.wdb"
package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"git.hdlfactory.com/HDL/wdbcvt/pkg/wdb"
)

func main() {
	pat := flag.String("pat", "", "the text to search for (required unless -hex is given)")
	hx := flag.String("hex", "", "the bytes to search for, in hex")
	ctx := flag.Int("ctx", 16, "how many bytes of context to print around a hit in the file")
	flag.Parse()
	var want []byte
	switch {
	case *hx != "":
		b, err := hex.DecodeString(*hx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "pagegrep: -hex:", err)
			os.Exit(2)
		}
		want = b
	case *pat != "":
		want = []byte(*pat)
	default:
		fmt.Fprintln(os.Stderr, "pagegrep: give -pat or -hex")
		os.Exit(2)
	}
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "pagegrep: give at least one .wdb file")
		os.Exit(2)
	}
	for _, path := range flag.Args() {
		if err := grep(path, want, *ctx); err != nil {
			fmt.Fprintln(os.Stderr, "pagegrep:", err)
			os.Exit(1)
		}
	}
}

func grep(path string, want []byte, ctx int) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fmt.Printf("%s: %d bytes\n", path, len(raw))
	hits := 0
	for i := 0; ; {
		j := bytes.Index(raw[i:], want)
		if j < 0 {
			break
		}
		at := i + j
		lo, hi := max(0, at-ctx), min(len(raw), at+len(want)+ctx)
		fmt.Printf("  file %#x: %s\n", at, hex.EncodeToString(raw[lo:hi]))
		hits++
		i = at + 1
	}
	f, err := wdb.Read(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	for a, pages := range f.Pages {
		for p, pg := range pages {
			for r, rec := range pg.Records {
				if !bytes.Contains(rec.Data, want) {
					continue
				}
				fmt.Printf("  arena %d page %d record %d: time %d key %#x len %d: %s\n",
					a, p, r, rec.Time, rec.Key, len(rec.Data), hex.EncodeToString(rec.Data))
				hits++
			}
		}
	}
	// A page holds more than its records: the header and whatever the
	// writer left between them. The records above are what the reader
	// sees, so a hit in a page that no record covers is worth knowing.
	for a, pages := range f.Pages {
		for p, pg := range pages {
			covered := 0
			for _, rec := range pg.Records {
				if bytes.Contains(rec.Data, want) {
					covered++
				}
			}
			if covered == 0 && pageHolds(pg, want) {
				fmt.Printf("  arena %d page %d: in the page but in no record\n", a, p)
				hits++
			}
		}
	}
	if hits == 0 {
		fmt.Printf("  not found in %s\n", strings.TrimSuffix(path, ".wdb"))
	}
	return nil
}

// pageHolds reports whether the pattern is anywhere in a page's records
// taken together, which is as much of a page as the reader keeps.
func pageHolds(pg wdb.Page, want []byte) bool {
	var b []byte
	for _, rec := range pg.Records {
		b = append(b, rec.Data...)
	}
	return bytes.Contains(b, want)
}
