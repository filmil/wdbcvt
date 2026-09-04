// SPDX-License-Identifier: Apache-2.0

package fst

import (
	"os"
	"path/filepath"
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
	want := []Var2{{Path: "tb.s", Bits: 1}, {Path: "tb.v", Bits: 8}}
	if len(vars) != len(want) {
		t.Fatalf("%d variables, want %d: %+v", len(vars), len(want), vars)
	}
	for i, v := range vars {
		if v.Path != want[i].Path || v.Bits != want[i].Bits {
			t.Errorf("variable %d is %v, want %v", i, v, want[i])
		}
	}
}
