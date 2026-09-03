// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Layout constants of the Xilinx ISim DBG 006 section. Every offset in
// the section's own table is relative to the start of its magic string.
const (
	dbgOffsets  = 18 // uint32 region offsets, after the magic, timestamp and precision
	dbgCounts   = 4  // uint32 counts after the offsets
	dbgWords    = 17 // uint32 header words after the counts
	classRecLen = 3  // words per value class entry of region 17

	scopeRecLen = 9  // uint32 words per scope record
	unitRecLen  = 9  // uint32 words per unit record
	declRecLen  = 11 // uint32 words per declaration record
	rangeRecLen = 6  // uint32 words per range record
	instRecLen  = 56 // bytes per instance record after the section
)

// UnitKind is the third word of a unit record.
type UnitKind uint32

// The unit kinds observed so far.
const (
	UnitTop     UnitKind = 0x13 // the root above the top level entity
	UnitEntity  UnitKind = 0x09 // an entity and architecture pair
	UnitProcess UnitKind = 0x0d // a process
	// UnitGenerate is a generate statement, and also each iteration of
	// a for generate: t7_gen_for has one for `g` and one per `\g(i)\`.
	UnitGenerate UnitKind = 0x0c
	// The Verilog unit kinds. A module, a named block, and a process:
	// an initial or always block, a continuous assignment, or the
	// implicit initial block a reg initializer makes. Found by
	// t11_v_bit_edge against t1_bit_one_edge, t11_v_always, t11_v_wire.
	UnitModule   UnitKind = 0x00
	UnitBlock    UnitKind = 0x05
	UnitVProcess UnitKind = 0x07
	// A task and a function, each a scope of its own holding its
	// arguments and local variables, and for a function its return
	// variable first: t12_v_task, t12_v_func.
	UnitTask     UnitKind = 0x03
	UnitFunction UnitKind = 0x04
	// A SystemVerilog interface, modport and package: t13_sv_iface,
	// t15_sv_iface_mp, t13_sv_pkg.
	UnitInterface UnitKind = 0x01
	UnitModport   UnitKind = 0x02
	UnitSVPackage UnitKind = 0x08
	// A VHDL package, a child of the root beside the top entity:
	// t9_port_rec. Under -debug all the precompiled library packages
	// appear the same way: t22_dbg_all.
	UnitPackage UnitKind = 0x0a
	// A VHDL function and procedure, present only under -debug
	// subprogram or all, with the parameters and variables as
	// declarations: inc of t22_dbg_subprog, bump of t22_dbg_sub_proc.
	UnitVHDLFunction  UnitKind = 0x11
	UnitVHDLProcedure UnitKind = 0x12
)

func (k UnitKind) String() string {
	switch k {
	case UnitTop:
		return "top"
	case UnitEntity:
		return "entity"
	case UnitProcess:
		return "process"
	case UnitGenerate:
		return "generate"
	case UnitModule:
		return "module"
	case UnitBlock:
		return "block"
	case UnitVProcess:
		return "vprocess"
	case UnitTask:
		return "task"
	case UnitFunction:
		return "function"
	case UnitInterface:
		return "interface"
	case UnitModport:
		return "modport"
	case UnitSVPackage:
		return "svpackage"
	case UnitPackage:
		return "package"
	case UnitVHDLFunction:
		return "vhdl function"
	case UnitVHDLProcedure:
		return "vhdl procedure"
	}
	return fmt.Sprintf("unit(%#x)", uint32(k))
}

// DeclKind is the ninth word of a declaration record.
type DeclKind uint32

// subprogram reports whether the kind is a parameter or a local of a
// VHDL subprogram: an object with no record, which appears only under
// -debug subprogram or all.
func (k DeclKind) subprogram() bool {
	return k == DeclLocal || k == DeclSignalParam
}

// vhdlData reports whether the kind is a VHDL object other than a
// signal: a variable, a constant, a generic or a subprogram local.
func (k DeclKind) vhdlData() bool {
	return k == DeclVariable || k == DeclConstant || k == DeclGeneric || k.subprogram()
}

