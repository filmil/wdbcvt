// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"fmt"
	"sort"
)

// A Mask marks byte offsets that carry no information about the design.
//
// A waveform database usually embeds things that change between two runs
// of the same simulation: a timestamp, a host name, a path, a build id.
// Those offsets differ for reasons unrelated to the design, and comparing
// two different designs without excluding them produces confident and
// wrong conclusions. See docs/corpus.md.
type Mask struct {
	off map[int64]bool
}

// NoiseMask returns the offsets at which a and b differ.
//
// Both inputs must be the same design simulated twice. Any offset that
// differs is by definition not carrying design information, because the
// design did not change.
//
// Where the two runs have different lengths, every offset from the
// shorter length onward is masked as well: a length that is not stable
// across runs cannot be read as a design property either.
func NoiseMask(a, b []byte) Mask {
	m := Mask{off: map[int64]bool{}}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			m.off[int64(i)] = true
		}
	}
	for i := n; i < len(a) || i < len(b); i++ {
		m.off[int64(i)] = true
	}
	return m
}

// Contains reports whether off is masked.
func (m Mask) Contains(off int64) bool { return m.off[off] }

// Len is how many offsets the mask covers.
func (m Mask) Len() int { return len(m.off) }

// Spans renders the mask as ascending contiguous ranges, which is how a
// human reads it.
func (m Mask) Spans() []Span {
	if len(m.off) == 0 {
		return nil
	}
	offs := make([]int64, 0, len(m.off))
	for o := range m.off {
		offs = append(offs, o)
	}
	sort.Slice(offs, func(i, j int) bool { return offs[i] < offs[j] })

	var out []Span
	start, prev := offs[0], offs[0]
	for _, o := range offs[1:] {
		if o != prev+1 {
			out = append(out, Span{Offset: start, Length: int(prev - start + 1)})
			start = o
		}
		prev = o
	}
	return append(out, Span{Offset: start, Length: int(prev - start + 1)})
}

// Span is a contiguous run of bytes.
type Span struct {
	Offset int64
	Length int
}

// A Delta is one contiguous run of bytes that differs between two
// databases, with the bytes from each side. Where the two files have
// different lengths, the shorter side's bytes are empty past its end.
type Delta struct {
	Offset int64
	A      []byte
	B      []byte
}

// String renders a delta as one line, bytes in hex.
func (d Delta) String() string {
	return fmt.Sprintf("%#010x  a=% x  b=% x", d.Offset, d.A, d.B)
}

// Compare returns the runs of bytes that differ between a and b, skipping
// every offset the mask covers, together with the difference in length.
//
// A nil mask compares every offset, which is what the two runs of one
// design want. Two different designs should always be compared through a
// mask built by NoiseMask.
func Compare(a, b []byte, m Mask) (deltas []Delta, lenDelta int64) {
	lenDelta = int64(len(b)) - int64(len(a))

	n := len(a)
	if len(b) > n {
		n = len(b)
	}

	at := func(s []byte, i int) (byte, bool) {
		if i < len(s) {
			return s[i], true
		}
		return 0, false
	}

	var cur *Delta
	for i := 0; i < n; i++ {
		if m.Contains(int64(i)) {
			cur = nil
			continue
		}
		av, aok := at(a, i)
		bv, bok := at(b, i)
		if aok && bok && av == bv {
			cur = nil
			continue
		}
		if cur == nil {
			deltas = append(deltas, Delta{Offset: int64(i)})
			cur = &deltas[len(deltas)-1]
		}
		if aok {
			cur.A = append(cur.A, av)
		}
		if bok {
			cur.B = append(cur.B, bv)
		}
	}
	return deltas, lenDelta
}
