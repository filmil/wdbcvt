// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// This file holds the Verilog and SystemVerilog side of the value model.
// It was worked out on tier 11 of the corpus; docs/format/values.md
// states each rule with the case that found it.
//
// A Verilog object is measured in bits, and its page record holds one
// pair of 32 bit words per 32 bits of the value: the low word of each
// pair carries the a bits and the high word the b bits, and bit i of
// the value is a + 2b at bit i%32 of pair i/32, which indexes the
// literal list 0 1 Z X of the logic type. Pair 0 holds bits 31:0.

// resolve follows typedef aliases to the entry they name.
func (f *File) resolve(t int) int {
	return f.resolveRanges(t, nil)
}

// resolveRanges follows typedef aliases to the entry they name. When an
// alias carries a constraint of its own, that constraint replaces the
// range list the target reads: t12_sv_typedef declares s as byte_t, an
// alias of the unnamed vector entry with (7, 0, downto), and the
// declaration of s carries no range of its own. The caller's list is
// left alone; the alias's copy is returned through rs.
func (f *File) resolveRanges(t int, rs **[]Range) int {
	for i := 0; i < len(f.Types); i++ {
		if t < 0 || t >= len(f.Types) || f.Types[t].Kind != KindAlias {
			return t
		}
		if rs != nil && len(f.Types[t].Ranges) > 0 {
			own := append([]Range(nil), f.Types[t].Ranges...)
			*rs = &own
		}
		t = f.Types[t].Target
	}
	return t
}

// verilog reports whether the type came from a Verilog source: bit 0 of
// its origin word is set. A VHDL type has bit 1 set instead.
func (f *File) verilog(t int) bool {
	t = f.resolve(t)
	if t < 0 || t >= len(f.Types) {
		return false
	}
	return f.Types[t].Origin&OriginVerilog != 0
}

// arrayDims returns the index ranges of one array type instance. A
// named type with its own constraint, `int` or `time`, carries it in
// the entry; an unnamed vector or memory carries (0, 0, -2) per
// dimension and takes the constraints from the declaration, or from
// the record field or values entry that names it.
func (f *File) arrayDims(ty *Type, rs *[]Range) ([]Range, error) {
	if ty.Dims <= len(ty.Ranges) && ty.Dims > 0 && ty.Ranges[0].Dir != -2 {
		return ty.Ranges[:ty.Dims], nil
	}
	return takeRanges(rs, ty.Dims, ty.Name)
}

// bitsOf is the declared size in bits of a type instance, the number
// Decl.Size holds for a Verilog object. A real counts 32 although its
// record slot holds a 64 bit float: t11_v_real declares 32. A packed
// array or struct is the sum of its parts, and so is an unpacked
// array: t11_v_mem4w5 packs four 5 bit words into 20 bits. An unpacked
// struct rounds every field up to a whole 32 bit word: t11_sv_struct3
// declares 96 for 1 + 4 + 8 bits, t11_sv_struct40 declares 96 for
// 40 + 1.
func (f *File) bitsOf(t int, rs *[]Range) (int, error) {
	t = f.resolveRanges(t, &rs)
	if t < 0 || t >= len(f.Types) {
		return 0, fmt.Errorf("type index %d of %d", t, len(f.Types))
	}
	ty := &f.Types[t]
	switch ty.Kind {
	case KindEnum:
		return 1, nil
	case KindValues:
		brs := ty.Ranges
		return f.bitsOf(ty.Elem, &brs)
	case KindReal:
		return 32, nil
	case KindArray:
		dims, err := f.arrayDims(ty, rs)
		if err != nil {
			return 0, err
		}
		n := 1
		for _, r := range dims {
			n *= r.Length()
		}
		eb, err := f.bitsOf(ty.Elem, rs)
		if err != nil {
			return 0, err
		}
		return n * eb, nil
	case KindRecord:
		total := 0
		for _, fd := range ty.Fields {
			frs := fd.Ranges
			fb, err := f.bitsOf(fd.Type, &frs)
			if err != nil {
				return 0, fmt.Errorf("%s.%s: %w", ty.Name, fd.Name, err)
			}
			if ty.Layout == LayoutUnpacked {
				fb = roundUp(fb, 32)
			}
			total += fb
		}
		return total, nil
	}
	return 0, fmt.Errorf("type %q has kind %s", ty.Name, ty.Kind)
}

