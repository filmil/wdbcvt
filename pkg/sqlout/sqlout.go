// SPDX-License-Identifier: Apache-2.0

// Package sqlout writes a decoded waveform database as an SQLite file,
// in the schema `github.com/filmil/go-vcd-parser` writes from a VCD, so
// that a query written against one reads the other. What this package
// decides is the mapping: which row each object becomes, how a record
// or an array is flattened into leaves, and how a value is spelled.
// See docs/sqlite-output.md.
package sqlout

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"git.hdlfactory.com/HDL/wdbcvt/pkg/wdb"
	_ "github.com/mattn/go-sqlite3"
)

// varKind is the type code of a row of Signals. The values are
// go-vcd-parser's vcd.VarKindCode, which is the order the VCD keywords
// are declared in there; a reader of the database compares against
// those numbers, so they are written out here rather than taken from
// the package, which does not export the table.
type varKind int

const (
	kindEvent varKind = iota
	kindInteger
	kindParameter
	kindReal
	kindReg
	kindSupply0
	kindSupply1
	kindTime
	kindTri
	kindTriand
	kindTrior
	kindTrireg
	kindTri0
	kindTri1
	kindWand
	kindWire
	kindWor
	kindLogic
	kindString
	kindUnknown
)

// schema is the DDL of go-vcd-parser's db package, with two changes.
//
// The columns are TEXT where that package writes STRING. SQLite gives
// a column whose declared type is neither TEXT nor INT numeric
// affinity, so a STRING column turns the value of a vector into a
// number: `00000001` is stored as 1, and a 22 bit value as the double
// 1.11e+21, which is the value destroyed. TEXT keeps what was written,
// and the table and column names, which is what a query names, do not
// move.
//
// The Meta table is new: the timestamps are in the file's own unit and
// nothing in the two original tables says what that unit is.
const schema = `
        CREATE TABLE
            Signals(
                Name TEXT PRIMARY KEY,
                Type INTEGER NOT NULL,
                Code TEXT NOT NULL,
                Size INTEGER NOT NULL
            );

        CREATE INDEX
            SignalsByCode
        ON
            Signals(Code, Name);

        CREATE TABLE
            Svalues(
                Id INTEGER PRIMARY KEY AUTOINCREMENT,
                Timestamp INTEGER NOT NULL,
                Code TEXT NOT NULL,
                Value TEXT NOT NULL,
                FOREIGN KEY(Code) REFERENCES Signals(Code)
            );

        CREATE INDEX
            SvaluesByCodeAndTimestamp
        ON
            Svalues(Code, Timestamp, Value);

        CREATE TABLE
            Meta(
                Key TEXT PRIMARY KEY,
                Value TEXT NOT NULL
            );
        `

// maxTx is how many value rows go into one transaction.
var maxTx = 500000

// leaf is one row of Signals and the last value written for it.
type leaf struct {
	code string
	kind varKind
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
	// carries the values, so this one's changes are not written again.
	alias bool
}

type writer struct {
	f    *wdb.File
	db   *sql.DB
	objs []*object
	// shared indexes the codes already given to a handle, so a second
	// object on it becomes an alias of the first, as a VCD writes one
	// identifier code for both.
	shared map[sharedKey][]string
	// scope is the path of the object being declared.
	scope []string
	sigs  *sql.Stmt
	next  int
	// used counts the names already written, because Name is the
	// primary key of Signals and a design can declare the same path
	// twice: //hdl/counter has four loop indexes named `i` in one
	// process, each on its own handle.
	used map[string]int
}

type sharedKey struct {
	handle uint64
	offset uint32
	typ    int
}

// Write converts f into an SQLite database at path.
func Write(f *wdb.File, path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("%s: %w", path, err)
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	defer db.Close()
	// The file is written once and read afterwards, so the rollback
	// journal buys nothing and costs a fsync per transaction.
	if _, err := db.Exec(`PRAGMA journal_mode=OFF; PRAGMA synchronous=OFF;`); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("%s: create schema: %w", path, err)
	}
	o := &writer{f: f, db: db, objs: make([]*object, len(f.Objects)),
		shared: map[sharedKey][]string{}, scope: []string{"/"},
		used: map[string]int{}}
	if err := o.meta(); err != nil {
		return err
	}
	if err := o.signals(); err != nil {
		return err
	}
	return o.values()
}

// meta records what the two original tables have nowhere to put: the
// unit the timestamps count in, and where the rows came from.
func (o *writer) meta() error {
	e := int(o.f.Debug.Precision)
	rows := [][2]string{
		{"generator", "wdbcvt"},
		{"source", "Vivado xsim waveform database"},
		{"timescale", timescale(e)},
		{"timescale_seconds", fmt.Sprintf("1e%d", e)},
		{"time_exponent", strconv.Itoa(e)},
		{"end_time", strconv.FormatUint(o.f.Header.EndTime, 10)},
	}
	for _, r := range rows {
		if _, err := o.db.Exec(`INSERT INTO Meta(Key, Value) VALUES (?, ?);`, r[0], r[1]); err != nil {
			return fmt.Errorf("meta %q: %w", r[0], err)
		}
	}
	return nil
}

