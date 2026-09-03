// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The corpus test decodes every database under //hdl/corpus and compares
// what it reads with the case's truth.json, which was written from the
// VHDL or Verilog source before the database was ever opened. Nothing here asserts
// against bytes the decoder produced. See docs/provenance.md.

type truthSignal struct {
	Scope        string        `json:"scope"`
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	Width        int           `json:"width"`
	Fields       []truthSignal `json:"fields"`
	Elements     int           `json:"elements"`
	ElementType  string        `json:"element_type"`
	ElementWidth int           `json:"element_width"`
	// Port is the mode of a signal declared in a port list: in, out,
	// inout, buffer or linkage. Empty for a signal declared in an
	// architecture.
	Port string `json:"port"`
	// Logged is false for a signal the simulation declares but never
	// logs, such as a package signal outside the logged hierarchy:
	// t9_pkg_sig. Absent means logged.
	Logged *bool `json:"logged"`
}

// truthRun is a long train of transitions written as a rule rather than
// listed: count changes from start_ns, step_ns apart, cycling through
// values. t9_tr70000 has one.
type truthRun struct {
	Signal  string   `json:"signal"`
	StartNS int64    `json:"start_ns"`
	StepNS  int64    `json:"step_ns"`
	Count   int      `json:"count"`
	Values  []string `json:"values"`
}

type truthTransition struct {
	TimeNS int64           `json:"time_ns"`
	TimePS int64           `json:"time_ps"`
	Signal string          `json:"signal"`
	Value  json.RawMessage `json:"value"`
}

// truthGeneric is one generic of one instance. The early cases name it
// k and give its integer value in K; t9_gen_types names generics of
// other types and spells the value the way the decoder renders it.
type truthGeneric struct {
	Instance string `json:"instance"`
	K        int    `json:"k"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Width    int    `json:"width"`
	// Scope names the scope of a parameter that is not under an
	// instance of tb: the package p of t13_sv_pkg holds p.W.
	Scope string `json:"scope"`
	// Logged is false for a parameter that has no record at all: the
	// package parameter of t13_sv_pkg.
	Logged *bool `json:"logged"`
}

// truthVariable is a process variable: Kind is "variable" for a declared
// one and "loop" for a for loop index. Initial is the declared initial
// value, which the database does not record; see the test.
type truthVariable struct {
	Scope   string `json:"scope"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Kind    string `json:"kind"`
	Initial string `json:"initial"`
	// Value is what the one record of a loop index holds, when the
	// truth states it: a generate index records its iteration value.
	Value string `json:"value"`
	// Logged is false for a constant that gets no record at all: one
	// declared in a package, t9_port_rec and t9_mark_gap.
	Logged *bool `json:"logged"`
}