// recordBytes is the number of bytes the value of the declaration takes
// in a page record: its byte size for a VHDL object, and 8 bytes per
// 32 bits of its declared size for a Verilog one.
func (f *File) recordBytes(dc Decl) (int, error) {
	n, err := f.Size(dc)
	if err != nil {
		return 0, err
	}
	if f.verilog(dc.Type) {
		return 8 * ((n + 31) / 32), nil
	}
	return n, nil
}

// changesVerilog lists the changes of a Verilog object. A record holds
// the whole value, or one or more of its 32 bit word pairs keyed at the
// handle plus eight times the index of the first pair, and the pairs it
// does not hold keep their last value: t11_v_vec64x writes 8 bytes at
// handle+8 when bit 40 alone changes, and the memories of tier 11 write
// the pair that holds the written element. A reg initialised in a
// declaration gets an all X record and then its value, both at time
// zero: t11_v_bit_edge. A memory written element by element in an
// initial block gets one record per element, all at time zero:
// t11_v_mem8 has eight.
//
// A whole value of chunkWhole record bytes or more is written as the
// chunks Chunks predicts, split at an arena boundary like a VHDL value,
// and the split falls inside a word pair: t12_v_vec1089 writes 280
// bytes as four records of 70. A pair write into such an object is
// still one record at its pair address: t12_v_vec4800x writes 8 bytes
// at handle+600 for bit 2400. So every record at one time is either a
// piece of a whole write, one of the addresses and lengths of
// chunkStarts, or a pair write, and the pieces of one whole write are
// joined by taking the i-th record at every piece address, the way
// Changes does for VHDL.
//
// Records of one arena are in write order, and records of different
// arenas at the same time cannot be ordered against each other. The
// records of the arena holding the handle set the order: a whole write
// sits where its first piece sits, and the pair writes of the other
// arenas follow them at that time, arena by arena. t12_v_mem40_t0
// writes forty elements at time zero across two arenas and the file
// keeps only which arena each went to.
func (f *File) changesVerilog(o Object, start, end uint64) ([]Change, error) {
	type rec struct {
		time, addr uint64
		data       []byte
		arena      int
		piece      int // index into starts, or -1 for a pair write
	}
	size := end - start
	starts := chunkStarts(start, size)
	pieceAt := map[uint64]int{}
	for i, a := range starts {
		pieceAt[a] = i
	}
	pieceLen := func(i int) uint64 {
		if i+1 < len(starts) {
			return starts[i+1] - starts[i]
		}
		return end - starts[i]
	}
	var recs []rec
	first, last := int(start/arenaSpan), int((end-1)/arenaSpan)
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
				piece := -1
				if i, ok := pieceAt[addr]; ok && uint64(len(r.Data)) == pieceLen(i) {
					piece = i
				} else if (addr-start)%8 != 0 || len(r.Data)%8 != 0 {
					return nil, fmt.Errorf("object handle %#x: record at %#x of %d bytes is neither a chunk nor whole word pairs", o.Handle, addr, len(r.Data))
				}
				recs = append(recs, rec{r.TimePS, addr, r.Data, k, piece})
			}
		}
	}
	if len(recs) == 0 {
		return nil, fmt.Errorf("object handle %#x is logged but has no records", o.Handle)
	}
	sort.SliceStable(recs, func(i, j int) bool { return recs[i].time < recs[j].time })
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
	for lo := 0; lo < len(recs); {
		hi := lo
		for hi < len(recs) && recs[hi].time == recs[lo].time {
			hi++
		}
		group := recs[lo:hi]
		// The i-th record at every piece address forms whole write i.
		// A pair write can have the address and length of a piece: the
		// 8 byte rest of a chunk split at an arena boundary looks like
		// the element write t12_v_mem40_t0 makes at that address. The
		// piece addresses with the fewest records set the count of
		// whole writes, and the surplus records elsewhere, taken from
		// the end, are pair writes.
		pieces := make([][]rec, len(starts))
		for _, r := range group {
			if r.piece >= 0 {
				pieces[r.piece] = append(pieces[r.piece], r)
			}
		}
		count := len(pieces[0])
		for _, p := range pieces {
			if len(p) < count {
				count = len(p)
			}
		}
		for i := range pieces {
			if len(pieces[i]) == count {
				continue
			}
			if l := pieceLen(i); l%8 != 0 || (starts[i]-start)%8 != 0 {
				return nil, fmt.Errorf("object handle %#x at %d ps: chunk at %#x has %d records, another has %d", o.Handle, recs[lo].time, starts[i], len(pieces[i]), count)
			}
			for _, r := range pieces[i][count:] {
				for j := range group {
					if group[j].addr == r.addr && group[j].arena == r.arena && group[j].piece >= 0 && &group[j].data[0] == &r.data[0] {
						group[j].piece = -1
					}
				}
			}
			pieces[i] = pieces[i][:count]
		}
		whole := 0
		emit := func(r rec) {
			if r.piece < 0 {
				apply(r)
			} else {
				for _, p := range pieces {
					apply(p[whole])
				}
				whole++
			}
			out = append(out, Change{TimePS: r.time, Data: append([]byte(nil), cur...)})
		}
		for _, r := range group {
			if r.arena == first && (r.piece <= 0) {
				emit(r)
			}
		}
		for _, r := range group {
			if r.arena != first && r.piece < 0 {
				emit(r)
			}
		}
		lo = hi
	}
	return out, nil
}

