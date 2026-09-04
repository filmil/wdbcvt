// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestStream holds the one pass decoder to the per object one: for
// every case and every design, Stream must produce exactly the changes
// Changes produces, object by object, in the same order. Changes is
// the reading every corpus truth is checked against, so this is what
// makes the streaming conversion trustworthy.
func TestStream(t *testing.T) {
	root := filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"))
	dirs := corpusCases(t)
	for _, d := range designs {
		dirs = append(dirs, filepath.Join(root, d))
	}
	for _, dir := range dirs {
		dir := dir
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			f, err := ReadFile(filepath.Join(dir, "sim.wdb"))
			if err != nil {
				t.Fatal(err)
			}
			got := make([][]Change, len(f.Objects))
			err = f.Stream(func(obj int, tm uint64, data []byte) error {
				got[obj] = append(got[obj], Change{Time: tm, Data: append([]byte(nil), data...)})
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			for i, o := range f.Objects {
				want, err := f.Changes(o)
				if err != nil {
					t.Fatal(err)
				}
				if len(got[i]) != len(want) {
					t.Errorf("object %d (%s): %d changes streamed, %d decoded",
						i, f.ObjectPath(o), len(got[i]), len(want))
					continue
				}
				for j := range want {
					if got[i][j].Time != want[j].Time || !bytes.Equal(got[i][j].Data, want[j].Data) {
						t.Errorf("object %d (%s) change %d: streamed %d %x, decoded %d %x",
							i, f.ObjectPath(o), j, got[i][j].Time, got[i][j].Data,
							want[j].Time, want[j].Data)
						break
					}
				}
			}
		})
	}
}

// TestStreamTimeOrder checks the property the FST writer depends on:
// the times never go backwards.
func TestStreamTimeOrder(t *testing.T) {
	for _, dir := range corpusCases(t) {
		dir := dir
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()
			f, err := ReadFile(filepath.Join(dir, "sim.wdb"))
			if err != nil {
				t.Fatal(err)
			}
			var last uint64
			first := true
			err = f.Stream(func(obj int, tm uint64, data []byte) error {
				if !first && tm < last {
					t.Fatalf("time %d after %d", tm, last)
				}
				last, first = tm, false
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
