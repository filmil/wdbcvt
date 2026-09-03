// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"bufio"
	"bytes"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/filmil/go-vcd-parser/vcd"
)

// The VCD cross-check reads the sim.vcd that xsim wrote next to every
// sim.wdb of the corpus, through go-vcd-parser, and compares it with
// what the decoder reads from the database. The VCD is the second
// guard of docs/provenance.md: xsim wrote both files from the same
// simulation, and the VCD format is public. Nothing here asserts
// against bytes the decoder produced.
//
// The comparison is strict where the VCD says anything. Every VCD
// variable must name an object of the database, and at every time the
// VCD lists a value, the last value the database holds at that time
// must spell the same. The VCD spelling is the four state one: an
// std_ulogic U, X, W and - become x, L becomes 0, H becomes 1, and a
// vector shorter than its declared size is extended to the left with
// 0 or with its own x or z, as the VCD standard says.
//
// The VCD writer leaves out much of what the database holds; see
// docs/format/vcd.md. vcdOmitted spells out the observed rule, and the
// test fails when an object is missing from the VCD without a listed
// reason, or present in the VCD with one, so that the rule stays as
// exact as the corpus can make it.

// vcdVar is one $var line with the path its $scope stack gives it.
type vcdVar struct {
	path string
	kind string
	size int
	code string
}

// vcdFile is what the test keeps of a parsed VCD: the variables and,
// per identifier code, the value at every time the file lists one.
type vcdFile struct {
	vars    []vcdVar
	changes map[string][]change
	// timescaleFS is the VCD's $timescale in femtoseconds.
	timescaleFS float64
}

