// SPDX-License-Identifier: Apache-2.0

// Package fstout writes a decoded waveform database as an FST file.
// The bytes of the file are libfst's; what this package decides is the
// mapping: which FST variable each object becomes, how a record or an
// array is flattened into leaves, and how a value is spelled. See
// docs/fst-output.md.
package fstout

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"git.hdlfactory.com/HDL/wdbcvt/pkg/fst"
	"git.hdlfactory.com/HDL/wdbcvt/pkg/wdb"
)

// leaf is one FST variable and the last value written to it.
type leaf struct {
	handle uint32
	kind   fst.VarType
	// path is the leaf's position inside the object's value: the field
	// and element indexes to walk down to it.
	path []int
	last string
	set  bool
}

// object holds what the writer needs per database object.
type object struct {
	decl   wdb.Decl
	leaves []*leaf
	// alias is set when another object on the same handle already
	// carries the values, so this one's changes are not written.
	alias bool
}

type writer struct {
	f    *wdb.File
	w    *fst.Writer
	objs []*object
	// shared indexes the leaves already declared for a handle, so a
	// second object on it becomes an alias rather than a second signal.
	shared map[sharedKey][]uint32
}

type sharedKey struct {
	handle uint64
	offset uint32
	typ    int
}

// Write converts f into an FST file at path.
func Write(f *wdb.File, path string) error {
	w, err := fst.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()
	out := &writer{f: f, w: w, objs: make([]*object, len(f.Objects)), shared: map[sharedKey][]uint32{}}
	// The database's time unit is a power of ten of a second, which is
	// what FST calls a timescale.
	w.SetTimescale(int(f.Debug.Precision))
	w.Comment("converted from a Vivado xsim waveform database by wdbcvt")
	byScope := make([][]int, len(f.Scopes))
	for i, o := range f.Objects {
		if o.Logged {
			byScope[o.Scope] = append(byScope[o.Scope], i)
		}
	}
	for _, root := range f.Scopes[0].Children() {
		if err := out.scope(root, byScope); err != nil {
			return err
		}
	}
	return out.values()
}

// scope declares one scope of the instance tree and everything under
// it.
func (o *writer) scope(s int, byScope [][]int) error {
	o.w.PushScope(plainName(o.f.Scopes[s].Name))
	defer o.w.PopScope()
	for _, i := range byScope[s] {
		if err := o.object(i); err != nil {
			return err
		}
	}
	for _, c := range o.f.Scopes[s].Children() {
		if err := o.scope(c, byScope); err != nil {
			return err
		}
	}
	return nil
}

// object declares the variables of one object: one for a scalar or a
// vector, and one per leaf of a record or an array of anything else,
// in a scope named after the object.
func (o *writer) object(i int) error {
	ob := o.f.Objects[i]
	dc := o.f.Decls[ob.Decl]
	n, err := o.f.ValueBytes(dc)
	if err != nil {
		return fmt.Errorf("%s: %w", o.f.ObjectPath(ob), err)
	}
	// The shape of a value follows from the declaration, so decoding a
	// value of zeroes gives the leaves without reading the file.
	shape, err := o.f.Decode(dc, make([]byte, n))
	if err != nil {
		return fmt.Errorf("%s: %w", o.f.ObjectPath(ob), err)
	}
	key := sharedKey{ob.Handle, ob.Offset, dc.Type}
	alias := o.shared[key]
	obj := &object{decl: dc, alias: alias != nil}
	o.objs[i] = obj
	if err := o.declare(plainName(dc.Name), shape, nil, obj, alias); err != nil {
		return fmt.Errorf("%s: %w", o.f.ObjectPath(ob), err)
	}
	if alias == nil {
		handles := make([]uint32, len(obj.leaves))
		for j, l := range obj.leaves {
			handles[j] = l.handle
		}
		o.shared[key] = handles
	}
	return nil
}

// declare walks a value's shape and creates one variable per leaf. A
// record and an array of anything but bits become a scope of their
// own, so a field or an element is a signal a viewer can show on its
// own. alias, when set, names the variables an earlier object on the
// same handle already created, leaf by leaf.
func (o *writer) declare(name string, v wdb.Value, path []int, obj *object, alias []uint32) error {
	ty := o.f.Base(v.Type)
	switch {
	case vector(o.f, ty):
		o.leaf(name, fst.VarWire, len(v.Elems), v, path, obj, alias)
	case ty.Kind == wdb.KindArray && charEnum(o.f, ty):
		o.leaf(name, fst.VarString, 0, v, path, obj, alias)
	case ty.Kind == wdb.KindArray:
		o.w.PushScope(name)
		defer o.w.PopScope()
		for i := range v.Elems {
			p := append(append([]int(nil), path...), i)
			if err := o.declare(strconv.Itoa(i), v.Elems[i], p, obj, alias); err != nil {
				return err
			}
		}
	case ty.Kind == wdb.KindRecord:
		o.w.PushScope(name)
		defer o.w.PopScope()
		for i := range v.Fields {
			p := append(append([]int(nil), path...), i)
			if err := o.declare(ty.Fields[i].Name, v.Fields[i], p, obj, alias); err != nil {
				return err
			}
		}
	default:
		kind, bits := scalarKind(o.f, ty)
		o.leaf(name, kind, bits, v, path, obj, alias)
	}
	return nil
}

