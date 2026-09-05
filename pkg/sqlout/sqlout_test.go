// SPDX-License-Identifier: Apache-2.0

package sqlout_test

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"git.hdlfactory.com/HDL/wdbcvt/pkg/sqlout"
	"git.hdlfactory.com/HDL/wdbcvt/pkg/wdb"
	"github.com/filmil/go-vcd-parser/cvt"
	"github.com/filmil/go-vcd-parser/db"
	"github.com/filmil/go-vcd-parser/vcd"
	_ "github.com/mattn/go-sqlite3"
)

// designs are simulations that write both a database and a VCD, and
// whose time unit is the picosecond the VCD's own timescale names, so
// the timestamps of the two databases are the same numbers.
var designs = []string{"hdl/counter", "hdl/uart"}

func root(t *testing.T) string {
	t.Helper()
	return filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"))
}

// reference builds the database go-vcd-parser writes from a VCD, which
// is the answer key: the rows this package writes for the signals the
// VCD also carries must say the same thing.
func reference(t *testing.T, vcdPath, out string) *sql.DB {
	t.Helper()
	f, err := os.Open(vcdPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ast, err := vcd.NewParser[vcd.File]().Parse(vcdPath, bufio.NewReaderSize(f, 1<<20))
	if err != nil {
		t.Fatalf("%s: %v", vcdPath, err)
	}
	ctx := context.Background()
	d, err := db.OpenDB(ctx, out)
	if err != nil {
		t.Fatal(err)
	}
	if err := cvt.Convert(ctx, ast, d); err != nil {
		t.Fatal(err)
	}
	return d
}

// signals reads a Signals table into a map of name to size.
func signals(t *testing.T, d *sql.DB) map[string]int {
	t.Helper()
	rows, err := d.Query(`SELECT Name, Size FROM Signals;`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var name string
		var size int
		if err := rows.Scan(&name, &size); err != nil {
			t.Fatal(err)
		}
		out[name] = size
	}
	return out
}

// changes reads a database into the values of each signal over time,
// by name, sorted by time.
func changes(t *testing.T, d *sql.DB) map[string][]change {
	t.Helper()
	rows, err := d.Query(`
        SELECT Signals.Name, Svalues.Timestamp, Svalues.Value
        FROM Svalues JOIN Signals ON Signals.Code = Svalues.Code
        ORDER BY Svalues.Id;`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[string][]change{}
	for rows.Next() {
		var name, value string
		var ts int64
		if err := rows.Scan(&name, &ts, &value); err != nil {
			t.Fatal(err)
		}
		out[name] = append(out[name], change{ts, value})
	}
	for _, l := range out {
		sort.SliceStable(l, func(i, j int) bool { return l[i].t < l[j].t })
	}
	return out
}

type change struct {
	t int64
	v string
}

// at is the value a signal holds at time t: the last change at or
// before it.
func at(l []change, t int64) (string, bool) {
	v, ok := "", false
	for _, c := range l {
		if c.t > t {
			break
		}
		v, ok = c.v, true
	}
	return v, ok
}

// extend spells a VCD value the way this package's rows do: a VCD
// drops the leading digits of a vector and extends what is left with
// its own leftmost digit, `0` for a `1`, so the two are compared at
// the declared width.
func extend(v string, size int) string {
	v = strings.ToLower(v)
	if size <= 1 || len(v) >= size {
		return v
	}
	pad := "0"
	switch v[0] {
	case 'x', 'z', 'u', 'w', 'l', 'h', '-':
		pad = string(v[0])
	}
	return strings.Repeat(pad, size-len(v)) + v
}

func TestAgainstVCD(t *testing.T) {
	for _, design := range designs {
		t.Run(design, func(t *testing.T) {
			dir := filepath.Join(root(t), design)
			out := filepath.Join(t.TempDir(), "out.db")
			f, err := wdb.ReadFile(filepath.Join(dir, "sim.wdb"))
			if err != nil {
				t.Fatal(err)
			}
			if err := sqlout.Write(f, out); err != nil {
				t.Fatal(err)
			}
			mine, err := sql.Open("sqlite3", out)
			if err != nil {
				t.Fatal(err)
			}
			defer mine.Close()
			ref := reference(t, filepath.Join(dir, "sim.vcd"), filepath.Join(t.TempDir(), "ref.db"))
			defer ref.Close()

			mySigs, refSigs := signals(t, mine), signals(t, ref)
			myVals, refVals := changes(t, mine), changes(t, ref)
			if len(refSigs) == 0 {
				t.Fatalf("the VCD of %s declares no signals", design)
			}
			for name, size := range refSigs {
				if _, ok := mySigs[name]; !ok {
					t.Errorf("%s: no row for %q, which the VCD declares", design, name)
					continue
				}
				if mySigs[name] != size {
					t.Errorf("%s: %q is %d bits, the VCD says %d", design, name, mySigs[name], size)
				}
				for _, c := range refVals[name] {
					got, ok := at(myVals[name], c.t)
					if !ok {
						t.Errorf("%s: %q has no value at %d, the VCD has %q", design, name, c.t, c.v)
						break
					}
					if want := extend(c.v, size); got != want {
						t.Errorf("%s: %q at %d is %q, the VCD says %q", design, name, c.t, got, want)
						break
					}
				}
			}
		})
	}
}

// TestExtraSignals holds the point of the output: the database carries
// the objects the VCD has nowhere to put.
func TestExtraSignals(t *testing.T) {
	dir := filepath.Join(root(t), "hdl/counter")
	out := filepath.Join(t.TempDir(), "out.db")
	f, err := wdb.ReadFile(filepath.Join(dir, "sim.wdb"))
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlout.Write(f, out); err != nil {
		t.Fatal(err)
	}
	d, err := sql.Open("sqlite3", out)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	sigs := signals(t, d)
	// A boolean, a time and an integer constant, none of which the
	// VCD of the same run declares; see docs/format/vcd.md.
	for _, name := range []string{"//tb/running", "//tb/period", "//tb/full_scale"} {
		if _, ok := sigs[name]; !ok {
			t.Errorf("no row for %q; the rows are %v", name, keys(sigs))
		}
	}
	vals := changes(t, d)
	if v, ok := at(vals["//tb/running"], 0); !ok || v != "TRUE" {
		t.Errorf("//tb/running at 0 is %q, want TRUE", v)
	}
}

func keys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestMeta checks the table this writer adds to the schema.
func TestMeta(t *testing.T) {
	dir := filepath.Join(root(t), "hdl/counter")
	out := filepath.Join(t.TempDir(), "out.db")
	f, err := wdb.ReadFile(filepath.Join(dir, "sim.wdb"))
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlout.Write(f, out); err != nil {
		t.Fatal(err)
	}
	d, err := sql.Open("sqlite3", out)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	got := map[string]string{}
	rows, err := d.Query(`SELECT Key, Value FROM Meta;`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			t.Fatal(err)
		}
		got[k] = v
	}
	if got["time_exponent"] != fmt.Sprint(int(f.Debug.Precision)) {
		t.Errorf("time_exponent is %q, want %d", got["time_exponent"], int(f.Debug.Precision))
	}
	if got["end_time"] != fmt.Sprint(f.Header.EndTime) {
		t.Errorf("end_time is %q, want %d", got["end_time"], f.Header.EndTime)
	}
	if got["generator"] != "wdbcvt" {
		t.Errorf("generator is %q, want wdbcvt", got["generator"])
	}
}