// The declaration kinds observed so far.
const (
	DeclSignal   DeclKind = 0x0e
	DeclVariable DeclKind = 0x0f // a variable declared in a process
	DeclGeneric  DeclKind = 0x12
	DeclConstant DeclKind = 0x13 // a constant, a for loop index or a generate index
	// DeclLocal is a parameter or a variable of a VHDL subprogram, seen
	// only under -debug subprogram or all. A parameter carries its mode
	// as a port does. Found by t22_dbg_subprog against t22_base.
	DeclLocal DeclKind = 0x14
	// DeclSignalParam is a signal parameter of a VHDL subprogram. It
	// carries its mode as a port does. Found by t23_sub_sig_prm
	// against t22_dbg_sub_proc.
	DeclSignalParam DeclKind = 0x15
	// The Verilog declaration kinds: a variable (reg, logic, integer
	// and the other variable types), a parameter, and a net (wire and
	// every port of t11_v_port). Found by tier 11. The other net
	// kinds follow wire in the order the Verilog standard lists them,
	// with 0x0b unseen where trireg would sit, since xsim refuses a
	// trireg. Found by tier 19. A uwire is a wire.
	DeclVar     DeclKind = 0x00
	DeclParam   DeclKind = 0x01
	DeclNet     DeclKind = 0x03
	DeclWand    DeclKind = 0x04
	DeclWor     DeclKind = 0x05
	DeclTri     DeclKind = 0x06
	DeclTriand  DeclKind = 0x07
	DeclTrior   DeclKind = 0x08
	DeclTri0    DeclKind = 0x09
	DeclTri1    DeclKind = 0x0a
	DeclSupply0 DeclKind = 0x0c
	DeclSupply1 DeclKind = 0x0d
)

// netKinds maps the Verilog net keyword to its declaration kind.
var netKinds = map[string]DeclKind{
	"wire": DeclNet, "uwire": DeclNet, "wand": DeclWand, "wor": DeclWor,
	"tri": DeclTri, "triand": DeclTriand, "trior": DeclTrior,
	"tri0": DeclTri0, "tri1": DeclTri1,
	"supply0": DeclSupply0, "supply1": DeclSupply1,
}

// IsNet reports whether the kind is one of the Verilog net kinds.
func (k DeclKind) IsNet() bool {
	for _, n := range netKinds {
		if n == k {
			return true
		}
	}
	return false
}

func (k DeclKind) String() string {
	switch k {
	case DeclSignal:
		return "signal"
	case DeclGeneric:
		return "generic"
	case DeclVariable:
		return "variable"
	case DeclConstant:
		return "constant"
	case DeclLocal:
		return "local"
	case DeclSignalParam:
		return "signal param"
	case DeclVar:
		return "var"
	case DeclParam:
		return "parameter"
	case DeclNet:
		return "net"
	}
	for name, n := range netKinds {
		if n == k && name != "uwire" {
			return name
		}
	}
	return fmt.Sprintf("decl(%#x)", uint32(k))
}

// Scope is one node of the instance tree: the root `_top`, an entity
// instance, or a process.
type Scope struct {
	Name string
	// Parent indexes File.Scopes, or is -1 for the root.
	Parent int
	// FirstChild and NumChildren locate the children, which are stored
	// consecutively. FirstChild is -1 when there are none.
	FirstChild, NumChildren int
	// FirstObject indexes File.Objects for the scope's first object, or
	// is -1 when the scope has none.
	FirstObject int
	// File indexes File.Files; Line is the source line.
	File, Line int
	// Unit indexes File.Units.
	Unit int
}

// Unit is one elaborated design unit: the root, an entity plus
// architecture, or a process. Instances of an entity with equal generics
// share one unit; different generics give one unit each.
type Unit struct {
	// Name and Arch are the entity and architecture names, empty for the
	// root and for processes.
	Name, Arch string
	Kind       UnitKind
	// FirstDecl and NumDecls locate the unit's declarations in File.Decls.
	FirstDecl, NumDecls int
	// File and Line locate the architecture; EntityFile and EntityLine the
	// entity declaration.
	File, Line             int
	EntityFile, EntityLine int
}

// PortMode is the tenth word of a declaration record: the port mode of
// a signal declared in an entity's port list, or PortNone for anything
// declared elsewhere. The values follow the order of the VHDL modes.
type PortMode uint32

// The port modes observed so far.
const (
	PortInout   PortMode = 0
	PortIn      PortMode = 1
	PortOut     PortMode = 2
	PortBuffer  PortMode = 3
	PortLinkage PortMode = 4
	PortNone    PortMode = 5 // not a port
)