// timescale spells a power of ten of a second the way a VCD's own
// $timescale does, so that the two databases carry the same text.
func timescale(exp int) string {
	units := map[int]string{0: "s", -3: "ms", -6: "us", -9: "ns", -12: "ps", -15: "fs"}
	for e := exp; e > exp-3 && e <= 0; e-- {
		if u, ok := units[e]; ok {
			return fmt.Sprintf("%d%s", pow10(exp-e), u)
		}
	}
	return fmt.Sprintf("1e%d s", exp)
}

// pow10 is ten to the n for the 0, 1 and 2 a timescale needs.
func pow10(n int) int {
	p := 1
	for i := 0; i < n; i++ {
		p *= 10
	}
	return p
}

// signals declares one row per leaf of every logged object, walking the
// instance tree so that the names come out in hierarchy order.
func (o *writer) signals() error {
	tx, err := o.db.Begin()
	if err != nil {
		return err
	}
	o.sigs, err = tx.Prepare(`INSERT INTO Signals(Name, Type, Code, Size) VALUES (?, ?, ?, ?);`)
	if err != nil {
		return err
	}
	byScope := make([][]int, len(o.f.Scopes))
	for i, ob := range o.f.Objects {
		if ob.Logged {
			byScope[ob.Scope] = append(byScope[ob.Scope], i)
		}
	}
	for _, root := range o.f.Scopes[0].Children() {
		if err := o.scopeSignals(root, byScope); err != nil {
			return err
		}
	}
	if err := o.sigs.Close(); err != nil {
		return err
	}
	return tx.Commit()
}

func (o *writer) scopeSignals(s int, byScope [][]int) error {
	o.push(plainName(o.f.Scopes[s].Name))
	defer o.pop()
	for _, i := range byScope[s] {
		if err := o.object(i); err != nil {
			return err
		}
	}
	for _, c := range o.f.Scopes[s].Children() {
		if err := o.scopeSignals(c, byScope); err != nil {
			return err
		}
	}
	return nil
}

func (o *writer) push(name string) { o.scope = append(o.scope, name) }
func (o *writer) pop()             { o.scope = o.scope[:len(o.scope)-1] }

// object declares the rows of one object: one for a scalar or a
// vector, and one per leaf of a record or an array of anything else.
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
		codes := make([]string, len(obj.leaves))
		for j, l := range obj.leaves {
			codes[j] = l.code
		}
		o.shared[key] = codes
	}
	return nil
}

// declare walks a value's shape and writes one row per leaf. A record
// and an array of anything but bits become another level of the name,
// so a field or an element is a signal of its own, as it is in the FST
// output. alias, when set, names the codes an earlier object on the
// same handle already took.
func (o *writer) declare(name string, v wdb.Value, path []int, obj *object, alias []string) error {
	ty := o.f.Base(v.Type)
	switch {
	case vector(o.f, ty):
		return o.leaf(name, kindWire, len(v.Elems), path, obj, alias)
	case ty.Kind == wdb.KindArray && charEnum(o.f, ty):
		return o.leaf(name, kindString, 1, path, obj, alias)
	case ty.Kind == wdb.KindArray:
		o.push(name)
		defer o.pop()
		for i := range v.Elems {
			p := append(append([]int(nil), path...), i)
			if err := o.declare(strconv.Itoa(i), v.Elems[i], p, obj, alias); err != nil {
				return err
			}
		}
	case ty.Kind == wdb.KindRecord:
		o.push(name)
		defer o.pop()
		for i := range v.Fields {
			p := append(append([]int(nil), path...), i)
			if err := o.declare(ty.Fields[i].Name, v.Fields[i], p, obj, alias); err != nil {
				return err
			}
		}
	default:
		kind, bits := scalarKind(o.f, ty)
		return o.leaf(name, kind, bits, path, obj, alias)
	}
	return nil
}

// leaf writes one row of Signals, or takes the code of the aliased one.
func (o *writer) leaf(name string, kind varKind, bits int, path []int, obj *object, alias []string) error {
	code := ""
	if alias != nil && len(obj.leaves) < len(alias) {
		code = alias[len(obj.leaves)]
	} else {
		code = idCode(o.next)
		o.next++
	}
	full := o.unique(strings.Join(append(append([]string(nil), o.scope...), name), "/"))
	if _, err := o.sigs.Exec(full, int(kind), code, bits); err != nil {
		return fmt.Errorf("signal %q: %w", full, err)
	}
	obj.leaves = append(obj.leaves, &leaf{code: code, kind: kind, path: append([]int(nil), path...)})
	return nil
}

