// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
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
// A record is one write of a contiguous part of the value, keyed by
// the byte address of its first byte within the handle space: the
// whole value at time 0 and after a whole assignment, or the bytes of
// a field, slice or element after a partial assignment, with the parts
// one delta assigns merged when they are adjacent. Found by the ctl
// record of //hdl/counter:sim, the offsets by the t32 cases. A Verilog
// value is word pairs, and a partial write covers whole pairs:
// t11_v_vec64x writes one pair at handle+8 for bit 40. A write of
// chunkWhole bytes or more is not one record. It is written as the
// run of chunks Chunks describes for its own length from its own
// address, each a record of its own, and a chunk that would cross an
// arena boundary is split there and goes on in the next arena at key
// 0. Found by t9_vec292 and t9_vec12000, the rule by the t10_vec sizes,
// and its use for a partial write by t32_wide_slice__ and
// t33_v_wsl_hi____.
//
// Records of one arena are in write order, and records of different
// arenas at the same time cannot be ordered against each other:
// t12_v_mem40_t0 writes forty elements at time zero across two arenas
// and the file keeps only which arena each went to. Changes orders the
// records by time, keeping file order within a time, and reads the
// writes of a time in that order. A record at the whole value's first
// chunk address with that chunk's length starts a whole write, whose
// other chunks are the first unused records of the time at the other
// chunk addresses, whichever arena holds them. Any other record starts
// a partial write: the longest chain of unused records of the time,
// each the first at the address where the previous one ends, whose
// addresses are the ones the rule predicts for the chain's length, or
// the record alone. The chain is found by address, not by position,
// because the rest of a chunk split at an arena boundary sits behind
// the other writes of the time in the next arena: t33_v_mem_row___
// writes four rows at time zero and arena 1 holds them last row
// first. Changes overlays every
// write onto the value and emits one value per write. It checks that
// the first write covers the object, that a lone record is shorter
// than a chunked write, and that a lone Verilog record is whole word
// pairs.
//
// The 8 byte rest of a chunk split at an arena boundary has the shape
// of a pair write, t12_v_mem40w32 at 0x800. When both sit at one
// time the whole write takes the first, and a pair write before the
// whole write in that arena is read as its rest; the final value of
// the time is right either way, because the arena keeps write order.
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
	// handle; Offset is nonzero for a port bound to a slice. A VHDL
	// offset counts bytes. A Verilog offset counts bits from the low
	// end of the actual, //hdl/serv:sim, so the object takes the word
	// pairs its bits fall in and shifts them down afterwards.
	verilog := f.verilog(dc.Type)
	start := o.Handle + uint64(o.Offset)
	shift, nbits := 0, 0
	if verilog && o.Offset != 0 {
		nbits, err = f.Size(dc)
		if err != nil {
			return nil, err
		}
		pa, pb := uint64(o.Offset)/32, (uint64(o.Offset)+uint64(nbits)-1)/32
		start = o.Handle + 8*pa
		size = 8 * (pb - pa + 1)
		shift = int(uint64(o.Offset) - 32*pa)
	}
	end := start + size
	first, last := int(start/arenaSpan), int((end-1)/arenaSpan)
	if last >= len(f.Arenas) {
		return nil, fmt.Errorf("object handle %#x with %d bytes reaches arena %d of %d", o.Handle, size, last, len(f.Arenas))
	}
	// The 16 byte record of an untyped time parameter runs into the
	// next object, so only the record at its own address is its own,
	// and only the first 8 bytes of it: t30_sv_ptm_two.
	timeParam := verilog && f.timeParam(dc)
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
				if timeParam {
					if addr != start || len(r.Data) < 8 {
						continue
					}
					recs = append(recs, rec{r.Time, addr, r.Data[:8]})
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
	// The records belong to the signal at the handle, so its size sets
	// the chunk boundaries, not the object's. They differ for a port
	// bound to a slice of a chunked signal: `bus_req_i` of
	// //hdl/neorv32:sim is 88 bytes at offset 1408 of a 1760 byte
	// array, and the array's initial write puts a chunk boundary
	// inside the port's 88 bytes.
	whole, wholeSize := start, size
	if o.Offset != 0 {
		for _, p := range f.Objects {
			if p.Handle != o.Handle || p.Offset != 0 {
				continue
			}
			n, err := f.recordBytes(f.Decls[p.Decl])
			if err != nil {
				continue
			}
			if uint64(n) > wholeSize {
				whole, wholeSize = p.Handle, uint64(n)
			}
		}
	}
	starts := chunkStarts(whole, wholeSize)
	pieceLen := func(i int) uint64 {
		if i+1 < len(starts) {
			return starts[i+1] - starts[i]
		}
		return whole + wholeSize - starts[i]
	}
	// A chunk of the signal that the object's own bytes do not reach
	// has no record here: the records were filtered to the object's
	// window.
	visible := func(i int) bool {
		return starts[i] < end && starts[i]+pieceLen(i) > start
	}
	cur := make([]byte, size)
	var out []Change
	apply := func(r rec) {
		lo, hi := r.addr, r.addr+uint64(len(r.data))
		if lo < start {
			lo = start
		}
		if hi > end {
			hi = end
		}
		copy(cur[lo-start:hi-start], r.data[lo-r.addr:hi-r.addr])
	}
	// write finds the records of the write group[i] starts, among the
	// unused records of the time, and returns their indexes.
	write := func(group []rec, used []bool, i int) ([]int, error) {
		r := group[i]
		first := -1
		for p := range starts {
			if visible(p) {
				first = p
				break
			}
		}
		if first >= 0 && r.addr == starts[first] && uint64(len(r.data)) == pieceLen(first) {
			idx := []int{i}
			for p := first + 1; p < len(starts); p++ {
				if !visible(p) {
					continue
				}
				found := -1
				for j := i + 1; j < len(group); j++ {
					if !used[j] && group[j].addr == starts[p] && uint64(len(group[j].data)) == pieceLen(p) {
						found = j
						break
					}
				}
				if found < 0 {
					idx = nil
					break
				}
				idx = append(idx, found)
			}
			if idx != nil {
				return idx, nil
			}
		}
		run := []int{i}
		total := uint64(len(r.data))
		for {
			found := -1
			for j := i + 1; j < len(group); j++ {
				if !used[j] && group[j].addr == r.addr+total {
					found = j
					break
				}
			}
			if found < 0 {
				break
			}
			run = append(run, found)
			total += uint64(len(group[found].data))
		}
		for ; len(run) > 1; run = run[:len(run)-1] {
			want := chunkStarts(r.addr, total)
			ok := len(want) == len(run)
			for k := 0; ok && k < len(want); k++ {
				ok = want[k] == group[run[k]].addr
			}
			if ok {
				return run, nil
			}
			total -= uint64(len(group[run[len(run)-1]].data))
		}
		if total >= chunkWhole {
			return nil, fmt.Errorf("object handle %#x: a record of %d bytes at %#x, time %d, is not a run of chunks", o.Handle, total, r.addr, r.time)
		}
		if verilog && !timeParam && ((r.addr-start)%8 != 0 || total%8 != 0) {
			return nil, fmt.Errorf("object handle %#x: record at %#x of %d bytes is neither a chunk nor whole word pairs", o.Handle, r.addr, total)
		}
		return run, nil
	}
	for lo := 0; lo < len(recs); {
		hi := lo
		for hi < len(recs) && recs[hi].time == recs[lo].time {
			hi++
		}
		group := recs[lo:hi]
		used := make([]bool, len(group))
		for i := range group {
			if used[i] {
				continue
			}
			idx, err := write(group, used, i)
			if err != nil {
				return nil, err
			}
			// The first write is the initial value of the whole signal;
			// a port bound to a slice at its start, t9_port_sliceto_,
			// sees a first write longer than its own value.
			if len(out) == 0 {
				lo, hi := group[idx[0]].addr, group[idx[0]].addr
				for _, j := range idx {
					if group[j].addr < lo {
						lo = group[j].addr
					}
					if e := group[j].addr + uint64(len(group[j].data)); e > hi {
						hi = e
					}
				}
				if lo > start || hi < end {
					return nil, fmt.Errorf("object handle %#x with %d bytes has a first write of %d bytes at %#x, which does not cover it", o.Handle, size, hi-lo, lo)
				}
			}
			for _, j := range idx {
				used[j] = true
				apply(group[j])
			}
			data := append([]byte(nil), cur...)
			if shift != 0 {
				data = shiftPairs(data, shift, nbits)
			}
			out = append(out, Change{Time: group[i].time, Data: data})
		}
		lo = hi
	}
	return out, nil
}