func (m PortMode) String() string {
	switch m {
	case PortInout:
		return "inout"
	case PortIn:
		return "in"
	case PortOut:
		return "out"
	case PortBuffer:
		return "buffer"
	case PortLinkage:
		return "linkage"
	case PortNone:
		return "none"
	}
	return fmt.Sprintf("mode %#x", uint32(m))
}

// Decl is one declared object: a signal, a port or a generic of a unit,
// or a variable or a constant of a process.
type Decl struct {
	Name string
	// File and Line locate the declaration.
	File, Line int
	// Size is the value size: bytes for a VHDL type, as a page record
	// holds it, and bits for a Verilog type, whose records hold word
	// pairs instead; see verilog.go.
	Size int
	// Type indexes File.Types.
	Type int
	// Ranges are the object's index constraints from the range table.
	Ranges []Range
	Kind   DeclKind
	// Class indexes Debug.Classes: the entry holding the value class of
	// the declaration's objects. Word 1 of the record; tier 31,
	// t31_sv_w1_swap against t31_sv_w1_i5, where swapping two
	// declarations swapped the word, and t12_v_params, whose six
	// declarations index the three entries [0 0 0] [3 0 0] [1 0 0] as
	// 0, 1, 2, 1, 0, 1.
	Class int
	// Mode is the port mode; PortNone for a signal that is not a port.
	// A port connected to a parent signal shares that signal's handle,
	// so the two objects read the same records: t8_port_in.
	Mode PortMode
}

// SourceFile is one entry of the file table.
type SourceFile struct {
	// Path is the path the file was compiled from. Local is the path the
	// same file resolves to on the machine that wrote the database; the
	// two differ for AMD's precompiled libraries.
	Path, Local string
}

// Object is one instance of a declaration in one scope. It is what a
// value page refers to, through Handle.
type Object struct {
	// Handle is the number a page record uses to name the object: the
	// arena index in the upper bits and the record key in the low 11.
	// Handles step by 0xf0 within an arena.
	Handle uint64
	// Offset is the byte offset of the object's value inside the value
	// recorded at Handle. It is nonzero for a port bound to a slice of
	// its actual: t9_port_slice binds a to x(0) of a 2 bit x and has
	// handle x, offset 1.
	Offset uint32
	// Scope and Decl index File.Scopes and File.Decls.
	Scope, Decl int
	// Logged is set for an object inside one of the header's logged
	// ranges: it has at least one record. An unlogged object has none:
	// a variable, a package object, or a signal outside the logged
	// hierarchy.
	Logged bool
	// Generic is set for a generic and for a variable: objects with no
	// second handle, whose instance record has 2 in its fifth word. A
	// signal has 1 there. A generic and a constant get one record
	// at time 0; a declared variable gets none, and its arena may not
	// exist.
	Generic bool
	// Storage is word 28 of the instance record, read as the storage
	// class of the object: 0 for a signal, 1 for a port on a language
	// boundary, 2 for a generic, constant, variable or loop index, 3
	// for a scalar subprogram local, 4 for an array subprogram local
	// and 6 for a signal parameter of a subprogram, tier 49.
	Storage uint32
	// Position is the tenth word of the instance record: for a Verilog
	// port, its place in the module's port list counted from 0, by the
	// port list and not by the connection or by the declaration order
	// of a non-ANSI header. It is 0 for a VHDL port, tier 48.
	Position uint32
}

// Arena selects the arena whose pages hold the object's records.
func (o Object) Arena() int { return int(o.Handle / arenaSpan) }

// Key is the record key the object has inside the arena's pages.
func (o Object) Key() uint32 { return uint32(o.Handle % arenaSpan) }

