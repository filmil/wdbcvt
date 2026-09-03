// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"sort"
)

// Record is one value change inside a page.
type Record struct {
	// TimePS is the simulation time of the change in picoseconds.
	TimePS uint64
	// Key names the object within the page; see Object.Key.
	Key uint32
	// Data is the value, encoded as docs/format.md describes per kind.
	Data []byte
}

// Page is one inflated value page.
type Page struct {
	// Ref locates the page's zlib stream in the file.
	Ref PageRef
	// T0 and T1 are the two times the page header holds. T0 is the time
	// of the first record and T1 the time of the last record minus T0.
	// The second page of t5_tr1000 shows the difference: its records run
	// from 600000 to 1000000 ps and it holds T0 600000, T1 400000.
	T0, T1  uint64
	Records []Record
}

// readPage inflates one page and splits it into records.
func readPage(d []byte, ref PageRef) (Page, error) {
	pg := Page{Ref: ref}
	zr, err := zlib.NewReader(bytes.NewReader(d[ref.Offset : ref.Offset+ref.CompressedLen]))
	if err != nil {
		return pg, fmt.Errorf("page at %#x: %w", ref.Offset, err)
	}
	defer zr.Close()
	raw, err := io.ReadAll(zr)
	if err != nil {
		return pg, fmt.Errorf("page at %#x: inflate: %w", ref.Offset, err)
	}
	if len(raw) != pageLen {
		return pg, fmt.Errorf("page at %#x inflates to %d bytes, want %d", ref.Offset, len(raw), pageLen)
	}
	c := &cursor{b: raw}
	pg.T0 = c.u64()
	pg.T1 = c.u64()
	n := int(c.u32())
	for i := 0; i < n; i++ {
		r := Record{TimePS: c.u64(), Key: c.u32()}
		l := int(c.u32())
		if !c.need(l) {
			break
		}
		r.Data = raw[c.p : c.p+l]
		c.p += l
		pg.Records = append(pg.Records, r)
	}
	if c.err != nil {
		return pg, fmt.Errorf("page at %#x: record %d: %w", ref.Offset, len(pg.Records), c.err)
	}
	for _, b := range raw[c.p:] {
		if b != 0 {
			return pg, fmt.Errorf("page at %#x has non-zero padding after %d records", ref.Offset, n)
		}
	}
	return pg, nil
}

// Change is one value of one object at one time.
type Change struct {
	TimePS uint64
	Data   []byte
}

