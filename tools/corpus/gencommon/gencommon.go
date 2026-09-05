// SPDX-License-Identifier: Apache-2.0

// Package gencommon holds what every corpus generator needs: where the
// corpus is, how a case is written out, and the normalisation the truth
// files apply to a declaration.
package gencommon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Workspace is the directory the corpus is written into. Under
// `bazel run` the program executes from its runfiles, and
// BUILD_WORKSPACE_DIRECTORY names the source tree; outside Bazel the
// current directory is the workspace.
func Workspace() string {
	if w := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); w != "" {
		return w
	}
	w, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return w
}

// Root is the corpus directory.
func Root() string { return filepath.Join(Workspace(), "hdl", "corpus") }

// hdrs gives the comment marker of the language a file is in. A VHDL
// file with a Verilog comment on its first line does not compile, and
// the tiers before 60 were fixed by hand after every run.
var hdrs = map[string]string{".vhdl": "--", ".vhd": "--"}

// Header is the SPDX line a generated source file starts with.
func Header(filename string) string {
	mark, ok := hdrs[filepath.Ext(filename)]
	if !ok {
		mark = "//"
	}
	return mark + " SPDX-License-Identifier: Apache-2.0\n"
}

// File is one source file of a case.
type File struct {
	Name string
	Body string
}

// TS is the timescale directive every SystemVerilog case starts with.
const TS = "`timescale 1ns / 1ps\n\n"

// Sig declares a signal of the truth. Every other key of the
// declaration follows as key and value pairs.
func Sig(scope, name, typ string, width int, kv ...any) *Obj {
	o := O("scope", scope, "name", name, "width", width, "type", typ)
	for i := 0; i < len(kv); i += 2 {
		o.Set(kv[i].(string), kv[i+1])
	}
	return o
}

// Tr is one value change of the truth.
func Tr(t int, signal, value string) *Obj {
	return O("time_ns", t, "signal", signal, "value", value)
}

// Bits formats a value as n binary digits.
func Bits(v uint64, n int) string {
	s := fmt.Sprintf("%b", v)
	if len(s) < n {
		s = strings.Repeat("0", n-len(s)) + s
	}
	return s
}

// nets are the Verilog net keywords. The database names no type for a
// vector, memory or struct field declared without a typedef, and gives a
// scalar reg, wire or logic the predefined type logic (bit for a two
// state scalar). The truth keeps the source keyword in "declared" and
// puts the database name in "type".
var nets = []string{"reg", "logic", "bit", "wire", "uwire", "wand", "wor",
	"tri", "triand", "trior", "tri0", "tri1", "supply0", "supply1"}

// Norm rewrites a declaration the way the database names it.
func Norm(sg *Obj) {
	typ := sg.Str("type")
	switch {
	case contains(nets, typ):
		sg.Set("declared", typ)
		if sg.Int("width") > 1 {
			sg.Set("type", "")
		} else if typ != "logic" && typ != "bit" {
			sg.Set("type", "logic")
		}
	case typ == "memory":
		sg.Set("declared", "memory of "+sg.Str("element_type"))
		sg.Set("type", "")
		sg.Set("element_type", "")
	}
	for _, f := range sg.Fields() {
		Norm(f)
	}
}

func contains(l []string, s string) bool {
	for _, e := range l {
		if e == s {
			return true
		}
	}
	return false
}

// WithX prefixes the records of a Verilog source with the all X record
// every declared variable holds at time zero. A .v source initialises
// every variable from an implicit initial block, so a four state object
// records all X and then its initial value; a real is not four state. A
// SystemVerilog source does that only for an enum or string initializer,
// and those cases list the X record themselves.
func WithX(files []File, signals, transitions []*Obj) []*Obj {
	verilog := false
	for _, f := range files {
		if strings.HasSuffix(f.Name, ".v") {
			verilog = true
		}
	}
	if !verilog {
		return transitions
	}
	var out []*Obj
	seen := map[string]bool{}
	for _, x := range transitions {
		name := x.Str("signal")
		if !seen[name] {
			seen[name] = true
			var sg *Obj
			for _, g := range signals {
				if g.Str("name") == name || g.Str("scope")+"."+g.Str("name") == name {
					sg = g
					break
				}
			}
			v := fmt.Sprint(x.Get("value"))
			if sg.Str("type") != "real" && !strings.HasPrefix(v, "X") && !strings.HasPrefix(v, "(X") {
				var val string
				if n := sg.Int("elements"); n > 0 {
					parts := make([]string, n)
					for i := range parts {
						parts[i] = strings.Repeat("X", sg.Int("element_width"))
					}
					val = "(" + strings.Join(parts, ", ") + ")"
				} else {
					val = strings.Repeat("X", sg.Int("width"))
				}
				out = append(out, Tr(0, name, val))
			}
		}
		out = append(out, x)
	}
	return out
}