// leaf creates one variable, or takes the handle of the aliased one.
func (o *writer) leaf(name string, kind fst.VarType, bits int, v wdb.Value, path []int, obj *object, alias []uint32) {
	var of uint32
	if alias != nil && len(obj.leaves) < len(alias) {
		of = alias[len(obj.leaves)]
	}
	h := o.w.Var(kind, name, bits, of)
	obj.leaves = append(obj.leaves, &leaf{handle: h, kind: kind, path: append([]int(nil), path...)})
}

// values streams every change of the file into the writer.
func (o *writer) values() error {
	var now uint64
	first := true
	err := o.f.Stream(func(i int, t uint64, data []byte) error {
		obj := o.objs[i]
		if obj == nil || obj.alias {
			return nil
		}
		if first || t != now {
			o.w.Time(t)
			now, first = t, false
		}
		v, err := o.f.Decode(obj.decl, data)
		if err != nil {
			return fmt.Errorf("%s: %w", o.f.ObjectPath(o.f.Objects[i]), err)
		}
		for _, l := range obj.leaves {
			at := v
			for _, p := range l.path {
				if len(at.Fields) > p {
					at = at.Fields[p]
				} else if len(at.Elems) > p {
					at = at.Elems[p]
				}
			}
			s := o.spell(l.kind, at)
			if l.set && s == l.last {
				continue
			}
			l.last, l.set = s, true
			switch l.kind {
			case fst.VarReal:
				x, _ := strconv.ParseFloat(s, 64)
				o.w.Real(l.handle, x)
			case fst.VarString:
				o.w.Str(l.handle, s)
			default:
				o.w.Value(l.handle, s)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !first {
		o.w.Time(o.f.Header.EndTime)
	}
	return nil
}

// spell renders one leaf value the way its variable holds it: the
// characters of the four or nine state logic for a wire, the two's
// complement bits for an integer or a time, the decimal text of a
// real, and the literal itself for everything carried as a string.
func (o *writer) spell(kind fst.VarType, v wdb.Value) string {
	switch kind {
	case fst.VarWire:
		ty := o.f.Base(v.Type)
		if ty.Kind == wdb.KindArray {
			var b strings.Builder
			for _, e := range v.Elems {
				b.WriteString(bitChar(e.Scalar))
			}
			return b.String()
		}
		return bitChar(v.Scalar)
	case fst.VarString:
		if len(v.Elems) > 0 {
			// A string is an array of characters, and the value of
			// each is the character itself.
			var b strings.Builder
			for _, e := range v.Elems {
				b.WriteString(e.Scalar)
			}
			return b.String()
		}
		return v.Scalar
	case fst.VarInteger, fst.VarTime:
		n, err := strconv.ParseInt(v.Scalar, 10, 64)
		if err != nil {
			return v.Scalar
		}
		bits := 32
		if kind == fst.VarTime || n > math.MaxInt32 || n < math.MinInt32 {
			bits = 64
		}
		return twos(n, bits)
	default:
		return v.Scalar
	}
}

// bitChar is the FST spelling of one logic value: the nine state
// characters in lower case, as VCD and FST write them.
func bitChar(s string) string {
	if s == "" {
		return "x"
	}
	return strings.ToLower(s[:1])
}

// twos renders n as bits characters of two's complement, high bit
// first, which is how FST holds an integer or a time.
func twos(n int64, bits int) string {
	b := make([]byte, bits)
	u := uint64(n)
	for i := 0; i < bits; i++ {
		b[bits-1-i] = byte('0' + (u>>uint(i))&1)
	}
	return string(b)
}

// plainName strips the bars and backslashes an extended identifier
// carries in the database: a generate iteration is `\g(0)\` in VHDL
// and a variable declared in a generate block is `\g[0].r ` in
// Verilog, and a viewer has no use for either decoration.
func plainName(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, `\`, ""))
}

// vector reports whether a type is an array of logic values, which FST
// holds as one variable of that many bits.
func vector(f *wdb.File, ty *wdb.Type) bool {
	return ty.Kind == wdb.KindArray && bitLike(f.Base(ty.Elem))
}

// bitLike reports whether an enumeration is a logic type: the two
// literals of BIT, the nine of STD_ULOGIC, or the four of Verilog
// logic, which the class word tells apart; see docs/format/types.md.
func bitLike(ty *wdb.Type) bool {
	return ty.Kind == wdb.KindEnum && (ty.Class == 1 || ty.Class == 2 || ty.Class == 3)
}

// charEnum reports whether a type is an array of characters, which FST
// holds as a string.
func charEnum(f *wdb.File, ty *wdb.Type) bool {
	e := f.Base(ty.Elem)
	return e.Kind == wdb.KindEnum && e.Class == 4
}

// scalarKind maps a scalar type onto an FST variable type and its
// width. Anything FST has no type for is carried as a string, which
// keeps the literal a viewer shows.
func scalarKind(f *wdb.File, ty *wdb.Type) (fst.VarType, int) {
	switch {
	case bitLike(ty):
		return fst.VarWire, 1
	case ty.Kind == wdb.KindInteger:
		return fst.VarInteger, 32
	case ty.Kind == wdb.KindReal:
		return fst.VarReal, 64
	case ty.Kind == wdb.KindPhysical:
		// A physical value is a count and the unit it counts,
		// `3000 um`, and FST has no type for that pair, so the text
		// carries it. FST_VT_VCD_TIME would hold the count alone.
		return fst.VarString, 0
	default:
		return fst.VarString, 0
	}
}