// plainPath strips the extended identifier bars the database puts
// around a generate iteration, so that `tb.\g(0)\.dut.s` compares with
// the truth's `tb.g(0).dut.s`. Found by t7_gen_for. A Verilog escaped
// identifier, `\g[0].r ` for a reg declared inside a generate block,
// is a backslash and a closing space: t12_v_gen_reg.
func plainPath(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `\`, ""), " ", "")
}

type truth struct {
	Case        string            `json:"case"`
	EndTimeNS   int64             `json:"end_time_ns"`
	EndTimePS   int64             `json:"end_time_ps"`
	Signals     []truthSignal     `json:"signals"`
	Transitions []truthTransition `json:"transitions"`
	Runs        []truthRun        `json:"transition_runs"`
	Generics    []truthGeneric    `json:"generics"`
	Variables   []truthVariable   `json:"variables"`
	// FinalPerTime names signals whose changes are compared by the
	// last value at each time only. The file keeps the write order of
	// one arena and nothing orders two arenas against each other, so
	// a value spanning two arenas and written in pieces at one time
	// comes back in an order the source does not fix: t12_v_mem40_t0.
	FinalPerTime []string `json:"final_per_time"`
}

// corpusCases lists every case directory that has a truth.json and a
// built database, as absolute paths under the test's runfiles.
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
	sort.Strings(cases)
	return cases
}

func loadTruth(t *testing.T, dir string) truth {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "truth.json"))
	if err != nil {
		t.Fatal(err)
	}
	var tr truth
	if err := json.Unmarshal(b, &tr); err != nil {
		t.Fatalf("%s: %v", dir, err)
	}
	return tr
}

// truthValue normalises a truth.json value to the decoder's rendering:
// a string as is, a list of strings as `(a, b, c)`.
func truthValue(raw json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}
	var l []string
	if err := json.Unmarshal(raw, &l); err != nil {
		return "", fmt.Errorf("value %s is neither a string nor a list", raw)
	}
	return "(" + strings.Join(l, ", ") + ")", nil
}

// sameValue compares a truth value with a decoded one. Types whose text
// has more than one spelling are compared numerically: a time in `7 ns`
// against `7000 ps`, a real `0.0` against `0`, and a boolean literal
// `FALSE` against `false`.
func sameValue(f *File, typ int, want, got string) bool {
	if strings.EqualFold(want, got) {
		return true
	}
	switch f.Types[typ].Kind {
	case KindReal:
		a, err1 := strconv.ParseFloat(want, 64)
		b, err2 := strconv.ParseFloat(got, 64)
		return err1 == nil && err2 == nil && a == b
	case KindPhysical:
		a, err1 := parseTime(f.Types[typ], want)
		b, err2 := parseTime(f.Types[typ], got)
		return err1 == nil && err2 == nil && a == b
	case KindArray:
		// An array of reals or times prints as `(a, b, ...)`, and each
		// element is compared as its own type: t10_real40.
		w, g := strings.TrimSuffix(strings.TrimPrefix(want, "("), ")"), strings.TrimSuffix(strings.TrimPrefix(got, "("), ")")
		ws, gs := strings.Split(w, ", "), strings.Split(g, ", ")
		if !strings.HasPrefix(want, "(") || !strings.HasPrefix(got, "(") || len(ws) != len(gs) {
			return false
		}
		for i := range ws {
			if !sameValue(f, f.Types[typ].Elem, ws[i], gs[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// parseTime turns `7 ns` into picoseconds using the type's own unit list.
func parseTime(ty Type, s string) (int64, error) {
	fs := strings.Fields(s)
	if len(fs) != 2 {
		return 0, fmt.Errorf("time %q is not `<n> <unit>`", s)
	}
	n, err := strconv.ParseInt(fs[0], 10, 64)
	if err != nil {
		return 0, err
	}
	for _, u := range ty.Units {
		if u.Name == fs[1] {
			return n * int64(u.Scale), nil
		}
	}
	return 0, fmt.Errorf("time %q has no unit of %s", s, ty.Name)
}

// checkType compares a truth signal's type description with the decoded
// type at index typ, recursing into record fields.
func checkType(t *testing.T, f *File, path string, want truthSignal, typ int) {
	t.Helper()
	ty := f.Types[typ]
	if !strings.EqualFold(ty.Name, want.Type) {
		t.Errorf("%s: type %q, truth says %q", path, ty.Name, want.Type)
	}
	// A SystemVerilog typedef names an alias entry; the struct or enum
	// itself is the entry it points at: t11_sv_struct.
	ty = f.Types[f.resolve(typ)]
	if len(want.Fields) > 0 {
		if ty.Kind != KindRecord {
			t.Errorf("%s: truth lists fields but the type is %s", path, ty.Kind)
			return
		}
		if len(ty.Fields) != len(want.Fields) {
			t.Errorf("%s: %d fields, truth says %d", path, len(ty.Fields), len(want.Fields))
			return
		}
		for i, fd := range ty.Fields {
			if fd.Name != want.Fields[i].Name {
				t.Errorf("%s: field %d is %q, truth says %q", path, i, fd.Name, want.Fields[i].Name)
			}
			checkType(t, f, path+"."+fd.Name, want.Fields[i], fd.Type)
		}
	}
	if want.ElementType != "" {
		if ty.Kind != KindArray {
			t.Errorf("%s: truth lists an element type but the type is %s", path, ty.Kind)
			return
		}
		if el := f.Types[ty.Elem]; !strings.EqualFold(el.Name, want.ElementType) {
			t.Errorf("%s: element type %q, truth says %q", path, el.Name, want.ElementType)
		}
	}
}

// changes lists (time, value) per leaf path of every non-generic object,
// dropping repeats of the same value so that a record written as a whole
// compares with truth transitions written per field.
type change struct {
	timePS uint64
	value  string
}

func decodedChanges(t *testing.T, f *File) (map[string][]change, map[string]int) {
	t.Helper()
	out := map[string][]change{}
	types := map[string]int{}
	for _, o := range f.Objects {
		if o.Generic {
			continue
		}
		dc := f.Decls[o.Decl]
		ch, err := f.Changes(o)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range ch {
			v, err := f.Decode(dc, c.Data)
			if err != nil {
				t.Fatalf("%s at %d ps: %v", f.ObjectPath(o), c.TimePS, err)
			}
			for _, leaf := range f.Leaves(plainPath(f.ObjectPath(o)), v) {
				s := f.String(leaf.Value)
				prev := out[leaf.Path]
				if len(prev) > 0 && prev[len(prev)-1].value == s {
					continue
				}
				out[leaf.Path] = append(prev, change{c.TimePS, s})
				types[leaf.Path] = leaf.Type
			}
		}
	}
	return out, types
}

// finalPerTime keeps the last change at each time.
func finalPerTime(ch []change) []change {
	var out []change
	for _, c := range ch {
		if len(out) > 0 && out[len(out)-1].timePS == c.timePS {
			out[len(out)-1] = c
			continue
		}
		out = append(out, c)
	}
	return out
}

func TestCorpus(t *testing.T) {
	for _, dir := range corpusCases(t) {
		dir := dir
		t.Run(filepath.Base(dir), func(t *testing.T) {
			tr := loadTruth(t, dir)
			f, err := ReadFile(filepath.Join(dir, "sim.wdb"))
			if err != nil {
				t.Fatal(err)
			}

			if got, want := f.Header.EndTimePS, uint64(tr.EndTimeNS)*1000+uint64(tr.EndTimePS); got != want {
				t.Errorf("end time %d ps, truth says %d", got, want)
			}

			// Every declared signal in the truth is an object with the
			// right scope, name and type, and the declared size agrees
			// with the size the type table implies.
			objByPath := map[string]Object{}
			for _, o := range f.Objects {
				objByPath[plainPath(f.ObjectPath(o))] = o
				dc := f.Decls[o.Decl]
				size, err := f.Size(dc)
				// A real parameter declares 16 where a real variable
				// declares 32: t12_v_params. The value is still one
				// 8 byte float.
				realParam := dc.Kind == DeclParam && f.Types[f.resolve(dc.Type)].Kind == KindReal && dc.Size == 16
				if err != nil {
					t.Errorf("%s: %v", f.ObjectPath(o), err)
				} else if size != dc.Size && !realParam {
					t.Errorf("%s: type table implies %d bytes, declaration says %d", f.ObjectPath(o), size, dc.Size)
				}
			}
			// Objects are counted by path: a process variable in an
			// entity instantiated n times is listed n times in each of
			// the n process scopes, with the n handles. t9_mark_two,
			// t9_var_inst3.
			signals, others := map[string]bool{}, map[string]bool{}
			for _, o := range f.Objects {
				if !o.Generic {
					signals[plainPath(f.ObjectPath(o))] = true
				} else {
					others[plainPath(f.ObjectPath(o))] = true
				}
			}
			if len(signals) != len(tr.Signals) {
				t.Errorf("%d signal objects, truth lists %d", len(signals), len(tr.Signals))
			}
			if want := len(tr.Generics) + len(tr.Variables); len(others) != want {
				t.Errorf("%d generic and variable objects, truth lists %d", len(others), want)
			}
			for _, s := range tr.Signals {
				path := s.Scope + "." + s.Name
				o, ok := objByPath[path]
				if !ok {
					t.Errorf("truth signal %s is not an object", path)
					continue
				}
				if o.Generic {
					t.Errorf("%s is a generic, truth says signal", path)
				}
				checkType(t, f, path, s, f.Decls[o.Decl].Type)
				// A port's mode is in the declaration: t8_port_in,
				// t8_port_out, t8_port_inout and t8_port_buffer.
				mode := "none"
				if s.Port != "" {
					mode = s.Port
				}
				if got := f.Decls[o.Decl].Mode.String(); got != mode {
					t.Errorf("%s: port mode %s, truth says %s", path, got, mode)
				}
				if logged := s.Logged == nil || *s.Logged; o.Logged != logged {
					t.Errorf("%s: logged %v, truth says %v", path, o.Logged, logged)
				}
			}

			// Values over time. Truth names a signal either by its full
			// path or by its name below the scope; try both.
			got, types := decodedChanges(t, f)
			want := map[string][]change{}
			for _, x := range tr.Transitions {
				path := x.Signal
				if _, ok := got[path]; !ok {
					for _, s := range tr.Signals {
						if x.Signal == s.Name || strings.HasPrefix(x.Signal, s.Name+".") {
							path = s.Scope + "." + x.Signal
							break
						}
					}
				}
				v, err := truthValue(x.Value)
				if err != nil {
					t.Fatal(err)
				}
				want[path] = append(want[path], change{uint64(x.TimeNS)*1000 + uint64(x.TimePS), v})
			}
			for _, r := range tr.Runs {
				path := r.Signal
				if _, ok := got[path]; !ok {
					path = "tb." + r.Signal
				}
				for i := 0; i < r.Count; i++ {
					ps := uint64(r.StartNS+int64(i)*r.StepNS) * 1000
					want[path] = append(want[path], change{ps, r.Values[i%len(r.Values)]})
				}
				sort.Slice(want[path], func(i, j int) bool { return want[path][i].timePS < want[path][j].timePS })
			}
			for path, w := range want {
				g, ok := got[path]
				if !ok {
					t.Errorf("truth has transitions for %s, the database has none", path)
					continue
				}
				for _, name := range tr.FinalPerTime {
					if path == name || path == "tb."+name {
						g, w = finalPerTime(g), finalPerTime(w)
					}
				}
				if len(g) != len(w) {
					t.Errorf("%s: %d changes, truth lists %d: got %v, want %v", path, len(g), len(w), g, w)
					continue
				}
				for i := range w {
					if g[i].timePS != w[i].timePS || !sameValue(f, types[path], w[i].value, g[i].value) {
						t.Errorf("%s change %d: got %d ps %q, truth says %d ps %q", path, i, g[i].timePS, g[i].value, w[i].timePS, w[i].value)
					}
				}
			}
			for path := range got {
				if _, ok := want[path]; !ok {
					t.Errorf("database has changes for %s, truth lists none", path)
				}
			}

			// Generics: one object per instance, valued at time zero.
			for _, g := range tr.Generics {
				name, value := g.Name, g.Value
				if name == "" {
					name, value = "k", strconv.Itoa(g.K)
				}
				path := "tb." + g.Instance + "." + name
				if g.Scope != "" {
					path = g.Scope + "." + name
				}
				o, ok := objByPath[path]
				if !ok || !o.Generic {
					t.Errorf("truth generic %s is not a generic object", path)
					continue
				}
				dc := f.Decls[o.Decl]
				if logged := g.Logged == nil || *g.Logged; o.Logged != logged {
					t.Errorf("%s: logged %v, truth says %v", path, o.Logged, logged)
				}
				// A Verilog parameter is the same kind of object with
				// its own declaration kind: t11_v_param.
				if dc.Kind != DeclGeneric && dc.Kind != DeclParam {
					t.Errorf("%s: declaration kind %s, want generic or parameter", path, dc.Kind)
				}
				if g.Type != "" && !strings.EqualFold(f.Types[dc.Type].Name, g.Type) {
					t.Errorf("%s: type %s, truth says %s", path, f.Types[dc.Type].Name, g.Type)
				}
				if g.Width != 0 && dc.Size != g.Width {
					t.Errorf("%s: %d bytes, truth says %d elements", path, dc.Size, g.Width)
				}
				ch, err := f.Changes(o)
				if err != nil {
					t.Fatal(err)
				}
				if g.Logged != nil && !*g.Logged {
					if len(ch) != 0 {
						t.Errorf("%s: %d changes, truth says not logged", path, len(ch))
					}
					continue
				}
				if len(ch) != 1 || ch[0].TimePS != 0 {
					t.Errorf("%s: %d changes, want one at time zero", path, len(ch))
					continue
				}
				v, err := f.Decode(dc, ch[0].Data)
				if err != nil {
					t.Fatal(err)
				}
				if got := f.String(v); !sameValue(f, dc.Type, value, got) {
					t.Errorf("%s = %s, truth says %s", path, got, value)
				}
			}

			// Variables: an object in the process scope, with no second
			// handle. The database does not log a declared variable's
			// changes, nor its initial value: t6_var_int changes v twice
			// and holds no record for it. A loop index gets one record
			// at time zero, holding zero rather than the loop's first
			// value: t5_tr1000 iterates from 1 and records 0. A generate
			// index or an architecture constant records its value.
			for _, vr := range tr.Variables {
				path := vr.Scope + "." + vr.Name
				o, ok := objByPath[path]
				if !ok || !o.Generic {
					t.Errorf("truth variable %s is not a variable object", path)
					continue
				}
				dc := f.Decls[o.Decl]
				if got := strings.ToLower(f.Types[dc.Type].Name); got != vr.Type {
					t.Errorf("%s: type %s, truth says %s", path, got, vr.Type)
				}
				ch, err := f.Changes(o)
				if err != nil {
					t.Fatal(err)
				}
				switch vr.Kind {
				case "variable":
					if dc.Kind != DeclVariable {
						t.Errorf("%s: declaration kind %s, want variable", path, dc.Kind)
					}
					if len(ch) != 0 {
						t.Errorf("%s: %d records, a declared variable has had none", path, len(ch))
					}
				case "loop", "constant":
					if dc.Kind != DeclConstant {
						t.Errorf("%s: declaration kind %s, want constant", path, dc.Kind)
					}
					if vr.Logged != nil && !*vr.Logged {
						if len(ch) != 0 || o.Logged {
							t.Errorf("%s: %d records, a package constant has had none", path, len(ch))
						}
						continue
					}
					if len(ch) != 1 || ch[0].TimePS != 0 {
						t.Errorf("%s: %d records, a loop index has had one at time zero", path, len(ch))
						continue
					}
					if vr.Value == "" {
						continue
					}
					v, err := f.Decode(dc, ch[0].Data)
					if err != nil {
						t.Fatal(err)
					}
					if v.Scalar != vr.Value {
						t.Errorf("%s = %s, truth says %s", path, v.Scalar, vr.Value)
					}
				default:
					t.Errorf("%s: unknown truth kind %q", path, vr.Kind)
				}
			}
		})
	}
}