// Debug is the decoded Xilinx ISim DBG 006 section.
type Debug struct {
	Timestamp uint32
	// Precision is the power of ten of the file's time unit in seconds:
	// -12 for the picosecond of every VHDL case and of a Verilog source
	// under `timescale 1ns / 1ps`, -9, -10 and -15 for a Verilog
	// precision of 1ns, 100ps and 1fs, tier 21.
	Precision int32
	// Offsets is the section's own table of region offsets, kept for
	// the regions that are not decoded yet.
	Offsets [dbgOffsets]uint32
	Counts  [dbgCounts]uint32
	// Words holds the 17 header words after the counts. Most are
	// sizes the regions repeat; words 14 and 15 are 0x101 and grow a
	// third byte under -debug readers and -debug subprogram: tier 24,
	// t24_dbg_readers_ against t24_two_drivers_ and the -debug
	// subprogram cases of tiers 22 and 23 against their siblings.
	Words [dbgWords]uint32
	// Classes is region 17: one entry per distinct class of initial
	// value among the objects, as many as word 13 counts. The other two
	// words of an entry are 0 in every case. VHDL files hold the one
	// class 0; a Verilog file classes its objects by type and
	// initializer: tier 25, t25_sv_two_class against t25_sv_two_same.
	Classes [][3]uint32
	// End is the file offset one past the section proper; the instance
	// records start there and run to the end of the directory entry's
	// length.
	End uint64
}

type dbgSection struct {
	d    []byte
	base uint64
	Debug
}

func (s *dbgSection) region(from, to int) ([]byte, error) {
	a, b := s.base+uint64(s.Offsets[from]), s.base+uint64(s.Offsets[to])
	if a > b || b > uint64(len(s.d)) {
		return nil, fmt.Errorf("DBG region %d..%d spans %#x..%#x", from, to, a, b)
	}
	return s.d[a:b], nil
}

// checkCounts holds header words 0 to 13 against the regions they
// count: word i counts region i+4, the way the four counts before them
// count regions 0 to 3. A record region is counted in records, a name
// pool in bytes before its padding, and an empty region holds 0.
// Region 17 ends at offset 2, the end of the section proper.
func (s *dbgSection) checkCounts() error {
	for k := 4; k <= 17; k++ {
		to := k + 1
		if k == 17 {
			to = 2
		}
		r, err := s.region(k, to)
		if err != nil {
			return err
		}
		w := int(s.Words[k-4])
		var want int
		switch k {
		case 4:
			want = w * rangeRecLen * 4
		case 9, 10, 11:
			// The pool's length before padding is the last NUL of a
			// name and one; a pool with no names has no bytes.
			if got := len(bytes.TrimRight(r, "\x00")); got > 0 || len(r) > 0 {
				want = got + 1
				if want != w {
					return fmt.Errorf("DBG header word %d is %d for a %d byte name pool in region %d", k-4, w, want, k)
				}
			}
			continue
		case 13, 14:
			want = w * 8
		case 15:
			want = (w*4 + 7) / 8 * 8
		case 17:
			want = (w*classRecLen*4 + 7) / 8 * 8
		default:
			want = 0
		}
		if len(r) != want {
			return fmt.Errorf("DBG header word %d is %d, which does not fit the %d bytes of region %d", k-4, w, len(r), k)
		}
	}
	return nil
}

func words(b []byte) []int32 {
	out := make([]int32, len(b)/4)
	for i := range out {
		out[i] = int32(binary.LittleEndian.Uint32(b[4*i:]))
	}
	return out
}

func poolString(pool []byte, off int32) (string, error) {
	if off < 0 || int(off) >= len(pool) {
		return "", fmt.Errorf("string offset %d outside a %d byte pool", off, len(pool))
	}
	return cstring(pool[off:]), nil
}

