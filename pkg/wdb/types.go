// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Kind is the low byte of a type entry's tag word. The values were read
// off the corpus: each corpus case declares one known VHDL type, and the
// tag of the entry with that type's name is its kind.
type Kind uint8

// The kinds observed so far. Every corpus type falls into one of these.
const (
	KindEnum     Kind = 0x03 // enumeration: BIT, STD_ULOGIC, BOOLEAN, CHARACTER, user enums, Verilog logic and bit
	KindValues   Kind = 0x04 // SystemVerilog enum: named values of a base type
	KindInteger  Kind = 0x05 // integer with bounds
	KindReal     Kind = 0x06 // floating point with bounds
	KindAlias    Kind = 0x07 // SystemVerilog typedef: a name for another entry
	KindAccess   Kind = 0x08 // VHDL access type: a designated type and two words
	KindFile     Kind = 0x0c // VHDL file type: an element type and two words
	KindPhysical Kind = 0x0d // physical type with units: TIME
	KindArray    Kind = 0x10 // array, constrained or not
	KindRecord   Kind = 0x11 // record with named fields
)

// The first word of most entries says which language the type came
// from, and is a bit set: bit 1 for VHDL, bit 0 for Verilog, bit 2 for
// a Verilog predefined type, and bit 3 for a time type in either
// language. Found by t11_v_bit_edge against t1_bit_one_edge, then the
// rest of tier 11 against tier 2; see docs/format/types.md.
const (
	OriginVHDL        = 0x2
	OriginVHDLTime    = 0xa
	OriginVerilog     = 0x1
	OriginVerilogPre  = 0x5
	OriginVerilogTime = 0xd
)

// The u16 after the first word of an array or record entry: 1 for a
// VHDL type, 3 for a packed Verilog type and 2 for an unpacked one.
// Found by t11_sv_struct against t11_sv_ustruct, and by t11_v_mem4,
// whose memory is unpacked and whose word is packed.
const (
	LayoutVHDL     = 1
	LayoutUnpacked = 2
	LayoutPacked   = 3
)

func (k Kind) String() string {
	switch k {
	case KindEnum:
		return "enum"
	case KindValues:
		return "values"
	case KindInteger:
		return "integer"
	case KindReal:
		return "real"
	case KindAlias:
		return "alias"
	case KindAccess:
		return "access"
	case KindFile:
		return "file"
	case KindPhysical:
		return "physical"
	case KindArray:
		return "array"
	case KindRecord:
		return "record"
	}
	return fmt.Sprintf("kind(%#x)", uint8(k))
}

// Range is one index constraint: Left, Right and a direction word, which
// is 1 for `to`, -1 for `downto` and -2 for an unconstrained dimension.
type Range struct {
	Left, Right, Dir int32
}

// Length is the number of elements the range spans.
func (r Range) Length() int {
	if r.Dir < 0 {
		return int(r.Left - r.Right + 1)
	}
	return int(r.Right - r.Left + 1)
}

func (r Range) String() string {
	switch r.Dir {
	case 1:
		return fmt.Sprintf("%d to %d", r.Left, r.Right)
	case -1:
		return fmt.Sprintf("%d downto %d", r.Left, r.Right)
	}
	return fmt.Sprintf("(%d, %d, dir %d)", r.Left, r.Right, r.Dir)
}

// TimeUnit is one unit of a physical type: its name and its size in
// picoseconds. TIME lists fs as 0, ps as 1, ns as 1000 and so on up to
// hr, and a TIME value in a page is a signed 64 bit picosecond count.
type TimeUnit struct {
	Name  string
	Scale uint64
}

// NamedValue is one value of a SystemVerilog enum.
type NamedValue struct {
	Name  string
	Value uint64
}

// Field is one field of a record type.
type Field struct {
	Name string
	// Type indexes File.Types.
	Type int
	// Ranges are the constraint triples that follow the type index.
	// Their meaning is still open; see docs/format.md.
	Ranges []Range
}