// unique makes a name the primary key of Signals will take. A path
// declared more than once, as the four loop indexes named `i` of
// //hdl/counter are, keeps the plain name on its first row and gets
// `#2`, `#3` and so on after it. A VCD has no such trouble because it
// names a signal by its code and lets two `$var` lines agree.
func (o *writer) unique(name string) string {
	o.used[name]++
	if n := o.used[name]; n > 1 {
		return fmt.Sprintf("%s#%d", name, n)
	}
	return name
}

// idCode names a signal the way a VCD does, with the printable ASCII
// characters from `!` upward, and more than one of them once they run
// out.
func idCode(n int) string {
	const lo, hi = '!', '~'
	const base = hi - lo + 1
	var b []byte
	for {
		b = append(b, byte(lo+rune(n%base)))
		n = n/base - 1
		if n < 0 {
			break
		}
	}
	return string(b)
}

// values streams every change of the file into rows of Svalues.
func (o *writer) values() error {
	tx, err := o.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO Svalues(Timestamp, Code, Value) VALUES (?, ?, ?);`)
	if err != nil {
		return err
	}
	n := 0
	err = o.f.Stream(func(i int, t uint64, data []byte) error {
		obj := o.objs[i]
		if obj == nil || obj.alias {
			return nil
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
			s := spell(l.kind, at)
			if l.set && s == l.last {
				continue
			}
			l.last, l.set = s, true
			if _, err := stmt.Exec(int64(t), l.code, s); err != nil {
				return fmt.Errorf("value of %q at %d: %w", l.code, t, err)
			}
			n++
			if n%maxTx == 0 {
				if err := stmt.Close(); err != nil {
					return err
				}
				if err := tx.Commit(); err != nil {
					return err
				}
				if tx, err = o.db.Begin(); err != nil {
					return err
				}
				if stmt, err = tx.Prepare(`INSERT INTO Svalues(Timestamp, Code, Value) VALUES (?, ?, ?);`); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := stmt.Close(); err != nil {
		return err
	}
	return tx.Commit()
}

// spell renders one leaf value as the text the row holds: the
// characters of the logic for a wire, the decimal of an integer or a
// real, and the literal itself for everything else. A VCD strips the
// leading zeroes of a vector and this does not, because the row has
// the width beside it and a fixed width value compares directly.
func spell(kind varKind, v wdb.Value) string {
	switch kind {
	case kindWire:
		if len(v.Elems) > 0 {
			var b strings.Builder
			for _, e := range v.Elems {
				b.WriteString(bitChar(e.Scalar))
			}
			return b.String()
		}
		return bitChar(v.Scalar)
	case kindString:
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
	default:
		return v.Scalar
	}
}

// bitChar is the VCD spelling of one logic value, in lower case.
func bitChar(s string) string {
	if s == "" {
		return "x"
	}
	return strings.ToLower(s[:1])
}

// plainName strips the bars and backslashes an extended identifier
// carries in the database: a generate iteration is `\g(0)\` in VHDL
// and a variable declared in a generate block is `\g[0].r ` in
// Verilog, and neither decoration belongs in a name a query matches on.
func plainName(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, `\`, ""))
}

// vector reports whether a type is an array of logic values, which one
// row holds as that many characters.
func vector(f *wdb.File, ty *wdb.Type) bool {
	return ty.Kind == wdb.KindArray && bitLike(f.Base(ty.Elem))
}

// bitLike reports whether an enumeration is a logic type: the two
// literals of BIT, the nine of STD_ULOGIC, or the four of Verilog
// logic, which the class word tells apart; see docs/format/types.md.
func bitLike(ty *wdb.Type) bool {
	return ty.Kind == wdb.KindEnum && (ty.Class == 1 || ty.Class == 2 || ty.Class == 3)
}

// charEnum reports whether a type is an array of characters, which one
// row holds as the string itself.
func charEnum(f *wdb.File, ty *wdb.Type) bool {
	e := f.Base(ty.Elem)
	return e.Kind == wdb.KindEnum && e.Class == 4
}

// scalarKind maps a scalar type onto a row's type code and width.
// Anything the VCD kinds have no name for is carried as a string,
// which keeps the literal a reader shows: a VHDL enumeration is its
// own literal, and a physical value is a count and its unit.
func scalarKind(f *wdb.File, ty *wdb.Type) (varKind, int) {
	switch {
	case bitLike(ty):
		return kindWire, 1
	case ty.Kind == wdb.KindInteger:
		return kindInteger, 32
	case ty.Kind == wdb.KindReal:
		return kindReal, 64
	case ty.Kind == wdb.KindPhysical:
		return kindTime, 1
	default:
		return kindString, 1
	}
}