// readDebug decodes the section named by the Xilinx DBG directory entry.
func readDebug(f *File, d []byte, dbg DirEntry) error {
	s := &dbgSection{d: d, base: dbg.Offset}
	sec := d[dbg.Offset : dbg.Offset+dbg.Length]
	if cstring(sec) != debugMagic {
		return fmt.Errorf("no %q magic at %#x", debugMagic, dbg.Offset)
	}
	c := &cursor{b: sec, p: len(debugMagic) + 1}
	s.Timestamp = c.u32()
	s.Precision = c.i32()
	if c.err == nil && (s.Precision > 0 || s.Precision < -15) {
		return fmt.Errorf("DBG time precision is 10^%d s, want -15 to 0", s.Precision)
	}
	for i := range s.Offsets {
		s.Offsets[i] = c.u32()
	}
	for i := range s.Counts {
		s.Counts[i] = c.u32()
	}
	for i := range s.Words {
		s.Words[i] = c.u32()
	}
	if c.err != nil {
		return fmt.Errorf("DBG header: %w", c.err)
	}
	// The directory entry's length covers the section and the instance
	// records that follow it; offset 2 is where the records start.
	s.End = dbg.Offset + uint64(s.Offsets[2])
	if rest := dbg.Offset + dbg.Length - s.End; s.End > dbg.Offset+dbg.Length || rest != uint64(s.Counts[2])*instRecLen {
		return fmt.Errorf("DBG section ends at %#x and the directory entry at %#x, which does not fit %d instance records",
			s.End, dbg.Offset+dbg.Length, s.Counts[2])
	}
	if err := s.checkCounts(); err != nil {
		return err
	}
	// Region 17 runs to the end of the section proper, as offset 2
	// says. Its entries are three words, padded to 8 bytes at the end.
	cw, err := s.region(17, 2)
	if err != nil {
		return err
	}
	n := int(s.Words[13])
	if len(cw) != (n*classRecLen*4+7)/8*8 {
		return fmt.Errorf("DBG region 17 holds %d bytes for %d value classes", len(cw), n)
	}
	for w, i := words(cw), 0; i < n; i++ {
		r := w[i*classRecLen:]
		s.Classes = append(s.Classes, [3]uint32{uint32(r[0]), uint32(r[1]), uint32(r[2])})
	}
	f.Debug = s.Debug

	scopePool, err := s.region(5, 10)
	if err != nil {
		return err
	}
	declPool, err := s.region(10, 11)
	if err != nil {
		return err
	}
	filePool, err := s.region(11, 12)
	if err != nil {
		return err
	}

	// Files: pairs of offsets into the file pool. -1 is an unused slot.
	fw, err := s.region(12, 14)
	if err != nil {
		return err
	}
	for w := words(fw); len(w) >= 2; w = w[2:] {
		var sf SourceFile
		if w[0] >= 0 {
			if sf.Path, err = poolString(filePool, w[0]); err != nil {
				return fmt.Errorf("file %d: %w", len(f.Files), err)
			}
		}
		if w[1] >= 0 {
			if sf.Local, err = poolString(filePool, w[1]); err != nil {
				return fmt.Errorf("file %d: %w", len(f.Files), err)
			}
		}
		f.Files = append(f.Files, sf)
	}

	// Scopes.
	sw, err := s.region(0, 1)
	if err != nil {
		return err
	}
	n = int(s.Counts[0])
	w := words(sw)
	if len(w) < n*scopeRecLen {
		return fmt.Errorf("DBG holds %d words for %d scopes", len(w), n)
	}
	for i := 0; i < n; i++ {
		r := w[i*scopeRecLen:]
		name, err := poolString(scopePool, r[0])
		if err != nil {
			return fmt.Errorf("scope %d: %w", i, err)
		}
		f.Scopes = append(f.Scopes, Scope{
			Name:        name,
			Parent:      int(r[1]),
			NumChildren: int(r[3]),
			FirstChild:  int(r[4]),
			FirstObject: int(r[5]),
			File:        int(r[6]),
			Line:        int(r[7]),
			Unit:        int(r[8]),
		})
	}

	// Units.
	uw, err := s.region(1, 3)
	if err != nil {
		return err
	}
	n = int(s.Counts[1])
	w = words(uw)
	if len(w) < n*unitRecLen {
		return fmt.Errorf("DBG holds %d words for %d units", len(w), n)
	}
	for i := 0; i < n; i++ {
		r := w[i*unitRecLen:]
		u := Unit{
			Kind:       UnitKind(r[2]),
			NumDecls:   int(r[3]),
			FirstDecl:  int(r[4]),
			File:       int(r[5]),
			Line:       int(r[6]),
			EntityFile: int(r[7]),
			EntityLine: int(r[8]),
		}
		if r[0] >= 0 {
			if u.Name, err = poolString(scopePool, r[0]); err != nil {
				return fmt.Errorf("unit %d: %w", i, err)
			}
		}
		if r[1] >= 0 {
			if u.Arch, err = poolString(scopePool, r[1]); err != nil {
				return fmt.Errorf("unit %d: %w", i, err)
			}
		}
		f.Units = append(f.Units, u)
	}

	// Ranges, then the declarations that point into them.
	rw, err := s.region(4, 5)
	if err != nil {
		return err
	}
	var ranges []Range
	for w := words(rw); len(w) >= rangeRecLen; w = w[rangeRecLen:] {
		// Each bound is a 64 bit pair; only the low word is used here.
		ranges = append(ranges, Range{Left: w[0], Right: w[2], Dir: w[4]})
	}

	dw, err := s.region(3, 4)
	if err != nil {
		return err
	}
	n = int(s.Counts[3])
	w = words(dw)
	if len(w) < n*declRecLen {
		return fmt.Errorf("DBG holds %d words for %d declarations", len(w), n)
	}
	for i := 0; i < n; i++ {
		r := w[i*declRecLen:]
		name, err := poolString(declPool, r[0])
		if err != nil {
			return fmt.Errorf("declaration %d: %w", i, err)
		}
		dc := Decl{
			Name:  name,
			File:  int(r[2]),
			Line:  int(r[3]),
			Size:  int(r[4]),
			Type:  int(r[5]),
			Kind:  DeclKind(r[8]),
			Mode:  PortMode(r[9]),
			Class: int(r[1]),
		}
		if dc.Type < 0 || dc.Type >= len(f.Types) {
			return fmt.Errorf("declaration %q has type index %d of %d", name, dc.Type, len(f.Types))
		}
		// A file without objects has no entries, and its declarations
		// hold 0: t0_nosig.
		if dc.Class < 0 || dc.Class >= len(f.Debug.Classes) && !(dc.Class == 0 && len(f.Debug.Classes) == 0) {
			return fmt.Errorf("declaration %q has value class index %d of %d", name, dc.Class, len(f.Debug.Classes))
		}
		nr, first := int(r[6]), int(r[7])
		if nr > 0 {
			if first < 0 || first+nr > len(ranges) {
				return fmt.Errorf("declaration %q has ranges %d+%d of %d", name, first, nr, len(ranges))
			}
			dc.Ranges = ranges[first : first+nr]
		}
		f.Decls = append(f.Decls, dc)
	}

	// Instances follow the section, up to the Xilinx DBG directory entry.
	n = int(s.Counts[2])
	a := s.End
	if a+uint64(n*instRecLen) > uint64(len(d)) {
		return fmt.Errorf("%d instance records at %#x run past the file", n, a)
	}
	for i := 0; i < n; i++ {
		r := d[a+uint64(i*instRecLen):]
		o := Object{
			Handle:   binary.LittleEndian.Uint64(r[0:]),
			Scope:    int(binary.LittleEndian.Uint32(r[16:])),
			Offset:   binary.LittleEndian.Uint32(r[20:]),
			Decl:     int(binary.LittleEndian.Uint64(r[32:])),
			Generic:  binary.LittleEndian.Uint32(r[28:]) == 2,
			Storage:  binary.LittleEndian.Uint32(r[28:]),
			Position: binary.LittleEndian.Uint32(r[40:]),
		}
		if o.Scope < 0 || o.Scope >= len(f.Scopes) {
			return fmt.Errorf("object %d has scope %d of %d", i, o.Scope, len(f.Scopes))
		}
		if o.Decl < 0 || o.Decl >= len(f.Decls) {
			return fmt.Errorf("object %d has declaration %d of %d", i, o.Decl, len(f.Decls))
		}
		f.Objects = append(f.Objects, o)
	}
	return nil
}

// TimeUnit names the unit every time in the file is counted in: the
// end time, the page bounds and the record times. It is 10 to the
// Debug.Precision seconds: `ps` for every VHDL case and for a Verilog
// source under `timescale 1ns / 1ps`, and `ns`, `100ps` or `fs` for
// the tier 21 time scales.
func (f *File) TimeUnit() string {
	names := map[int32]string{0: "s", -3: "ms", -6: "us", -9: "ns", -12: "ps", -15: "fs"}
	p := f.Debug.Precision
	base := p
	scale := 1
	for base%3 != 0 {
		base--
		scale *= 10
	}
	if scale == 1 {
		return names[base]
	}
	return fmt.Sprintf("%d%s", scale, names[base])
}

// TimeFS converts a time in the file's unit to femtoseconds.
func (f *File) TimeFS(t uint64) uint64 {
	for p := f.Debug.Precision; p > -15; p-- {
		t *= 10
	}
	return t
}