// unpackBits turns the word pairs of a record into one value per bit,
// most significant bit first: 0, 1, 2 for Z and 3 for X.
func unpackBits(data []byte, nbits int) ([]byte, error) {
	if need := 8 * ((nbits + 31) / 32); len(data) < need {
		return nil, fmt.Errorf("%d bits need %d bytes, have %d", nbits, need, len(data))
	}
	bits := make([]byte, nbits)
	for i := 0; i < nbits; i++ {
		w := i / 32
		a := binary.LittleEndian.Uint32(data[8*w:])
		b := binary.LittleEndian.Uint32(data[8*w+4:])
		bits[nbits-1-i] = byte(a>>(i%32)&1) | byte(b>>(i%32)&1)<<1
	}
	return bits, nil
}

// decodeVerilog decodes the record bytes of one Verilog type instance.
// An unpacked struct is decoded by slots: the last field sits in the
// lowest word pairs and the first field in the highest, each field in
// as many pairs as its bits need, laid out as a value of its own would
// be. t11_sv_struct3 puts c in pair 0, b in pair 1 and a in pair 2;
// t11_sv_struct40 puts b in pair 0 and the 40 bit a in pairs 1 and 2.
// A real takes one pair holding the float: t11_sv_struct_r. Everything
// else is a string of bits decoded by decodeBits.
func (f *File) decodeVerilog(t int, data []byte, rs *[]Range) (Value, error) {
	rt := f.resolveRanges(t, &rs)
	ty := &f.Types[rt]
	v := Value{Type: t}
	switch {
	case ty.Kind == KindReal:
		if len(data) < 8 {
			return v, fmt.Errorf("%s: need 8 bytes, have %d", ty.Name, len(data))
		}
		v.Scalar = strconv.FormatFloat(math.Float64frombits(binary.LittleEndian.Uint64(data)), 'g', -1, 64)
		return v, nil
	case ty.Kind == KindRecord && ty.Layout == LayoutUnpacked:
		v.Fields = make([]Value, len(ty.Fields))
		w := 0
		for i := len(ty.Fields) - 1; i >= 0; i-- {
			fd := ty.Fields[i]
			frs := fd.Ranges
			fb, err := f.bitsOf(fd.Type, &frs)
			if err != nil {
				return v, fmt.Errorf("%s.%s: %w", ty.Name, fd.Name, err)
			}
			nw := (fb + 31) / 32
			if len(data) < 8*(w+nw) {
				return v, fmt.Errorf("%s.%s: need %d bytes, have %d", ty.Name, fd.Name, 8*(w+nw), len(data))
			}
			frs = fd.Ranges
			fv, err := f.decodeVerilog(fd.Type, data[8*w:8*(w+nw)], &frs)
			if err != nil {
				return v, fmt.Errorf("%s.%s: %w", ty.Name, fd.Name, err)
			}
			v.Fields[i] = fv
			w += nw
		}
		return v, nil
	}
	brs := *rs
	nb, err := f.bitsOf(t, &brs)
	if err != nil {
		return v, err
	}
	bits, err := unpackBits(data, nb)
	if err != nil {
		return v, fmt.Errorf("%s: %w", ty.Name, err)
	}
	return f.decodeBits(t, bits, rs)
}