// Type is one entry of the type table.
type Type struct {
	Kind Kind
	// Name is the type name as the table stores it. VHDL predefined and
	// IEEE types are upper case, user types keep the source case, and a
	// Verilog vector, memory or struct with no typedef has no name.
	Name string
	// Origin is the first word of the entry, one of the Origin values.
	// A physical type and an alias do not carry it.
	Origin uint32
	// Layout is the u16 after Origin in an array or record entry, one
	// of the Layout values.
	Layout uint16

	// Variant is the second word of an enumeration or real entry. It is
	// 2 for a VHDL enumeration, 0 for VHDL REAL, 0 for Verilog logic and
	// 1 for Verilog bit and real. What it selects is open.
	Variant uint32
	// Class is the third word of an enumeration entry. It follows the
	// literals, not the type name: 2 for '0' '1', 3 for the nine
	// STD_ULOGIC literals, 4 for any other set with a character literal
	// in it, 5 for identifiers only, tier 20. It is 1 for Verilog logic
	// and bit.
	Class uint32
	// Trailer is the last word of an enumeration or real entry. For a
	// VHDL enumeration it is the byte size of a value: 1 up to 256
	// literals and 4 beyond, t20_enum_300. It is 1 for VHDL REAL and 0
	// for the Verilog entries. See docs/format/types.md.
	Trailer uint32
	// Literals are the enumeration literals in declaration order.
	// Character literals keep their quotes, as in `'U'`. Verilog logic
	// lists 0 1 Z X and bit lists 0 1 0 0.
	Literals []string
	// Values lists the named values of a SystemVerilog enum, whose base
	// type is Elem, constrained by Ranges. Found by t11_sv_enum.
	Values []NamedValue
	// Target is the entry an alias names. Found by t11_sv_enum, whose
	// typedef state_t is an alias of the unnamed values entry. An alias
	// of an unnamed vector carries the vector's constraint in Ranges:
	// t12_sv_typedef, whose byte_t lists (7, 0, downto).
	Target int

	// Low and High bound an integer type.
	Low, High int32
	// FLow and FHigh bound a real type.
	FLow, FHigh float64

	// Units lists a physical type's units.
	Units []TimeUnit

	// Words holds the two words after the element type of a file or
	// access type. Every file type holds 8 and 40, every access type
	// 8 and 48, and an access variable declares 48 bytes while a file
	// variable declares 0. What the words mean is open.
	Words []uint32

	// Elem is the type index of an array's element type. Dims is its
	// dimension count and Indexes holds one index type per dimension:
	// t18_arr_2dim has 2 dims and (1, 1) for a (0 to 1, 0 to 2) array,
	// t2_array2d has 1 dim and (3) for an array of vectors. Index is
	// Indexes[0]. Ranges holds one Range per dimension, then the element
	// constraints in the order they appear; an unconstrained dimension is
	// (0, 0, -2). The word before the ranges is their count: t5_int_arr
	// has 1 for (0 to 3), t2_array2d has 2 for (0 to 3) (7 downto 0),
	// t11_v_mem4 has 2 for two unconstrained triples.
	Elem, Index int
	Dims        int
	Indexes     []int
	Ranges      []Range

	// Fields lists a record's fields in declaration order.
	Fields []Field
}

const rangeEnd = -99 // terminates a run of constraint triples

// readTypes decodes the type table from the Xilinx RTTI section.
func readTypes(d []byte, rtti DirEntry) ([]Type, error) {
	sec := d[rtti.Offset : rtti.Offset+rtti.Length]
	if cstring(sec) != typeMagic {
		return nil, fmt.Errorf("no %q magic at %#x", typeMagic, rtti.Offset)
	}
	if len(sec) < 40 {
		return nil, fmt.Errorf("type table is %d bytes, shorter than its header", len(sec))
	}
	n := int(binary.LittleEndian.Uint32(sec[32:]))
	end := int(binary.LittleEndian.Uint32(sec[36:]))
	if end > len(sec) {
		return nil, fmt.Errorf("type table entries end at %#x, past the section", end)
	}
	types := make([]Type, 0, n)
	p := 40
	for i := 0; i < n; i++ {
		if p+8 > end {
			return nil, fmt.Errorf("type %d starts at %#x, past the entries", i, p)
		}
		elen := int(binary.LittleEndian.Uint32(sec[p:]))
		tag := binary.LittleEndian.Uint32(sec[p+4:])
		if elen < 8 || p+elen > end {
			return nil, fmt.Errorf("type %d has length %d at %#x", i, elen, p)
		}
		t, err := readType(Kind(tag&0xff), sec[p+8:p+elen])
		if err != nil {
			return nil, fmt.Errorf("type %d: %w", i, err)
		}
		types = append(types, t)
		p += elen
	}
	// A table of uint64 entry offsets follows the entries. Check it,
	// because it is the one internal cross reference the section offers.
	if end+8*n > len(sec) {
		return nil, fmt.Errorf("type offset table runs past the section")
	}
	q := 40
	for i, t := range types {
		got := binary.LittleEndian.Uint64(sec[end+8*i:])
		if got != uint64(q) {
			return nil, fmt.Errorf("type %d (%s) is at %#x but the offset table says %#x", i, t.Name, q, got)
		}
		q += int(binary.LittleEndian.Uint32(sec[q:]))
	}
	return types, nil
}

