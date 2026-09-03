// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Value is a decoded object value: a scalar, or a tree of them.
type Value struct {
	// Type indexes File.Types.
	Type int
	// Scalar holds a scalar's text: an enumeration literal with any
	// quotes removed, a decimal integer, a real, or a picosecond count.
	Scalar string
	// Elems holds an array's elements, left to right.
	Elems []Value
	// Fields holds a record's fields in declaration order.
	Fields []Value
}

// layout is the size and alignment of one type instance in a page
// record, as measured on the corpus. See docs/format.md, "Values".
type layout struct {
	size, align int
}

func roundUp(n, a int) int {
	if a <= 1 {
		return n
	}
	return (n + a - 1) / a * a
}

// takeRanges consumes n ranges from the front of the constraint list.
func takeRanges(rs *[]Range, n int, what string) ([]Range, error) {
	if len(*rs) < n {
		return nil, fmt.Errorf("%s needs %d ranges, %d left", what, n, len(*rs))
	}
	out := (*rs)[:n]
	*rs = (*rs)[n:]
	return out, nil
}

// fieldRanges is the constraint list a record field brings for its type.
// A field whose type is itself a record starts with one extra triple
// before the nested constraints; what it means is open, so it is dropped.
func (f *File) fieldRanges(fd Field) []Range {
	rs := fd.Ranges
	if fd.Type >= 0 && fd.Type < len(f.Types) && f.Types[fd.Type].Kind == KindRecord && len(rs) > 0 {
		rs = rs[1:]
	}
	return rs
}

// layoutOf computes the layout of type t, consuming index constraints
// from rs for every array dimension it meets.
func (f *File) layoutOf(t int, rs *[]Range) (layout, error) {
	if t < 0 || t >= len(f.Types) {
		return layout{}, fmt.Errorf("type index %d of %d", t, len(f.Types))
	}
	ty := &f.Types[t]
	switch ty.Kind {
	case KindEnum:
		n := ty.enumSize()
		return layout{n, n}, nil
	case KindInteger:
		return layout{4, 4}, nil
	case KindReal, KindPhysical:
		return layout{8, 8}, nil
	case KindArray:
		dims, err := takeRanges(rs, ty.Dims, ty.Name)
		if err != nil {
			return layout{}, err
		}
		n := 1
		for _, r := range dims {
			n *= r.Length()
		}
		el, err := f.layoutOf(ty.Elem, rs)
		if err != nil {
			return layout{}, err
		}
		return layout{n * roundUp(el.size, el.align), el.align}, nil
	case KindRecord:
		off := 0
		for _, fd := range ty.Fields {
			frs := f.fieldRanges(fd)
			l, err := f.layoutOf(fd.Type, &frs)
			if err != nil {
				return layout{}, fmt.Errorf("%s.%s: %w", ty.Name, fd.Name, err)
			}
			off = roundUp(off, l.align) + l.size
		}
		return layout{roundUp(off, 8), 8}, nil
	case KindFile:
		// A file variable declares 0 bytes and never has a record:
		// INPUT and OUTPUT of textio in t22_dbg_all.
		return layout{}, nil
	}
	return layout{}, fmt.Errorf("type %q has kind %s", ty.Name, ty.Kind)
}

// Size is the declared size of the object as the type table implies it:
// bytes of a page record for a VHDL object, bits for a Verilog one. It
// is checked against Decl.Size by the corpus test. See docs/format/values.md.
func (f *File) Size(dc Decl) (int, error) {
	rs := dc.Ranges
	if f.verilog(dc.Type) {
		n, err := f.bitsOf(dc.Type, &rs)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", dc.Name, err)
		}
		return n, nil
	}
	l, err := f.layoutOf(dc.Type, &rs)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", dc.Name, err)
	}
	return l.size, nil
}

// Decode turns the raw bytes of a change into a Value, using the
// declaration's type and index constraints.
func (f *File) Decode(dc Decl, data []byte) (Value, error) {
	rs := dc.Ranges
	if f.verilog(dc.Type) {
		v, err := f.decodeVerilog(dc.Type, data, &rs)
		if err != nil {
			return v, fmt.Errorf("%s: %w", dc.Name, err)
		}
		return v, nil
	}
	v, n, err := f.decode(dc.Type, data, &rs)
	if err != nil {
		return v, fmt.Errorf("%s: %w", dc.Name, err)
	}
	if n != len(data) {
		return v, fmt.Errorf("%s: decoded %d of %d bytes", dc.Name, n, len(data))
	}
	return v, nil
}