// Changes returns every recorded value of the object, in page order,
// which is time order in every observed file. An object whose arena
// was never written has no changes; t6_var_int shows a declared process
// variable in that state.
//
// A value of chunkWhole bytes or more is not one record. It is written
// as the run of chunks Chunks describes, each a record of its own whose
// key is the byte address of the chunk within the handle space, and a
// chunk that would cross an arena boundary is split there and goes on
// in the next arena at key 0. Changes reassembles the run into one
// value per change, and checks that the run is the one Chunks predicts
// when the object owns the whole value. Found by t9_vec292 and
// t9_vec12000, the rule by the t10_vec sizes.
func (f *File) Changes(o Object) ([]Change, error) {
	if !o.Logged {
		return nil, nil
	}
	size := uint64(f.Decls[o.Decl].Size)
	if size == 0 {
		size = 1
	}
	// The object's bytes are [start, end) of the value recorded at its
	// handle; Offset is nonzero for a port bound to a slice.
	start := o.Handle + uint64(o.Offset)
	end := start + size
	first, last := int(start/arenaSpan), int((end-1)/arenaSpan)
	if last >= len(f.Arenas) {
		return nil, fmt.Errorf("object handle %#x with %d bytes reaches arena %d of %d", o.Handle, size, last, len(f.Arenas))
	}
	// Collect every record that overlaps the object, by chunk address,
	// in page order. A wide value is written as a run of chunks, each a
	// record keyed by its byte address, and a chunk never crosses an
	// arena boundary: t9_vec292, t9_vec4096.
	type chunk struct {
		time uint64
		data []byte
	}
	runs := map[uint64][]chunk{}
	var addrs []uint64
	for k := first; k <= last; k++ {
		if f.Arenas[k].Offset == 0 {
			return nil, fmt.Errorf("object handle %#x with %d bytes spans arena %d, which was never written", o.Handle, size, k)
		}
		for _, pg := range f.Pages[k] {
			for _, r := range pg.Records {
				addr := uint64(k)*arenaSpan + uint64(r.Key)
				if addr >= end || addr+uint64(len(r.Data)) <= start {
					continue
				}
				if _, seen := runs[addr]; !seen {
					addrs = append(addrs, addr)
				}
				runs[addr] = append(runs[addr], chunk{r.TimePS, r.Data})
			}
		}
	}
	sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })
	if len(addrs) == 0 {
		return nil, fmt.Errorf("object handle %#x is logged but has no records", o.Handle)
	}
	if o.Offset == 0 {
		if err := checkChunks(o.Handle, size, addrs); err != nil {
			return nil, err
		}
	}
	// The i-th record at each address belongs to the i-th change.
	n := len(runs[addrs[0]])
	var out []Change
	for i := 0; i < n; i++ {
		c := Change{TimePS: runs[addrs[0]][i].time}
		next := start
		for _, addr := range addrs {
			if len(runs[addr]) != n {
				return nil, fmt.Errorf("object handle %#x: chunk at %#x has %d records, the first chunk has %d", o.Handle, addr, len(runs[addr]), n)
			}
			ch := runs[addr][i]
			if ch.time != c.TimePS {
				return nil, fmt.Errorf("object handle %#x: change %d: chunk at %#x is at %d ps, the first chunk at %d ps", o.Handle, i, addr, ch.time, c.TimePS)
			}
			lo, hi := addr, addr+uint64(len(ch.data))
			if lo < start {
				lo = start
			}
			if hi > end {
				hi = end
			}
			if lo != next {
				return nil, fmt.Errorf("object handle %#x: change %d: chunk at %#x follows one ending at %#x", o.Handle, i, addr, next)
			}
			c.Data = append(c.Data, ch.data[lo-addr:hi-addr]...)
			next = hi
		}
		if next != end {
			return nil, fmt.Errorf("object handle %#x: change %d: chunks end at %#x, the value ends at %#x", o.Handle, i, next, end)
		}
		out = append(out, c)
	}
	return out, nil
}

// chunkWhole is the smallest value size that is written as chunks.
// t10_vec274 is one record and t10_vec275 is two.
const chunkWhole = 275

// Chunks returns how a value of size bytes is split into records: the
// number of chunks and the length of every chunk but the last, which
// takes the remainder. A value below chunkWhole bytes is one record.
//
// The rule was fitted to 55 values from 200 to 30000 bytes, the t9_vec
// and t10_vec cases, and holds for every one: with s = size + 24 the
// count is 2 * ceil(s / 299), and the chunk length is size / count
// rounded down. t10_vec275 gives 2 chunks of 137 and 138, t10_vec276
// gives 4 of 69, t10_vec574 and t10_vec575 sit either side of 4 and 6,
// t10_vec30000 gives 202 of 148 with a last chunk of 252. Why the
// constants are 24 and 299 is not known.
func Chunks(size uint64) (count, length uint64) {
	if size < chunkWhole {
		return 1, size
	}
	s := size + 24
	count = 2 * ((s + 298) / 299)
	return count, size / count
}

// checkChunks refuses a run of record addresses that is not the one
// Chunks predicts for a value of size bytes at handle, allowing for the
// split of a chunk at an arena boundary.
func checkChunks(handle, size uint64, addrs []uint64) error {
	count, length := Chunks(size)
	var want []uint64
	next := handle
	for i := uint64(0); i < count; i++ {
		at := handle + i*length
		// An arena boundary inside the previous chunk starts a record.
		if b := (next/arenaSpan + 1) * arenaSpan; b < at {
			want = append(want, b)
		}
		want = append(want, at)
		next = at
	}
	if b := (next/arenaSpan + 1) * arenaSpan; b < handle+size {
		want = append(want, b)
	}
	if len(want) != len(addrs) {
		return fmt.Errorf("object handle %#x with %d bytes has records at %d addresses, the chunk rule predicts %d (%d chunks of %d)", handle, size, len(addrs), len(want), count, length)
	}
	for i := range want {
		if want[i] != addrs[i] {
			return fmt.Errorf("object handle %#x with %d bytes has a record at %#x where the chunk rule predicts %#x", handle, size, addrs[i], want[i])
		}
	}
	return nil
}
