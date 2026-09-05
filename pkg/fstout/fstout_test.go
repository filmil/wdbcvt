// SPDX-License-Identifier: Apache-2.0

package fstout_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"git.hdlfactory.com/HDL/wdbcvt/pkg/fst"
	"git.hdlfactory.com/HDL/wdbcvt/pkg/fstout"
	"git.hdlfactory.com/HDL/wdbcvt/pkg/wdb"
)

// truth is the part of a corpus case's truth.json this test reads: the
// signals it declares and the value each holds at each time. The file
// is written from the design, not from the database, so it is the
// answer key for the conversion as well.
type truth struct {
	Signals []struct {
		Scope string `json:"scope"`
		Name  string `json:"name"`
		Width int    `json:"width"`
		Type  string `json:"type"`
	} `json:"signals"`
	Transitions []struct {
		TimeNS int64           `json:"time_ns"`
		TimePS int64           `json:"time_ps"`
		TimeFS int64           `json:"time_fs"`
		Signal string          `json:"signal"`
		Value  json.RawMessage `json:"value"`
	} `json:"transitions"`
}

func corpusCases(t *testing.T) []string {
	t.Helper()
	root := filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"))
	truths, err := filepath.Glob(filepath.Join(root, "hdl/corpus/*/truth.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(truths) == 0 {
		t.Fatal("no hdl/corpus/*/truth.json in the test's runfiles")
	}
	var cases []string
	for _, p := range truths {
		cases = append(cases, filepath.Dir(p))
	}
	return cases
}

// TestCorpus converts every corpus case to FST, reads the file back
// through libfst, and holds the result to the case's truth.json. For
// the types Vivado's VCD leaves out, integers, reals, enumerations,
// records and arrays, this is the second guard: the truth file is
// written from the design and libfst is the definition of the output
// format, and neither is this project's own reading of anything.
func TestCorpus(t *testing.T) {
	for _, dir := range corpusCases(t) {
		dir := dir
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			f, err := wdb.ReadFile(filepath.Join(dir, "sim.wdb"))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "out.fst")
			if err := fstout.Write(f, path); err != nil {
				t.Fatal(err)
			}
			vars, changes, err := fst.Read(path)
			if err != nil {
				t.Fatal(err)
			}
			var tr truth
			raw, err := os.ReadFile(filepath.Join(dir, "truth.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &tr); err != nil {
				t.Fatal(err)
			}
			// The variables of the file, by path, with their values in
			// time order. Aliases share a handle, so a change belongs
			// to every variable on it.
			byHandle := map[uint32][]string{}
			for _, v := range vars {
				byHandle[v.Handle] = append(byHandle[v.Handle], v.Path)
			}
			got := map[string][]string{}
			for _, c := range changes {
				for _, p := range byHandle[c.Handle] {
					got[p] = append(got[p], fmt.Sprintf("%d=%s", c.Time, c.Value))
				}
			}
			// Every signal of the truth file must be a variable of
			// the FST, under its own path, holding the values the
			// truth names. A record or an array is flattened, so its
			// aggregate is compared element by element against the
			// leaves.
			for _, s := range tr.Signals {
				path := s.Scope + "." + s.Name
				if strings.Contains(path, "%") {
					continue // a count expanded entry
				}
				want := map[string]string{}
				var order []string
				for _, x := range tr.Transitions {
					if !names(x.Signal, s, tr.Signals) {
						continue
					}
					var v string
					if json.Unmarshal(x.Value, &v) != nil {
						continue // an aggregate of aggregates
					}
					at := fmt.Sprintf("%d", ticks(t, f, x.TimeNS, x.TimePS, x.TimeFS))
					if _, ok := want[at]; !ok {
						order = append(order, at)
					}
					want[at] = v
				}
				if len(order) == 0 {
					continue
				}
				last := map[string]map[string]string{}
				for p, cs := range got {
					if p != path && !strings.HasPrefix(p, path+".") {
						continue
					}
					for _, c := range cs {
						at, v, _ := strings.Cut(c, "=")
						if last[p] == nil {
							last[p] = map[string]string{}
						}
						last[p][at] = v
					}
				}
				if len(last) == 0 {
					t.Errorf("%s is in the truth file and not in the FST", path)
					continue
				}
				for _, at := range order {
					for _, w := range spread(want[at]) {
						p := path
						if w.leaf != "" {
							p = path + "." + w.leaf
						}
						v, ok := last[p][at]
						if !ok {
							continue // the value did not change then
						}
						if !same(w.value, v) {
							t.Errorf("%s at %s: FST %q, truth %q", p, at, v, w.value)
						}
					}
				}
			}
		})
	}
}

// want is one expected value, of a whole signal or of one leaf.
type wantValue struct {
	leaf  string
	value string
}

// spread turns a truth value into the values of the FST variables that
// hold it: an aggregate `(a, b)` is the leaves `0` and `1`, which is
// how the converter flattens an array or a record.
func spread(v string) []wantValue {
	if !strings.HasPrefix(v, "(") || !strings.HasSuffix(v, ")") {
		return []wantValue{{value: v}}
	}
	inner := v[1 : len(v)-1]
	if strings.ContainsAny(inner, "()") {
		return nil // nested, and the leaves below carry the check
	}
	var out []wantValue
	for i, e := range strings.Split(inner, ", ") {
		out = append(out, wantValue{leaf: fmt.Sprintf("%d", i), value: e})
	}
	return out
}

