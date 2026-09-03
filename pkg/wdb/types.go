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
	KindEnum     Kind = 0x03 // enumeration: BIT, STD_ULOGIC, BOOLEAN, CHARACTER, user enums
	KindInteger  Kind = 0x05 // integer with bounds
	KindReal     Kind = 0x06 // floating point with bounds
	KindPhysical Kind = 0x0d // physical type with units: TIME
	KindArray    Kind = 0x10 // array, constrained or not
	KindRecord   Kind = 0x11 // record with named fields
)

func (k Kind) String() string {
	switch k {
	case KindEnum:
		return "enum"
	case KindInteger:
		return "integer"
	case KindReal:
		return "real"
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
	// Name is the VHDL type name as the table stores it. Predefined and
	// IEEE types are upper case, user types keep the source case.
	Name string

	// Class is the third word of an enumeration entry: 2 for BIT, 3 for
	// STD_ULOGIC, 5 for BOOLEAN and user enumerations. What it selects is
	// open.
	Class uint32
	// Literals are the enumeration literals in declaration order.
	// Character literals keep their quotes, as in `'U'`.
	Literals []string

	// Low and High bound an integer type.
	Low, High int32
	// FLow and FHigh bound a real type.
	FLow, FHigh float64

	// Units lists a physical type's units.
	Units []TimeUnit

	// Elem and Index are type indexes for an array's element and index
	// types. Dims is its dimension count. Constrained says whether the
	// entry itself fixes the bounds; Ranges holds one Range per
	// dimension, then the element constraints in the order they appear.
	Elem, Index int
	Dims        int
	Constrained bool
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

// readType decodes one entry body: the name and the kind specific part.
func readType(kind Kind, body []byte) (Type, error) {
	t := Type{Kind: kind}
	c := &cursor{b: body}
	t.Name = c.str()
	switch kind {
	case KindEnum:
		c.expect(2, "enum word 0")
		c.expect(2, "enum word 1")
		t.Class = c.u32()
		n := int(c.u32())
		for i := 0; i < n && c.err == nil; i++ {
			t.Literals = append(t.Literals, c.str())
		}
		c.expect(1, "enum trailer")
	case KindInteger:
		c.expect(2, "integer word 0")
		t.Low = c.i32()
		t.High = c.i32()
		c.expect(1, "integer trailer")
	case KindReal:
		c.expect(2, "real word 0")
		c.expect(0, "real word 1")
		t.FLow = c.f64()
		t.FHigh = c.f64()
		c.expect(1, "real trailer")
	case KindPhysical:
		c.expect(0xa, "physical word 0")
		n := int(c.u32())
		for i := 0; i < n && c.err == nil; i++ {
			name := c.str()
			t.Units = append(t.Units, TimeUnit{Name: name, Scale: c.u64()})
		}
	case KindArray:
		c.expect(2, "array word 0")
		c.u16() // 1
		c.u16() // 0xa0
		t.Elem = int(c.u32())
		t.Dims = int(c.u32())
		t.Index = int(c.u32())
		switch v := c.u32(); v {
		case 1:
			t.Constrained = false
		case 2:
			t.Constrained = true
		default:
			if c.err == nil {
				c.err = fmt.Errorf("array constraint word %#x", v)
			}
		}
		t.Ranges = c.ranges()
	case KindRecord:
		c.expect(2, "record word 0")
		c.u16() // 1
		c.u16() // 0xb
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