// cursor walks a little endian byte slice.
type cursor struct {
	b   []byte
	p   int
	err error
}

func (c *cursor) need(n int) bool {
	if c.err != nil {
		return false
	}
	if c.p+n > len(c.b) {
		c.err = fmt.Errorf("need %d bytes at %#x, have %d", n, c.p, len(c.b)-c.p)
		return false
	}
	return true
}

func (c *cursor) u16() uint16 {
	if !c.need(2) {
		return 0
	}
	v := binary.LittleEndian.Uint16(c.b[c.p:])
	c.p += 2
	return v
}

func (c *cursor) u32() uint32 {
	if !c.need(4) {
		return 0
	}
	v := binary.LittleEndian.Uint32(c.b[c.p:])
	c.p += 4
	return v
}

func (c *cursor) i32() int32 { return int32(c.u32()) }

func (c *cursor) u64() uint64 {
	if !c.need(8) {
		return 0
	}
	v := binary.LittleEndian.Uint64(c.b[c.p:])
	c.p += 8
	return v
}

func (c *cursor) f64() float64 { return math.Float64frombits(c.u64()) }

func (c *cursor) str() string {
	if c.err != nil {
		return ""
	}
	for i := c.p; i < len(c.b); i++ {
		if c.b[i] == 0 {
			s := string(c.b[c.p:i])
			c.p = i + 1
			return s
		}
	}
	c.err = fmt.Errorf("unterminated string at %#x", c.p)
	return ""
}

// ranges reads constraint triples up to the rangeEnd marker.
func (c *cursor) ranges() []Range {
	var out []Range
	for c.err == nil {
		left := c.i32()
		if left == rangeEnd || c.err != nil {
			return out
		}
		out = append(out, Range{Left: left, Right: c.i32(), Dir: c.i32()})
	}
	return out
}

func (c *cursor) expect(want uint32, what string) {
	if got := c.u32(); c.err == nil && got != want {
		c.err = fmt.Errorf("%s: got %#x, want %#x", what, got, want)
	}
}

func (c *cursor) expect16(want uint16, what string) {
	if got := c.u16(); c.err == nil && got != want {
		c.err = fmt.Errorf("%s: got %#x, want %#x", what, got, want)
	}
}

// origin reads the first word of an entry and refuses a value that has
// not been observed.
func (c *cursor) origin() uint32 {
	v := c.u32()
	switch v {
	case OriginVHDL, OriginVHDLTime, OriginVerilog, OriginVerilogPre, OriginVerilogTime:
	default:
		if c.err == nil {
			c.err = fmt.Errorf("origin word %#x", v)
		}
	}
	return v
}

// layout reads the u16 that follows the origin of an array or record.
func (c *cursor) layout() uint16 {
	v := c.u16()
	switch v {
	case LayoutVHDL, LayoutUnpacked, LayoutPacked:
	default:
		if c.err == nil {
			c.err = fmt.Errorf("layout word %#x", v)
		}
	}
	return v
}

// trailer reads the last word of an enumeration or real entry, which is
// 1 for a VHDL type and 0 for a Verilog one.
func (c *cursor) trailer(what string) uint32 {
	got := c.u32()
	if c.err == nil && got != 0 && got != 1 && got != 4 {
		c.err = fmt.Errorf("%s trailer: got %#x, want 0, 1 or 4", what, got)
	}
	return got
}

// enumSize is the byte size of one value of an enumeration type: the
// trailer word when it is 4, else 1.
func (t *Type) enumSize() int {
	if t.Trailer == 4 {
		return 4
	}
	return 1
}

