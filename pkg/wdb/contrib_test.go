// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The contrib test reads the databases under //testdata/contrib, which
// people outside this repository sent in. They have no truth.json, so
// the test checks only that the reader accepts each file and decodes
// every change of every object, and that each file holds the structure
// it was sent to show. It asserts nothing about the values it reads. See
// testdata/contrib/README.md.

// contribDir is where the contributed databases sit in the test's
// runfiles.
func contribDir() string {
	return filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"), "testdata/contrib")
}

// readContrib reads one contributed database, named by its path under
// testdata/contrib.
func readContrib(t *testing.T, name string) *File {
	t.Helper()
	f, err := ReadFile(filepath.Join(contribDir(), name))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// decodeAll decodes every change of every object f holds, and fails
// the test at the first error.
func decodeAll(t *testing.T, f *File) {
	t.Helper()
	for _, o := range f.Objects {
		dc := f.Decls[o.Decl]
		if o.Generic || dc.Kind.subprogram() {
			continue
		}
		ch, err := f.Changes(o)
		if err != nil {
			t.Fatalf("%s: %v", f.ObjectPath(o), err)
		}
		for _, c := range ch {
			if _, err := f.Decode(dc, c.Data); err != nil {
				t.Fatalf("%s at %d %s: %v", f.ObjectPath(o), c.Time, f.TimeUnit(), err)
			}
		}
	}
}

func TestContribDecodes(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(contribDir(), "*", "*.wdb"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no testdata/contrib/*/*.wdb in the test's runfiles")
	}
	sort.Strings(paths)
	for _, p := range paths {
		p := p
		name, err := filepath.Rel(contribDir(), p)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			decodeAll(t, readContrib(t, name))
		})
	}
}

// This Vivado 2026.1 file from filmil/wdbcvt#1 holds SystemVerilog
// structs, packed arp_tb headers among them, whose type table entries
// have origin 1 and end in a number rather than -99. Before the fix
// the reader refused the file at the first such entry, type 51, with
// "record trailer: got 18, want -99".
func TestContribPR1StructTrailer(t *testing.T) {
	f := readContrib(t, "gh_pr1/repro_0001_record_trailer_arp_tb.wdb")
	n := 0
	for _, ty := range f.Types {
		if ty.Kind == KindRecord && ty.Origin == OriginVerilog {
			n++
		}
	}
	if n == 0 {
		t.Error("no record type of Verilog origin in the type table")
	}
}

// In this Vivado 2026.1 file from filmil/wdbcvt#1,
// gen_level[3].stage_sum of ip_checksum, a logic [18:0] [1:0], has a
// first write that covers half of the object. Before the fix the
// reader refused the file there, with "object handle 0x36c8 with 16
// bytes has a first write of 8 bytes at 0x36c8, which does not cover
// it".
func TestContribPR1PartialFirstWrite(t *testing.T) {
	f := readContrib(t, "gh_pr1/repro_0003_partial_first_write_ip_checksum.wdb")
	var found []Object
	for _, o := range f.Objects {
		if strings.Contains(f.ObjectPath(o), "gen_level[3].stage_sum") {
			found = append(found, o)
		}
	}
	if len(found) == 0 {
		t.Fatal("no object gen_level[3].stage_sum")
	}
	for _, o := range found {
		ch, err := f.Changes(o)
		if err != nil {
			t.Fatalf("%s: %v", f.ObjectPath(o), err)
		}
		if len(ch) == 0 {
			t.Errorf("%s: no changes", f.ObjectPath(o))
		}
	}
}