// MemTrs gives the records of a memory initialised element by element:
// one change per element, all at time zero, after the all X record.
// desc puts m[0] at the right.
func MemTrs(w, n, at int, val uint64, desc bool) []*Obj {
	x, z := strings.Repeat("X", w), Bits(0, w)
	cur := make([]string, n)
	for i := range cur {
		cur[i] = x
	}
	out := []*Obj{Tr(0, "m", "("+strings.Join(cur, ", ")+")")}
	idx := func(i int) int {
		if desc {
			return n - 1 - i
		}
		return i
	}
	for i := 0; i < n; i++ {
		cur[idx(i)] = z
		out = append(out, Tr(0, "m", "("+strings.Join(cur, ", ")+")"))
	}
	cur[idx(at)] = Bits(val, w)
	out = append(out, Tr(50, "m", "("+strings.Join(cur, ", ")+")"))
	return out
}

// Case is one corpus case as a generator describes it.
type Case struct {
	Name    string
	Axis    string
	Differs string
	Files   []File
	Signals []*Obj
	Trs     []*Obj
	End     int  // zero means 100, the usual end time
	Extra   *Obj // keys the truth carries beside the usual ones
	NoX     bool // the source does not write the all X records of a .v
}

// Emit writes the sources, the BUILD file and the truth of one case.
func Emit(c Case) {
	if len(c.Name) != 16 {
		panic("case name is not 16 characters: " + c.Name)
	}
	d := filepath.Join(Root(), c.Name)
	must(os.MkdirAll(d, 0o755))
	var srcs []string
	for _, f := range c.Files {
		must(os.WriteFile(filepath.Join(d, f.Name), []byte(Header(f.Name)+f.Body), 0o644))
		srcs = append(srcs, f.Name)
	}
	var b strings.Builder
	b.WriteString("# SPDX-License-Identifier: Apache-2.0\n\n" +
		"load(\"//build:wdb_case.bzl\", \"wdb_case\")\n\n" +
		"package(default_visibility = [\"//visibility:public\"])\n\n" +
		"wdb_case(\n    name = \"" + c.Name + "\",\n    srcs = [\n")
	for _, s := range srcs {
		b.WriteString("        \"" + s + "\",\n")
	}
	b.WriteString("    ],\n)\n")
	must(os.WriteFile(filepath.Join(d, "BUILD.bazel"), []byte(b.String()), 0o644))

	for _, sg := range c.Signals {
		Norm(sg)
	}
	trs := c.Trs
	if !c.NoX {
		trs = WithX(c.Files, c.Signals, trs)
	}
	end := c.End
	if end == 0 {
		end = 100
	}
	t := O("case", c.Name, "axis", c.Axis, "differs_from", c.Differs,
		"end_time_ns", end, "signals", objs(c.Signals), "transitions", objs(trs))
	if c.Extra != nil {
		t.Update(c.Extra)
	}
	f, err := os.OpenFile(filepath.Join(d, "truth.json"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	must(err)
	must(WriteJSON(f, t))
	must(f.Close())
	fmt.Println(c.Name)
}

// objs keeps an empty list an empty list rather than a null.
func objs(l []*Obj) []*Obj {
	if l == nil {
		return []*Obj{}
	}
	return l
}

// WriteFile writes one more file into a case directory, for the cases
// that carry an xsim script.
func WriteFile(caseName, fileName, body string) {
	must(os.WriteFile(filepath.Join(Root(), caseName, fileName), []byte(body), 0o644))
}

// PatchBuild rewrites the BUILD file of a case, which is how a case adds
// an attribute the plain wdb_case call does not carry.
func PatchBuild(caseName string, replace func(string) string) {
	p := filepath.Join(Root(), caseName, "BUILD.bazel")
	b, err := os.ReadFile(p)
	must(err)
	must(os.WriteFile(p, []byte(replace(string(b))), 0o644))
}

// Fill substitutes the %(key)s placeholders of a template.
func Fill(tmpl string, kv ...string) string {
	if len(kv)%2 != 0 {
		panic("Fill: odd number of arguments")
	}
	rep := make([]string, 0, len(kv))
	for i := 0; i < len(kv); i += 2 {
		rep = append(rep, "%("+kv[i]+")s", kv[i+1])
	}
	out := strings.NewReplacer(rep...).Replace(tmpl)
	if i := strings.Index(out, "%("); i >= 0 {
		panic("Fill: placeholder left in template: " + out[i:min(i+20, len(out))])
	}
	return out
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