// shiftPairs takes nbits bits of the Verilog word pairs in data, from
// bit shift up, and returns them as word pairs of their own, bit 0 at
// bit 0: the value of a port bound to a slice of its actual.
func shiftPairs(data []byte, shift, nbits int) []byte {
	out := make([]byte, 8*((nbits+31)/32))
	for i := 0; i < nbits; i++ {
		s, d := i+shift, i
		a := binary.LittleEndian.Uint32(data[8*(s/32):])
		b := binary.LittleEndian.Uint32(data[8*(s/32)+4:])
		oa := binary.LittleEndian.Uint32(out[8*(d/32):])
		ob := binary.LittleEndian.Uint32(out[8*(d/32)+4:])
		oa |= (a >> (s % 32) & 1) << (d % 32)
		ob |= (b >> (s % 32) & 1) << (d % 32)
		binary.LittleEndian.PutUint32(out[8*(d/32):], oa)
		binary.LittleEndian.PutUint32(out[8*(d/32)+4:], ob)
	}
	return out
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
// constants are 24 and 299 is not known. The remainder can be longer
// than a chunk by up to count - 1 bytes; chunkLens says what happens
// to it then.
func Chunks(size uint64) (count, length uint64) {
	if size < chunkWhole {
		return 1, size
	}
	s := size + 24
	count = 2 * ((s + 298) / 299)
	return count, size / count
}

// chunkLens lists the record lengths of a value of size bytes, in
// order: the chunks of Chunks, with the remainder chunked again by the
// same rule when it is 276 bytes or more. A remainder of exactly 275
// bytes is one record, where a value of 275 bytes is two chunks. Found
// by //hdl/potato:sim, whose 32768 byte memories end in four records
// of 89 where one of 356 was expected; the threshold by t39_vec20120
// against t39_vec20121, remainders of 275 and 276, and the rule by
// the other tier 39 remainders from 280 to 356.
func chunkLens(size uint64) []uint64 {
	count, length := Chunks(size)
	if count == 1 {
		return []uint64{size}
	}
	var lens []uint64
	for i := uint64(1); i < count; i++ {
		lens = append(lens, length)
	}
	rest := size - (count-1)*length
	if rest > chunkWhole {
		return append(lens, chunkLens(rest)...)
	}
	return append(lens, rest)
}

// chunkStarts lists the record addresses chunkLens predicts for a
// value of size bytes at handle, allowing for the split of a chunk at
// an arena boundary: the chunk goes on in the next arena as a record
// of its own at key 0.
func chunkStarts(handle, size uint64) []uint64 {
	var want []uint64
	at := handle
	for _, l := range chunkLens(size) {
		want = append(want, at)
		// An arena boundary inside the chunk starts a record.
		if b := (at/arenaSpan + 1) * arenaSpan; b < at+l {
			want = append(want, b)
		}
		at += l
	}
	return want
}
