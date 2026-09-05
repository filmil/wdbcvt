// SPDX-License-Identifier: Apache-2.0

// Package gent60 writes tier 60, SystemVerilog under -debug all. Tier 61
// builds its cases the same way, and tier 62 reuses the design, so the
// tier is a library as well as a program.
package gent60

import (
	"strings"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

// SV is the design: a logic, whatever the case declares beside it, and a
// write to each at 50 ns.
const SV = `
// Corpus case: %(brief)s
//
// Axis: %(axis)s

` + "`" + `timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    %(decl)s

    initial begin
        #50 s = 1'b1;
        %(write)s
        #50 $finish;
    end
endmodule
`

const tcl = `open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/*]
log_wave -recursive *
%s
close_vcd
exit
`

// Opt is what a case declares beside the logic: its signals and records,
// the keys its truth carries beyond the usual ones, an xsim script, and
// whether it runs under -debug all, which every case of the tier does
// except where it says otherwise.
type Opt struct {
	Signals []*c.Obj
	Trs     []*c.Obj
	Extra   *c.Obj
	NoDbg   bool
	Tcl     []string
}

// Plus appends the signals and records of another option set, for a case
// that declares two objects.
func (o Opt) Plus(p Opt) Opt {
	o.Signals = append(append([]*c.Obj{}, o.Signals...), p.Signals...)
	o.Trs = append(append([]*c.Obj{}, o.Trs...), p.Trs...)
	return o
}

// WithTcl runs the case under an xsim script.
func (o Opt) WithTcl(lines ...string) Opt {
	o.Tcl = lines
	return o
}

// Case writes one case of the tier.
func Case(name, brief, decl, write, differs string, o Opt) {
	axis := "debugging. " + brief + " beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section."
	body := c.Fill(SV, "brief", brief, "axis", axis, "decl", decl, "write", write)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: "tb.sv", Body: body}},
		Signals: append([]*c.Obj{c.Sig("tb", "s", "logic", 1)}, o.Signals...),
		Trs:     append([]*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}, o.Trs...),
		Extra:   o.Extra, NoX: true})
	attrs := ""
	if o.Tcl != nil {
		c.WriteFile(name, "xsim.tcl", strings.Replace(tcl, "%s", strings.Join(o.Tcl, "\n"), 1))
		attrs += "    tcl = \"xsim.tcl\",\n"
	}
	if !o.NoDbg {
		attrs += "    xelab_args = [\n        \"-debug\",\n        \"all\",\n    ],\n"
	}
	c.PatchBuild(name, func(b string) string {
		return strings.Replace(b, "    ],\n)\n", "    ],\n"+attrs+")\n", 1)
	})
}

// Held declares an object that is logged with one 32 bit record of zeros
// at time 0 and nothing at its write, which is what a string or a class
// handle does under -debug all.
func Held(name, typ string) Opt {
	return Opt{
		Signals: []*c.Obj{c.Sig("tb", name, typ, 32)},
		Trs:     []*c.Obj{c.Tr(0, name, strings.Repeat("0", 32))},
	}
}

// Unlogged declares an object that gets a declaration and a handle but
// is never logged, which is what a queue, a dynamic array or an
// associative array does.
func Unlogged(name string) Opt {
	return Opt{Signals: []*c.Obj{c.Sig("tb", name, "", 32, "logged", false)}}
}