func (f *File) decode(t int, data []byte, rs *[]Range) (Value, int, error) {
	ty := &f.Types[t]
	v := Value{Type: t}
	need := func(n int) error {
		if len(data) < n {
			return fmt.Errorf("%s: need %d bytes, have %d", ty.Name, n, len(data))
		}
		return nil
	}
	switch ty.Kind {
	case KindEnum:
		n := ty.enumSize()
		if err := need(n); err != nil {
			return v, 0, err
		}
		i := int(data[0])
		if n == 4 {
			i = int(binary.LittleEndian.Uint32(data))
		}
		if i >= len(ty.Literals) {
			return v, 0, fmt.Errorf("%s: literal index %d of %d", ty.Name, i, len(ty.Literals))
		}
		v.Scalar = strings.Trim(ty.Literals[i], "'")
		return v, n, nil
	case KindInteger:
		if err := need(4); err != nil {
			return v, 0, err
		}
		v.Scalar = strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(data))), 10)
		return v, 4, nil
	case KindReal:
		if err := need(8); err != nil {
			return v, 0, err
		}
		v.Scalar = strconv.FormatFloat(math.Float64frombits(binary.LittleEndian.Uint64(data)), 'g', -1, 64)
		return v, 8, nil
	case KindPhysical:
		if err := need(8); err != nil {
			return v, 0, err
		}
		v.Scalar = strconv.FormatInt(int64(binary.LittleEndian.Uint64(data)), 10) + " " + ty.baseUnit()
		return v, 8, nil
	case KindArray:
		dims, err := takeRanges(rs, ty.Dims, ty.Name)
		if err != nil {
			return v, 0, err
		}
		n := 1
		for _, r := range dims {
			n *= r.Length()
		}
		// Every element is bound by the same constraints, so each one
		// decodes from a copy of the list and the list advances once.
		ers := *rs
		el, err := f.layoutOf(ty.Elem, &ers)
		if err != nil {
			return v, 0, err
		}
		stride := roundUp(el.size, el.align)
		if err := need(n * stride); err != nil {
			return v, 0, err
		}
		for i := 0; i < n; i++ {
			ers = *rs
			e, _, err := f.decode(ty.Elem, data[i*stride:(i+1)*stride], &ers)
			if err != nil {
				return v, 0, err
			}
			v.Elems = append(v.Elems, e)
		}
		*rs = ers
		return v, n * stride, nil
	case KindRecord:
		off := 0
		for _, fd := range ty.Fields {
			frs := f.fieldRanges(fd)
			l, err := f.layoutOf(fd.Type, &frs)
			if err != nil {
				return v, 0, fmt.Errorf("%s.%s: %w", ty.Name, fd.Name, err)
			}
			off = roundUp(off, l.align)
			if err := need(off + l.size); err != nil {
				return v, 0, err
			}
			frs = f.fieldRanges(fd)
			fv, _, err := f.decode(fd.Type, data[off:off+l.size], &frs)
			if err != nil {
				return v, 0, fmt.Errorf("%s.%s: %w", ty.Name, fd.Name, err)
			}
			v.Fields = append(v.Fields, fv)
			off += l.size
		}
		size := roundUp(off, 8)
		if err := need(size); err != nil {
			return v, 0, err
		}
		return v, size, nil
	}
	return v, 0, fmt.Errorf("type %q has kind %s", ty.Name, ty.Kind)
}

// String renders a value the way the corpus truth files write it: an
// array of character literals as one string of them (`10100101`),
// other arrays in parentheses, records as
// `(name => value, ...)`. A Verilog value that decoded to a scalar
// spelling, a decimal integer or a named enum value, prints that.
// charEnum reports whether an enumeration type has character literals:
// a VHDL literal written in quotes (`'0'`, and CHARACTER has those next
// to identifiers such as `nul`) or a Verilog value spelling (`Z`). An
// array of such a type prints as one string, an array of identifier
// literals (t20_enum_300_arr) as a list.
func (t *Type) charEnum() bool {
	if t.Kind != KindEnum {
		return false
	}
	for _, l := range t.Literals {
		if len(l) == 1 || strings.HasPrefix(l, "'") {
			return true
		}
	}
	return false
}

// baseUnit names the unit a physical value is counted in: the one
// whose scale is 1, `ps` for TIME and `um` for the dist_t of
// t21_phys_user. The scale of `fs` is 0, so TIME counts picoseconds.
func (t *Type) baseUnit() string {
	for _, u := range t.Units {
		if u.Scale == 1 {
			return u.Name
		}
	}
	return "?"
}

func (f *File) String(v Value) string {
	ty := &f.Types[f.resolve(v.Type)]
	if v.Scalar != "" {
		return v.Scalar
	}
	switch ty.Kind {
	case KindArray:
		if f.Types[f.resolve(ty.Elem)].charEnum() {
			var b strings.Builder
			for _, e := range v.Elems {
				b.WriteString(e.Scalar)
			}
			return b.String()
		}
		parts := make([]string, len(v.Elems))
		for i, e := range v.Elems {
			parts[i] = f.String(e)
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case KindRecord:
		parts := make([]string, len(v.Fields))
		for i, fv := range v.Fields {
			parts[i] = ty.Fields[i].Name + " => " + f.String(fv)
		}
		return "(" + strings.Join(parts, ", ") + ")"
	}
	return v.Scalar
}

// Leaf is one non-record part of a value, named by its dotted path.
type Leaf struct {
	Path  string
	Type  int
	Value Value
}

// Leaves flattens records into their fields, recursively, the way the
// truth files name them: `s.delta_f.bravo`. Arrays stay whole.
func (f *File) Leaves(path string, v Value) []Leaf {
	ty := &f.Types[f.resolve(v.Type)]
	if ty.Kind != KindRecord {
		return []Leaf{{Path: path, Type: v.Type, Value: v}}
	}
	var out []Leaf
	for i, fv := range v.Fields {
		out = append(out, f.Leaves(path+"."+ty.Fields[i].Name, fv)...)
	}
	return out
}