// plainIdent matches the identifiers go-vcd-parser 0.1.0 reads. xsim
// also writes VHDL extended identifiers such as `\g(0)\`, Verilog
// escaped ones such as `\g.r `, and generate scope names such as
// `g[0].dut`. hideNames swaps every other identifier of a $scope or
// $var line for a plain one before parsing, and returns the map back.
var plainIdent = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func hideNames(src []byte) ([]byte, map[string]string) {
	names := map[string]string{}
	hide := func(id string) string {
		if plainIdent.MatchString(id) {
			return id
		}
		key := fmt.Sprintf("hidden%d", len(names))
		names[key] = id
		return key
	}
	var out bytes.Buffer
	sc := bufio.NewScanner(bytes.NewReader(src))
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		line := sc.Text()
		fs := strings.Fields(line)
		switch {
		case len(fs) == 4 && fs[0] == "$scope":
			fs[2] = hide(fs[2])
			line = strings.Join(fs, " ")
		case len(fs) >= 6 && fs[0] == "$var":
			fs[4] = hide(fs[4])
			line = strings.Join(fs, " ")
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return out.Bytes(), names
}

func loadVCD(t *testing.T, path string) *vcdFile {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	src, names := hideNames(raw)
	ast, err := vcd.NewParser[vcd.File]().Parse(path, bytes.NewReader(src))
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	unhide := func(id string) string {
		if orig, ok := names[id]; ok {
			return orig
		}
		return id
	}
	out := &vcdFile{changes: map[string][]change{}}
	var scope []string
	for _, d := range ast.DeclarationCommand {
		switch {
		case d.Scope != nil:
			scope = append(scope, unhide(d.Scope.Id))
		case d.Upscope != nil:
			scope = scope[:len(scope)-1]
		case d.Var != nil:
			v := d.Var
			out.vars = append(out.vars, vcdVar{
				path: plainPath(strings.Join(append(append([]string{}, scope...), unhide(v.Id.Name)), ".")),
				kind: v.VarType,
				size: v.Size,
				code: v.Code,
			})
		case d.Timescale != nil:
			// The VCD is written in the simulation precision, which
			// is the database's time unit as well, so the test
			// compares the two times as they are: tier 21.
			out.timescaleFS = d.Timescale.AsSeconds() * 1e15
		}
	}
	var now uint64
	add := func(vc *vcd.ValueChangeT) {
		code := vc.GetIdCode()
		out.changes[code] = append(out.changes[code], change{now, vc.GetValue()})
	}
	for _, s := range ast.SimulationCommand {
		switch {
		case s.SimulationTime != nil:
			now = s.SimulationTime.Value()
		case s.Dumpvars != nil:
			for _, vc := range s.Dumpvars.ValueChange {
				add(vc)
			}
		case s.ValueChange != nil:
			add(s.ValueChange)
		}
	}
	return out
}

// vcdExtend pads a VCD vector value to its declared size, the way the
// standard says a reader must: with 0 when the leftmost bit is 0 or
// 1, and with the leftmost bit itself when it is x or z.
func vcdExtend(v string, size int) string {
	if len(v) >= size || len(v) == 0 {
		return v
	}
	pad := "0"
	if v[0] == 'x' || v[0] == 'z' {
		pad = v[:1]
	}
	return strings.Repeat(pad, size-len(v)) + v
}

// vcdSpell renders a decoded value the way the VCD writer spells it.
// A real comes back as its decimal text with an r in front, the way
// the VCD writes it; everything else is a string of 0, 1, x and z.
func (f *File) vcdSpell(v Value) string {
	ty := &f.Types[f.resolve(v.Type)]
	switch ty.Kind {
	case KindRecord:
		// A packed struct is its fields' bits, first field on the
		// left. An unpacked struct puts each field in a slot of whole
		// 32 bit words: t11_sv_ustruct, t11_sv_struct3 and
		// t11_sv_struct40. A real field inside one does not spell its
		// value at all, t11_sv_struct_r, and comes back as 32 question
		// marks so that the comparison fails on it. A packed union
		// is its widest field, since the fields share the bits:
		// t24_sv_union____.
		var b strings.Builder
		for _, fv := range v.Fields {
			s := f.vcdSpell(fv)
			if ty.Layout == LayoutUnion {
				if len(s) > b.Len() {
					b.Reset()
					b.WriteString(s)
				}
				continue
			}
			if ty.Layout == LayoutUnpacked {
				if strings.HasPrefix(s, "r") {
					s = strings.Repeat("?", 32)
				}
				s = vcdExtend(s, (len(s)+31)/32*32)
			}
			b.WriteString(s)
		}
		return b.String()
	case KindValues:
		// A SystemVerilog enum is written as the bits of its value.
		for _, ev := range ty.Values {
			if ev.Name == v.Scalar {
				return vcdExtend(strconv.FormatUint(ev.Value, 2), f.bitWidth(ty.Elem))
			}
		}
	case KindArray:
		if v.Scalar == "" {
			var b strings.Builder
			for _, e := range v.Elems {
				b.WriteString(f.vcdSpell(e))
			}
			return b.String()
		}
		// A predefined Verilog integral type is spelled in decimal by
		// the decoder; the VCD writes its bits.
		if n, ok := new(big.Int).SetString(v.Scalar, 10); ok {
			w := f.bitWidth(v.Type)
			if n.Sign() < 0 {
				n.Add(n, new(big.Int).Lsh(big.NewInt(1), uint(w)))
			}
			return vcdExtend(n.Text(2), w)
		}
	case KindReal:
		return "r" + v.Scalar
	}
	var b strings.Builder
	for _, c := range strings.Trim(v.Scalar, "'") {
		switch c {
		case 'U', 'X', 'W', '-':
			b.WriteByte('x')
		case 'L':
			b.WriteByte('0')
		case 'H':
			b.WriteByte('1')
		case 'Z':
			b.WriteByte('z')
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

// bitWidth is the number of bits of a vector type: the product of its
// range lengths.
func (f *File) bitWidth(t int) int {
	ty := &f.Types[f.resolve(t)]
	w := 1
	for _, r := range ty.Ranges {
		w *= r.Length()
	}
	return w
}

// sameVCDValue compares one VCD value with one spelled database value,
// reals by their number, everything else bit by bit after extension.
func sameVCDValue(kind string, size int, got, want string) bool {
	if strings.HasPrefix(want, "r") {
		a, err1 := strconv.ParseFloat(got, 64)
		b, err2 := strconv.ParseFloat(want[1:], 64)
		return err1 == nil && err2 == nil && a == b
	}
	// An untyped Verilog parameter is declared with size 0; both sides
	// then extend to the longer of the two.
	if size < len(want) {
		size = len(want)
	}
	return vcdExtend(got, size) == vcdExtend(want, size)
}

// vcdOmitted names the reason the VCD writer leaves an object out, or
// returns "" for an object the VCD carries. It is the observed rule
// over the corpus; see docs/format/vcd.md.
func (f *File) vcdOmitted(o Object) string {
	dc := f.Decls[o.Decl]
	if !strings.HasPrefix(plainPath(f.ObjectPath(o)), "tb.") {
		return "outside the tb hierarchy that the script logs"
	}
	ty := &f.Types[f.resolve(dc.Type)]
	if f.verilog(dc.Type) {
		if ty.Kind == KindArray && ty.Layout == LayoutUnpacked && f.Types[dc.Type].Kind != KindAlias {
			return "an unpacked array not declared through a typedef"
		}
		return ""
	}
	switch dc.Kind {
	case DeclGeneric:
		return "a VHDL generic"
	case DeclConstant:
		return "a VHDL constant"
	}
	// BIT and STD_ULOGIC are told by their literals, so that a subtype
	// entry named STD_LOGIC, t8_port_inout, counts too.
	bitLike := func(t int) bool {
		e := &f.Types[f.resolve(t)]
		if e.Kind != KindEnum {
			return false
		}
		lits := strings.Join(e.Literals, "")
		return lits == "'0''1'" || lits == "'U''X''0''1''Z''W''L''H''-'"
	}
	if bitLike(dc.Type) || ty.Kind == KindArray && bitLike(ty.Elem) {
		return ""
	}
	return "a VHDL type other than BIT, STD_ULOGIC and their vectors: " + ty.Kind.String() + " " + ty.Name
}

// vcdDeviations lists the objects whose VCD values do not spell what
// the database holds, by case and path, with the reason. The test
// expects the mismatch, so that a fix in a later Vivado shows up.
var vcdDeviations = map[string]string{
	"t11_sv_struct_r_ tb.s": "a real field of an unpacked struct is written as a 32 bit slot of bits that are not the value",
}

// designs lists the directories of the designs outside the corpus
// that have a database and a VCD but no truth.json. The VCD is their
// only check. //hdl/counter:sim is a record port with per field
// assignments, and its ctl record found the VHDL partial write.
var designs = []string{"hdl/counter", "hdl/uart"}

func TestVCD(t *testing.T) {
	dirs := corpusCases(t)
	root := filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"))
	for _, d := range designs {
		dirs = append(dirs, filepath.Join(root, d))
	}
	for _, dir := range dirs {
		dir := dir
		t.Run(filepath.Base(dir), func(t *testing.T) {
			f, err := ReadFile(filepath.Join(dir, "sim.wdb"))
			if err != nil {
				t.Fatal(err)
			}
			v := loadVCD(t, filepath.Join(dir, "sim.vcd"))
			if unit := float64(f.TimeFS(1)); v.timescaleFS < unit*0.999 || v.timescaleFS > unit*1.001 {
				t.Fatalf("VCD timescale %g fs, the database counts in %s", v.timescaleFS, f.TimeUnit())
			}

			objByPath := map[string]Object{}
			for _, o := range f.Objects {
				objByPath[plainPath(f.ObjectPath(o))] = o
			}
			seen := map[string]bool{}
			for _, vv := range v.vars {
				o, ok := objByPath[vv.path]
				if !ok {
					t.Errorf("VCD %s %q has no object in the database", vv.kind, vv.path)
					continue
				}
				seen[vv.path] = true
				dc := f.Decls[o.Decl]
				// The VCD names a net by its kind, wand for a wand, and
				// the declaration kind word says the same: t19_v_wand.
				// A wire and a uwire are both wire.
				if dc.Kind.IsNet() && vv.kind != dc.Kind.String() && !(dc.Kind == DeclNet && vv.kind == "wire") {
					t.Errorf("%s: VCD kind %s, declaration kind %s", vv.path, vv.kind, dc.Kind)
				}
				ch, err := f.Changes(o)
				if err != nil {
					t.Fatal(err)
				}
				var want []change
				for _, c := range ch {
					val, err := f.Decode(dc, c.Data)
					if err != nil {
						t.Fatalf("%s at %d %s: %v", vv.path, c.Time, f.TimeUnit(), err)
					}
					want = append(want, change{c.Time, f.vcdSpell(val)})
				}
				want = finalPerTime(want)
				got := v.changes[vv.code]
				if reason := f.vcdOmitted(o); reason != "" {
					t.Errorf("%s is in the VCD but vcdOmitted says it is %s", vv.path, reason)
				}
				deviation := vcdDeviations[filepath.Base(dir)+" "+vv.path]
				// The VCD spells an untyped time parameter as the 4-state
				// reading of its float64 storage, with z bits where the
				// value has none: t28_sv_prm_time_, and every t30 ptm case.
				if deviation == "" && f.timeParam(dc) {
					deviation = "an untyped time parameter, written as the bits of its float64 storage"
				}
				mismatch := false
				if len(got) != len(want) {
					mismatch = true
					t.Errorf("%s: VCD lists %d times, the database %d", vv.path, len(got), len(want))
				}
				for i := 0; i < len(got) && i < len(want); i++ {
					if got[i].time != want[i].time {
						mismatch = true
						t.Errorf("%s: VCD change %d at %d, the database at %d %s", vv.path, i, got[i].time, want[i].time, f.TimeUnit())
						break
					}
					if !sameVCDValue(vv.kind, vv.size, got[i].value, want[i].value) {
						mismatch = true
						if deviation == "" {
							t.Errorf("%s at %d %s: VCD %s, the database %s", vv.path, got[i].time, f.TimeUnit(), got[i].value, want[i].value)
						}
					}
				}
				if deviation != "" && !mismatch {
					t.Errorf("%s matches its VCD; vcdDeviations lists it as %s", vv.path, deviation)
				}
			}

			// Two VCD variables with one identifier code are one signal
			// seen from two scopes. The database gives them one handle
			// and one offset: t12_v_port_wire, t13_sv_iface.
			byCode := map[string]vcdVar{}
			for _, vv := range v.vars {
				o := objByPath[vv.path]
				first, ok := byCode[vv.code]
				if !ok {
					byCode[vv.code] = vv
					continue
				}
				fo := objByPath[first.path]
				if o.Handle != fo.Handle || o.Offset != fo.Offset {
					t.Errorf("%s and %s share VCD code %q but have handles %#x+%d and %#x+%d",
						first.path, vv.path, vv.code, fo.Handle, fo.Offset, o.Handle, o.Offset)
				}
			}

			// Every logged object the VCD leaves out has a reason.
			for _, o := range f.Objects {
				p := plainPath(f.ObjectPath(o))
				if seen[p] || !o.Logged {
					continue
				}
				if f.vcdOmitted(o) == "" {
					dc := f.Decls[o.Decl]
					t.Errorf("%s is not in the VCD and vcdOmitted has no reason: %s of %s", p, dc.Kind, f.Types[f.resolve(dc.Type)].Name)
				}
			}
		})
	}
}
