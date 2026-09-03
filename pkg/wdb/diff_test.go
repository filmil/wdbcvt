// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"bytes"
	"reflect"
	"testing"
)

func TestNoiseMaskFindsDifferingOffsets(t *testing.T) {
	a := []byte{1, 2, 3, 4, 5}
	b := []byte{1, 9, 3, 9, 5}
	m := NoiseMask(a, b)
	if m.Len() != 2 {
		t.Fatalf("Len = %d, want 2", m.Len())
	}
	for _, off := range []int64{1, 3} {
		if !m.Contains(off) {
			t.Errorf("offset %d should be masked", off)
		}
	}
	for _, off := range []int64{0, 2, 4} {
		if m.Contains(off) {
			t.Errorf("offset %d should not be masked", off)
		}
	}
}

func TestNoiseMaskIdenticalRunsMaskNothing(t *testing.T) {
	a := []byte{1, 2, 3}
	if m := NoiseMask(a, a); m.Len() != 0 {
		t.Errorf("Len = %d, want 0 for two identical runs", m.Len())
	}
}

func TestNoiseMaskCoversTheLengthDifference(t *testing.T) {
	// A length that is not stable across runs cannot be read as a
	// design property either, so the tail is masked too.
	m := NoiseMask([]byte{1, 2}, []byte{1, 2, 3, 4})
	if !m.Contains(2) || !m.Contains(3) {
		t.Errorf("the tail past the shorter length should be masked, got spans %v", m.Spans())
	}
}

func TestMaskSpansAreContiguousAndSorted(t *testing.T) {
	a := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	b := []byte{1, 1, 0, 0, 1, 0, 1, 1}
	got := NoiseMask(a, b).Spans()
	want := []Span{{0, 2}, {4, 1}, {6, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Spans() = %v, want %v", got, want)
	}
}

func TestCompareWithoutMask(t *testing.T) {
	a := []byte{1, 2, 3, 4}
	b := []byte{1, 9, 9, 4}
	d, ld := Compare(a, b, Mask{})
	if ld != 0 {
		t.Errorf("lenDelta = %d, want 0", ld)
	}
	if len(d) != 1 {
		t.Fatalf("deltas = %v, want one contiguous run", d)
	}
	if d[0].Offset != 1 || !bytes.Equal(d[0].A, []byte{2, 3}) || !bytes.Equal(d[0].B, []byte{9, 9}) {
		t.Errorf("delta = %+v, want offset 1 a=[2 3] b=[9 9]", d[0])
	}
}

func TestCompareSkipsMaskedOffsets(t *testing.T) {
	// The design difference is at offset 3. Offset 1 is noise, and a
	// mask must keep it out of the result.
	run1 := []byte{1, 0xAA, 3, 4}
	run2 := []byte{1, 0xBB, 3, 4}
	m := NoiseMask(run1, run2)

	other := []byte{1, 0xCC, 3, 7}
	d, _ := Compare(run1, other, m)
	if len(d) != 1 {
		t.Fatalf("deltas = %v, want only the unmasked difference", d)
	}
	if d[0].Offset != 3 {
		t.Errorf("delta offset = %d, want 3 (offset 1 is noise)", d[0].Offset)
	}
}

func TestCompareMaskBreaksARun(t *testing.T) {
	// A masked offset in the middle must split one run into two, not
	// silently join them.
	a := []byte{0, 0, 0, 0}
	b := []byte{1, 1, 1, 1}
	m := Mask{off: map[int64]bool{1: true}}
	d, _ := Compare(a, b, m)
	if len(d) != 2 {
		t.Fatalf("deltas = %v, want two runs split by the mask", d)
	}
	if d[0].Offset != 0 || d[1].Offset != 2 {
		t.Errorf("offsets = %d and %d, want 0 and 2", d[0].Offset, d[1].Offset)
	}
}

func TestCompareReportsLengthDifference(t *testing.T) {
	d, ld := Compare([]byte{1, 2}, []byte{1, 2, 3}, Mask{})
	if ld != 1 {
		t.Errorf("lenDelta = %d, want 1", ld)
	}
	if len(d) != 1 || d[0].Offset != 2 {
		t.Fatalf("deltas = %v, want one at offset 2", d)
	}
	if len(d[0].A) != 0 || !bytes.Equal(d[0].B, []byte{3}) {
		t.Errorf("delta = %+v, want the shorter side empty", d[0])
	}
}

func TestCompareIdenticalFiles(t *testing.T) {
	a := []byte{1, 2, 3}
	d, ld := Compare(a, a, Mask{})
	if len(d) != 0 || ld != 0 {
		t.Errorf("Compare of a file with itself = %v, %d; want no differences", d, ld)
	}
}