// readType decodes one entry body: the name and the kind specific part.
func readType(kind Kind, body []byte) (Type, error) {
	t := Type{Kind: kind}
	c := &cursor{b: body}
	t.Name = c.str()
	switch kind {
	case KindEnum:
		t.Origin = c.origin()
		t.Variant = c.u32()
		t.Class = c.u32()
		n := int(c.u32())
		for i := 0; i < n && c.err == nil; i++ {
			t.Literals = append(t.Literals, c.str())
		}
		t.Trailer = c.trailer("enum")
	case KindValues:
		// [u32 1][u32 base][u32 n][u32 8] then n times name NUL [u64
		// value], then [u32 count] and count constraint triples for
		// the base type with no end marker. The 8 is the byte size of
		// a value. t11_sv_enum has a count of 0 over base int,
		// t11_sv_enum4 a count of 1 and (3 downto 0) over an unnamed
		// logic vector.
		t.Origin = c.origin()
		t.Elem = int(c.u32())
		n := int(c.u32())
		c.expect(8, "values word 3")
		for i := 0; i < n && c.err == nil; i++ {
			v := NamedValue{Name: c.str()}
			v.Value = c.u64()
			t.Values = append(t.Values, v)
		}
		nr := int(c.u32())
		for j := 0; j < nr && c.err == nil; j++ {
			t.Ranges = append(t.Ranges, Range{Left: c.i32(), Right: c.i32(), Dir: c.i32()})
		}
	case KindInteger:
		t.Origin = c.origin()
		t.Low = c.i32()
		t.High = c.i32()
		c.expect(1, "integer trailer")
	case KindReal:
		t.Origin = c.origin()
		t.Variant = c.u32()
		t.FLow = c.f64()
		t.FHigh = c.f64()
		t.Trailer = c.trailer("real")
	case KindAlias:
		t.Origin = c.origin()
		t.Target = int(c.u32())
		nr := int(c.u32())
		for j := 0; j < nr && c.err == nil; j++ {
			t.Ranges = append(t.Ranges, Range{Left: c.i32(), Right: c.i32(), Dir: c.i32()})
		}
	case KindAccess, KindFile:
		// The designated or element type, then two words whose
		// meaning is open: every file type holds 8 and 40, every
		// access type 8 and 48, whatever the element. See
		// docs/format.md.
		t.Origin = c.origin()
		t.Elem = int(c.u32())
		for i := 0; i < 2 && c.err == nil; i++ {
			t.Words = append(t.Words, c.u32())
		}
	case KindPhysical:
		t.Origin = c.origin()
		n := int(c.u32())
		for i := 0; i < n && c.err == nil; i++ {
			name := c.str()
			t.Units = append(t.Units, TimeUnit{Name: name, Scale: c.u64()})
		}
	case KindArray:
		t.Origin = c.origin()
		t.Layout = c.layout()
		c.expect16(0xa0, "array word 1 high half")
		t.Elem = int(c.u32())
		t.Dims = int(c.u32())
		for i := 0; i < t.Dims && c.err == nil; i++ {
			t.Indexes = append(t.Indexes, int(c.u32()))
		}
		if len(t.Indexes) > 0 {
			t.Index = t.Indexes[0]
		}
		nr := int(c.u32())
		t.Ranges = c.ranges()
		if c.err == nil && nr != len(t.Ranges) {
			c.err = fmt.Errorf("array range count word %d, %d triples follow", nr, len(t.Ranges))
		}
	case KindRecord:
		t.Origin = c.origin()
		t.Layout = c.layout()
		c.expect16(0xb, "record word 1 high half")
		n := int(c.u32())
		for i := 0; i < n && c.err == nil; i++ {
			f := Field{Name: c.str()}
			f.Type = int(c.u32())
			nr := int(c.u32())
			for j := 0; j < nr && c.err == nil; j++ {
				f.Ranges = append(f.Ranges, Range{Left: c.i32(), Right: c.i32(), Dir: c.i32()})
			}
			t.Fields = append(t.Fields, f)
		}
		if v := c.i32(); c.err == nil && v != rangeEnd {
			c.err = fmt.Errorf("record trailer: got %d, want %d", v, rangeEnd)
		}
	default:
		return t, fmt.Errorf("type %q has unknown kind %#x", t.Name, uint8(kind))
	}
	if c.err != nil {
		return t, fmt.Errorf("type %q (%s): %w", t.Name, kind, c.err)
	}
	if c.p != len(body) {
		return t, fmt.Errorf("type %q (%s): %d trailing bytes", t.Name, kind, len(body)-c.p)
	}
	return t, nil
}
