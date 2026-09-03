// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"encoding/binary"
	"fmt"
	"math"
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
		// The fields of a packed union overlap, so it takes the
		// bits of its widest field: t24_sv_union____.
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
			if ty.Layout == LayoutUnion {
				total = max(total, fb)
				continue
			}
			total += fb
		}
		return total, nil
	}
	return 0, fmt.Errorf("type %q has kind %s", ty.Name, ty.Kind)
}

// ValueClass is the value class code of the declaration's objects:
// the first word of the Debug.Classes entry Decl.Class indexes, or 0
// for a file without objects. What the codes stand for is open; see
// docs/format/hierarchy.md.
func (f *File) ValueClass(dc Decl) uint32 {
	if dc.Class < len(f.Debug.Classes) {
		return f.Debug.Classes[dc.Class][0]
	}
	return 0
}

// timeParam reports whether the declaration is an untyped parameter
// holding a time literal: parameter T = 10ns. Its declaration is an
// unnamed 64 bit vector of value class 4, where every other unnamed
// vector parameter is class 1 or 3, and its storage is the float64 of
// the value in the time unit, 8 bytes, with the record written 16
// bytes long as a 64 bit vector's would be: the second 8 bytes are
// whatever follows, the next parameter's value in t30_sv_ptm_two.
// Tier 28, t28_sv_prm_time against t28_sv_prm_tmtyp; tier 30.
func (f *File) timeParam(dc Decl) bool {
	return dc.Kind == DeclParam && f.verilog(dc.Type) && f.Types[dc.Type].Name == "" && dc.Size == 64 && f.ValueClass(dc) == 4
}

// recordBytes is the number of bytes the value of the declaration takes
// in a page record: its byte size for a VHDL object, 8 bytes per
// 32 bits of its declared size for a Verilog one, and 8 for the
// float64 of an untyped time parameter.
func (f *File) recordBytes(dc Decl) (int, error) {
	if f.timeParam(dc) {
		return 8, nil
	}
	n, err := f.Size(dc)
	if err != nil {
		return 0, err
	}
	if f.verilog(dc.Type) {
		return 8 * ((n + 31) / 32), nil
	}
	return n, nil
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

// slotted reports whether a Verilog type instance takes whole word
// pairs of its own inside an unpacked array: a real, or an unpacked
// struct. Anything else packs into the array's bit string.
func (f *File) slotted(t int) bool {
	ty := &f.Types[f.resolve(t)]
	return ty.Kind == KindReal || ty.Kind == KindRecord && ty.Layout == LayoutUnpacked
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
	case ty.Kind == KindArray && ty.Layout == LayoutUnpacked && f.slotted(ty.Elem):
		// An unpacked array of real gives each element one pair, the
		// last element lowest, as an unpacked struct does its fields:
		// t13_sv_real_arr writes r[1] of real r [0:1] into pair 0.
		// An unpacked array of unpacked structs does the same with a
		// slot of one pair per rounded field: t35_sv_ust_arr__ writes
		// m[1] of rec_t m [0:1] into pairs 0 and 1 of 4, b in pair 0
		// and a in pair 1.
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
			return v, fmt.Errorf("%s: %w", ty.Name, err)
		}
		slot := 8 * ((eb + 31) / 32)
		if len(data) != slot*n {
			return v, fmt.Errorf("%s: %d bytes for %d slots of %d", ty.Name, len(data), n, slot)
		}
		v.Elems = make([]Value, n)
		for i := 0; i < n; i++ {
			ers := *rs
			e, err := f.decodeVerilog(ty.Elem, data[slot*(n-1-i):slot*(n-i)], &ers)
			if err != nil {
				return v, err
			}
			v.Elems[i] = e
		}
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
		switch ty.Layout {
		case LayoutPacked:
		case LayoutUnion:
			// Every field of a packed union reads the same bits,
			// a narrower field the low ones: t24_sv_union____.
			for _, fd := range ty.Fields {
				frs := fd.Ranges
				fb, err := f.bitsOf(fd.Type, &frs)
				if err != nil {
					return v, fmt.Errorf("%s.%s: %w", ty.Name, fd.Name, err)
				}
				if fb > len(bits) {
					return v, fmt.Errorf("%s.%s: needs %d of %d bits", ty.Name, fd.Name, fb, len(bits))
				}
				frs = fd.Ranges
				fv, err := f.decodeBits(fd.Type, bits[len(bits)-fb:], &frs)
				if err != nil {
					return v, fmt.Errorf("%s.%s: %w", ty.Name, fd.Name, err)
				}
				v.Fields = append(v.Fields, fv)
			}
			return v, nil
		default:
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