// Main writes the tier.
func Main() {
	Case("t60_dbg_none____", "nothing", "", "", "t11_sv_bit______", Opt{})
	Case("t60_dbg_vec_____", "a 4 bit logic vector", "logic [3:0] v = 4'b0000;", "v = 4'b0101;", "t60_dbg_none____",
		Opt{Signals: []*c.Obj{c.Sig("tb", "v", "logic", 4)}, Trs: []*c.Obj{c.Tr(0, "v", "0000"), c.Tr(50, "v", "0101")}})
	Case("t60_dbg_int_____", "an int", "int i = 7;", "i = 9;", "t60_dbg_none____",
		Opt{Signals: []*c.Obj{c.Sig("tb", "i", "int", 32)}, Trs: []*c.Obj{c.Tr(0, "i", "7"), c.Tr(50, "i", "9")}})
	Case("t60_dbg_real____", "a real", "real r = 0.5;", "r = 1.5;", "t60_dbg_none____",
		Opt{Signals: []*c.Obj{c.Sig("tb", "r", "real", 64)}, Trs: []*c.Obj{c.Tr(0, "r", "0.5"), c.Tr(50, "r", "1.5")}})
	Case("t60_dbg_str_____", "a string variable", `string str = "ab";`, `str = "xyz";`, "t60_dbg_none____", Held("str", "string"))
	Case("t60_dbg_queue___", "a queue", "int q[$];", "q.push_back(5);", "t60_dbg_int_____", Unlogged("q"))
	Case("t60_dbg_dynarr__", "a dynamic array", "int d[];", "d = new[2]; d[1] = 5;", "t60_dbg_int_____", Unlogged("d"))
	Case("t60_dbg_assoc___", "an associative array keyed by string", "int a[string];", `a["k"] = 5;`, "t60_dbg_int_____", Unlogged("a"))
	Case("t60_dbg_assoc_i_", "an associative array keyed by int", "int a[int];", "a[3] = 5;", "t60_dbg_assoc___", Unlogged("a"))
	Case("t60_dbg_class___", "a class with one int field", "class c_t; int f = 1; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_int_____", Held("h", "c_t"))
	Case("t60_dbg_class_2_", "a class with two fields", "class c_t; int f = 1; logic [3:0] g; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class___", Held("h", "c_t"))
	Case("t60_dbg_class_d_", "a derived class", "class b_t; int f = 1; endclass\n    class c_t extends b_t; int g; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class___", Held("h", "c_t"))
	Case("t60_dbg_class_2h", "two handles of one class", "class c_t; int f = 1; endclass\n    c_t h;\n    c_t h2;", "h = new; h2 = new; h.f = 5;", "t60_dbg_class___",
		Held("h", "c_t").Plus(Held("h2", "c_t")))
	Case("t60_dbg_class_n_", "a class handle constructed at time 0", "class c_t; int f = 1; endclass\n    c_t h = new;", "h.f = 5;", "t60_dbg_class___", Held("h", "c_t"))
	// set_value /tb/str cd after run 25 ns ends the batch script without
	// a message, and the database closes at 25 ns; so the string is named
	// in log_wave instead.
	Case("t60_dbg_str_log_", "a string named in log_wave", `string str = "ab";`, `str = "xyz";`, "t60_dbg_str_____",
		Opt{Signals: []*c.Obj{c.Sig("tb", "str", "string", 32, "records", 2)}, Trs: Held("str", "string").Trs}.
			WithTcl("log_wave /tb/str", "run -all"))
	Case("t60_dbg_q_log___", "a queue named in log_wave", "int q[$];", "q.push_back(5);", "t60_dbg_queue___",
		Unlogged("q").WithTcl("log_wave /tb/q", "run -all"))
	Case("t60_dbg_struct__", "a packed struct", "typedef struct packed { logic a; logic [2:0] b; } st_t;\n    st_t st = '0;", "st = '{1, 3'b011};", "t60_dbg_vec_____",
		Opt{Signals: []*c.Obj{c.O("scope", "tb", "name", "st", "type", "st_t", "fields", []*c.Obj{
			c.O("name", "a", "width", 1, "type", "logic"), c.O("name", "b", "width", 3, "type", "")})},
			Trs: []*c.Obj{c.Tr(0, "st.a", "0"), c.Tr(0, "st.b", "000"), c.Tr(50, "st.a", "1"), c.Tr(50, "st.b", "011")}})
	Case("t60_dbg_mem_____", "a memory of two 4 bit words", "logic [3:0] m [0:1] = '{0, 0};", "m[1] = 4'd9;", "t60_dbg_vec_____",
		Opt{Signals: []*c.Obj{c.Sig("tb", "m", "memory", 8, "elements", 2, "element_width", 4, "element_type", "logic")},
			Trs: []*c.Obj{c.Tr(0, "m", "(0000, 0000)"), c.Tr(50, "m", "(0000, 1001)")}})
}