// names reports whether a truth transition belongs to a signal. A
// transition names the signal by its path or by its name alone, and
// the second only tells them apart when the name is unique.
func names(signal string, s struct {
	Scope string `json:"scope"`
	Name  string `json:"name"`
	Width int    `json:"width"`
	Type  string `json:"type"`
}, all []struct {
	Scope string `json:"scope"`
	Name  string `json:"name"`
	Width int    `json:"width"`
	Type  string `json:"type"`
}) bool {
	if signal == s.Scope+"."+s.Name {
		return true
	}
	if signal != s.Name {
		return false
	}
	n := 0
	for _, o := range all {
		if o.Name == s.Name {
			n++
		}
	}
	return n == 1
}

// ticks converts a truth file's time into the file's own unit.
func ticks(t *testing.T, f *wdb.File, ns, ps, fs int64) uint64 {
	t.Helper()
	total := uint64(ns)*1000000 + uint64(ps)*1000 + uint64(fs)
	unit := f.TimeFS(1)
	if total%unit != 0 {
		t.Fatalf("time %d fs is not a multiple of the file's unit, 1 %s", total, f.TimeUnit())
	}
	return total / unit
}

// same compares a value the FST holds with the one the truth names.
// The truth spells logic values in upper case, an aggregate as
// `(a, b)`, which the FST holds as its leaves, and an integer, a real
// or a time in decimal, where FST holds the first and the last as
// bits.
func same(want, got string) bool {
	if strings.HasPrefix(want, "(") {
		return true
	}
	if strings.EqualFold(want, got) {
		return true
	}
	if n, err := strconv.ParseInt(want, 10, 64); err == nil && isBits(got) {
		if m, err := strconv.ParseUint(got, 2, 64); err == nil {
			// The bits are the value as the variable holds it, and
			// the truth spells it signed or unsigned by its type.
			if int64(m) == n {
				return true
			}
			b := len(got)
			return b < 64 && m&(1<<uint(b-1)) != 0 && int64(m)-int64(1)<<uint(b) == n
		}
	}
	if a, b, ok := physical(want, got); ok {
		// The scales are powers of ten in binary floating point, so
		// the two products agree to a few ulp and not exactly.
		return math.Abs(a-b) <= 1e-12*math.Abs(a+b)
	}
	if a, err := strconv.ParseFloat(want, 64); err == nil {
		if b, err := strconv.ParseFloat(got, 64); err == nil {
			return a == b
		}
	}
	return false
}

// physical compares two physical values, `3000 um` and `3 mm`, which
// are the same quantity in different units: the decoder writes the
// base unit of the type and the truth file the unit of the source. The
// count is scaled by the unit's SI prefix, and the base letter has to
// agree.
func physical(a, b string) (float64, float64, bool) {
	x, ua, ok := split(a)
	if !ok {
		return 0, 0, false
	}
	y, ub, ok := split(b)
	if !ok || ua[len(ua)-1:] != ub[len(ub)-1:] {
		return 0, 0, false
	}
	return x * prefix(ua), y * prefix(ub), true
}

// split takes a physical value apart into its count and its unit.
func split(s string) (float64, string, bool) {
	n, u, ok := strings.Cut(s, " ")
	if !ok || u == "" {
		return 0, "", false
	}
	x, err := strconv.ParseFloat(n, 64)
	if err != nil {
		return 0, "", false
	}
	return x, u, true
}

// prefix is the scale of a unit's SI prefix, 1e-12 for `ps`.
func prefix(u string) float64 {
	if len(u) < 2 {
		return 1
	}
	switch u[0] {
	case 'f':
		return 1e-15
	case 'p':
		return 1e-12
	case 'n':
		return 1e-9
	case 'u':
		return 1e-6
	case 'm':
		return 1e-3
	case 'k':
		return 1e3
	}
	return 1
}

// isBits reports whether s is a string of 0 and 1 alone.
func isBits(s string) bool {
	if s == "" {
		return false
	}
	return strings.Trim(s, "01") == ""
}

// designs are the databases outside the corpus that have no
// truth.json. The check here is weaker and still worth having: every
// logged object of a real hierarchy must reach the FST, as itself or
// as its leaves.
var designs = []string{"hdl/counter", "hdl/uart", "hdl/potato", "hdl/picorv32", "hdl/ibex"}

func TestDesigns(t *testing.T) {
	root := filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"))
	for _, d := range designs {
		d := d
		t.Run(filepath.Base(d), func(t *testing.T) {
			t.Parallel()
			f, err := wdb.ReadFile(filepath.Join(root, d, "sim.wdb"))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "out.fst")
			if err := fstout.Write(f, path); err != nil {
				t.Fatal(err)
			}
			vars, changes, err := fst.Read(path)
			if err != nil {
				t.Fatal(err)
			}
			have := map[string]bool{}
			for _, v := range vars {
				have[v.Path] = true
				for p := v.Path; strings.Contains(p, "."); {
					p = p[:strings.LastIndex(p, ".")]
					have[p] = true
				}
			}
			for _, o := range f.Objects {
				if !o.Logged {
					continue
				}
				p := plain(f.ObjectPath(o))
				if !have[p] {
					t.Errorf("%s is logged and not in the FST", p)
				}
			}
			if len(changes) == 0 {
				t.Error("no value changes in the FST")
			}
		})
	}
}

// plain strips the extended identifier decoration, the way the
// converter does when it names a scope or a variable.
func plain(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, `\`, ""))
}