// decodeBits decodes a value from its bits, most significant first. A
// packed array or struct concatenates its parts with the leftmost
// element, or the first field, at the top; an unpacked array does the
// same at bit granularity: t11_v_vec8_asc, t11_sv_struct, t11_v_mem4,
// t11_v_mem4_desc, t11_v_mem4w5.
func (f *File) decodeBits(t int, bits []byte, rs *[]Range) (Value, error) {
	rt := f.resolveRanges(t, &rs)
	ty := &f.Types[rt]
	v := Value{Type: t}
	switch ty.Kind {
	case KindEnum:
		if len(bits) != 1 {
			return v, fmt.Errorf("%s: %d bits for one literal", ty.Name, len(bits))
		}
		if int(bits[0]) >= len(ty.Literals) {
			return v, fmt.Errorf("%s: literal index %d of %d", ty.Name, bits[0], len(ty.Literals))
		}
		v.Scalar = strings.Trim(ty.Literals[bits[0]], "'")
		return v, nil
	case KindValues:
		brs := ty.Ranges
		bv, err := f.decodeBits(ty.Elem, bits, &brs)
		if err != nil {
			return v, fmt.Errorf("%s: %w", ty.Name, err)
		}
		if n, ok := bitsUint(bits); ok {
			for _, nv := range ty.Values {
				if nv.Value == n {
					v.Scalar = nv.Name
					return v, nil
				}
			}
		}
		v.Scalar = f.String(bv)
		return v, nil
	case KindArray:
		dims, err := f.arrayDims(ty, rs)
		if err != nil {
			return v, err
		}
		n := 1
		for _, r := range dims {
			n *= r.Length()
		}
		ers := *rs
		eb, err := f.bitsOf(ty.Elem, &ers)
		if err != nil {
			return v, err
		}
		if len(bits) != n*eb {
			return v, fmt.Errorf("%s: %d bits for %d elements of %d", ty.Name, len(bits), n, eb)
		}
		for i := 0; i < n; i++ {
			ers = *rs
			e, err := f.decodeBits(ty.Elem, bits[i*eb:(i+1)*eb], &ers)
			if err != nil {
				return v, err
			}
			v.Elems = append(v.Elems, e)
		}
		*rs = ers
		if s, ok := f.integral(ty, bits); ok {
			v.Scalar = s
		}
		return v, nil
	case KindRecord:
		if ty.Layout != LayoutPacked {
			return v, fmt.Errorf("%s: an unpacked struct inside a packed value has not been observed", ty.Name)
		}
		off := 0
		for _, fd := range ty.Fields {
			frs := fd.Ranges
			fb, err := f.bitsOf(fd.Type, &frs)
			if err != nil {
				return v, fmt.Errorf("%s.%s: %w", ty.Name, fd.Name, err)
			}
			if off+fb > len(bits) {
				return v, fmt.Errorf("%s.%s: needs bits %d to %d of %d", ty.Name, fd.Name, off, off+fb, len(bits))
			}
			frs = fd.Ranges
			fv, err := f.decodeBits(fd.Type, bits[off:off+fb], &frs)
			if err != nil {
				return v, fmt.Errorf("%s.%s: %w", ty.Name, fd.Name, err)
			}
			v.Fields = append(v.Fields, fv)
			off += fb
		}
		if off != len(bits) {
			return v, fmt.Errorf("%s: fields take %d of %d bits", ty.Name, off, len(bits))
		}
		return v, nil
	}
	return v, fmt.Errorf("type %q has kind %s", ty.Name, ty.Kind)
}

// bitsUint is the unsigned value of a bit string with no X or Z, up to
// 64 bits wide.
func bitsUint(bits []byte) (uint64, bool) {
	if len(bits) > 64 {
		return 0, false
	}
	var n uint64
	for _, b := range bits {
		if b > 1 {
			return 0, false
		}
		n = n<<1 | uint64(b)
	}
	return n, true
}

// integral spells a value of a predefined integral type in decimal, the
// way the truth files and xsim print it: signed for integer, int,
// shortint, byte and longint, unsigned for time. A value with an X or
// Z bit, or of any other type, keeps its bit string.
func (f *File) integral(ty *Type, bits []byte) (string, bool) {
	if ty.Origin != OriginVerilogPre && ty.Origin != OriginVerilogTime {
		return "", false
	}
	n, ok := bitsUint(bits)
	if !ok {
		return "", false
	}
	switch ty.Name {
	case "integer", "int", "shortint", "byte", "longint":
		w := uint(len(bits))
		if w < 64 && n&(1<<(w-1)) != 0 {
			return strconv.FormatInt(int64(n)-int64(1)<<w, 10), true
		}
		return strconv.FormatInt(int64(n), 10), true
	case "time":
		return strconv.FormatUint(n, 10), true
	}
	return "", false
}
