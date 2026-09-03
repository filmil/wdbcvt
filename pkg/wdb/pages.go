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
	// Time is the simulation time of the change in the file's time
	// unit; see File.TimeUnit.
	Time uint64
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
	// from 600000 to 1000000 and it holds T0 600000, T1 400000.
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
		r := Record{Time: c.u64(), Key: c.u32()}
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
	// Time is in the file's time unit; see File.TimeUnit.
	Time uint64
	Data []byte
}

// Changes returns every recorded value of the object, one per write,
// in time order. An object whose arena was never written has no
// changes; t6_var_int shows a declared process variable in that state.
//
// A VHDL record is one write of a contiguous part of the value, keyed
// by the byte address of its first byte within the handle space: the
// whole value at time 0 and after a whole assignment, or the bytes of
// a field, slice or element after a partial assignment, with the parts
// one delta assigns merged when they are adjacent. Found by the ctl
// record of //hdl/counter:sim, the offsets by the t32 cases. A write
// of chunkWhole bytes or more is not one record. It is written as the
// run of chunks Chunks describes for its own length from its own
// address, each a record of its own, and a chunk that would cross an
// arena boundary is split there and goes on in the next arena at key
// 0. Found by t9_vec292 and t9_vec12000, the rule by the t10_vec sizes,
// and its use for a partial write by t32_wide_slice__.
//
// Changes overlays every write onto the value in file order, which is
// time order in every observed file, and emits one value per write.
// It checks that the first write covers the object, and that every
// write of chunkWhole bytes or more is the run of chunks the rule
// predicts.
//
// A Verilog object is written by word pairs instead; see changesVerilog.
func (f *File) Changes(o Object) ([]Change, error) {
	if !o.Logged {
		return nil, nil
	}
	dc := f.Decls[o.Decl]
	n, err := f.recordBytes(dc)
	if err != nil {
		return nil, err
	}
	size := uint64(n)
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
	if f.verilog(dc.Type) {
		return f.changesVerilog(o, start, end)
	}
	// Collect every record that overlaps the object, in file order, and
	// order them by time. A record keeps its place among those of its
	// time, so two writes of one byte in two deltas of one time stay in
	// write order: t32_rec_delta___.
	type rec struct {
		time uint64
		addr uint64
		data []byte
	}
	var recs []rec
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
				recs = append(recs, rec{r.Time, addr, r.Data})
			}
		}
	}
	if len(recs) == 0 {
		return nil, fmt.Errorf("object handle %#x is logged but has no records", o.Handle)
	}
	sort.SliceStable(recs, func(i, j int) bool { return recs[i].time < recs[j].time })
	// A write is the longest run of records of one time, each starting
	// where the previous one ends, whose addresses are the chunks the
	// rule predicts for the run's length; a record that starts no such
	// run is a write of its own and is shorter than a chunked write.
	write := func(i int) (int, uint64, error) {
		total := uint64(len(recs[i].data))
		j := i + 1
		for j < len(recs) && recs[j].time == recs[i].time && recs[j].addr == recs[j-1].addr+uint64(len(recs[j-1].data)) {
			total += uint64(len(recs[j].data))
			j++
		}
		for ; j > i+1; j-- {
			want := chunkStarts(recs[i].addr, total)
			if len(want) == j-i {
				ok := true
				for k, w := range want {
					ok = ok && w == recs[i+k].addr
				}
				if ok {
					return j, total, nil
				}
			}
			total -= uint64(len(recs[j-1].data))
		}
		if total >= chunkWhole {
			return 0, 0, fmt.Errorf("object handle %#x: a record of %d bytes at %#x, time %d, is not a run of chunks", o.Handle, total, recs[i].addr, recs[i].time)
		}
		return j, total, nil
	}
	cur := make([]byte, size)
	var out []Change
	for i := 0; i < len(recs); {
		j, total, err := write(i)
		if err != nil {
			return nil, err
		}
		// The first write is the initial value of the whole signal;
		// a port bound to a slice at its start, t9_port_sliceto_,
		// sees a first write longer than its own value.
		if len(out) == 0 && (recs[i].addr > start || recs[i].addr+total < end) {
			return nil, fmt.Errorf("object handle %#x with %d bytes has a first write of %d bytes at %#x, which does not cover it", o.Handle, size, total, recs[i].addr)
		}
		for ; i < j; i++ {
			r := recs[i]
			lo, hi := r.addr, r.addr+uint64(len(r.data))
			if lo < start {
				lo = start
			}
			if hi > end {
				hi = end
			}
			copy(cur[lo-start:hi-start], r.data[lo-r.addr:hi-r.addr])
		}
		out = append(out, Change{Time: recs[i-1].time, Data: append([]byte(nil), cur...)})
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

// chunkStarts lists the record addresses Chunks predicts for a value of
// size bytes at handle, allowing for the split of a chunk at an arena
// boundary: the chunk goes on in the next arena as a record of its own
// at key 0.
func chunkStarts(handle, size uint64) []uint64 {
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
	return want
}
