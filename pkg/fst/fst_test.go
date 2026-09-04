// SPDX-License-Identifier: Apache-2.0

package fst

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWrite writes a small file and checks that libfst produced one:
// the reader side of the check comes later.
func TestWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.fst")
	w, err := Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w.SetTimescale(-12)
	w.PushScope("tb")
	s := w.Var(VarWire, "s", 1, 0)
	v := w.Var(VarWire, "v", 8, 0)
	w.PopScope()
	w.Time(0)
	w.Value(s, "x")
	w.Value(v, "xxxxxxxx")
	w.Time(50000)
	w.Value(s, "1")
	w.Value(v, "10100101")
	w.Time(100000)
	w.Close()
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() == 0 {
		t.Fatal("empty file")
	}
	t.Logf("%s: %d bytes", path, st.Size())
	vars, start, end, err := ReadVars(path)
	if err != nil {
		t.Fatal(err)
	}
	if start != 0 || end != 100000 {
		t.Errorf("times %d..%d, want 0..100000", start, end)
	}
	want := []Var{{Path: "tb.s", Bits: 1}, {Path: "tb.v", Bits: 8}}
	if len(vars) != len(want) {
		t.Fatalf("%d variables, want %d: %+v", len(vars), len(want), vars)
	}
	for i, v := range vars {
		if v.Path != want[i].Path || v.Bits != want[i].Bits {
			t.Errorf("variable %d is %v, want %v", i, v, want[i])
		}
	}
}

// TestReadChanges writes one variable of each shape and reads every
// value change back through libfst.
func TestReadChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.fst")
	w, err := Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w.SetTimescale(-12)
	w.PushScope("tb")
	v := w.Var(VarWire, "v", 4, 0)
	alias := w.Var(VarWire, "same", 4, v)
	r := w.Var(VarReal, "r", 64, 0)
	s := w.Var(VarString, "s", 0, 0)
	w.PopScope()
	w.Time(0)
	w.Value(v, "xxxx")
	w.Real(r, 0)
	w.Str(s, "IDLE")
	w.Time(50000)
	w.Value(v, "0101")
	w.Real(r, 1.5)
	w.Str(s, "RUN")
	w.Time(100000)
	w.Close()

	vars, changes, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(vars) != 4 {
		t.Fatalf("%d variables, want 4: %+v", len(vars), vars)
	}
	if vars[1].Handle != vars[0].Handle {
		t.Errorf("the alias has handle %d, the variable %d", vars[1].Handle, vars[0].Handle)
	}
	_ = alias
	got := map[string][]string{}
	byHandle := map[uint32]string{}
	for _, vr := range vars {
		if _, ok := byHandle[vr.Handle]; !ok {
			byHandle[vr.Handle] = vr.Path
		}
	}
	for _, c := range changes {
		got[byHandle[c.Handle]] = append(got[byHandle[c.Handle]], fmt.Sprintf("%d=%s", c.Time, c.Value))
	}
	want := map[string]string{
		"tb.v": "0=xxxx 50000=0101",
		"tb.r": "0=0 50000=1.5",
		"tb.s": "0=IDLE 50000=RUN",
	}
	for path, w := range want {
		if strings.Join(got[path], " ") != w {
			t.Errorf("%s: %q, want %q", path, strings.Join(got[path], " "), w)
		}
	}
}
