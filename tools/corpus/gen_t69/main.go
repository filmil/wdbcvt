// SPDX-License-Identifier: Apache-2.0

// Tier 69: the value class of the declaration forms the sweeps so far
// have not reached.
package main

import (
	"strings"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

// opt is what a case declares beside the logic: the objects it adds and
// their records, a second module or package beside the top, and the
// arguments the case elaborates under.
type opt struct {
	sigs     []*c.Obj
	trs      []*c.Obj
	absent   []*c.Obj
	generics []*c.Obj
	extra    string
	xelab    []string
	// top names the top scope when elaboration renames it, as an
	// override on the command line does.
	top string
}

const sv = `
// Corpus case: %(brief)s
//
// Axis: %(axis)s

` + "`" + `timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    %(decl)s

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
%(extra)s`

func kase(name, brief, decl, differs string, o opt) {
	axis := "value class. " + brief + " beside a logic, to see which value class of region 17 the declaration takes, and whether any form reaches the codes 2 and 5 that no case has produced."
	body := c.Fill(sv, "brief", brief, "axis", axis, "decl", decl, "extra", o.extra)
	extra := c.O()
	if o.absent != nil {
		extra.Set("absent", o.absent)
	}
	if o.generics != nil {
		extra.Set("generics", o.generics)
	}
	if len(extra.Keys()) == 0 {
		extra = nil
	}
	top := o.top
	if top == "" {
		top = "tb"
	}
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: "tb.sv", Body: body}},
		Signals: append([]*c.Obj{c.Sig(top, "s", "logic", 1)}, o.sigs...),
		Trs:     append([]*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}, o.trs...),
		Extra:   extra, NoX: true})
	if o.xelab != nil {
		attrs := "    xelab_args = [\n"
		for _, a := range o.xelab {
			attrs += "        \"" + a + "\",\n"
		}
		attrs += "    ],\n"
		c.PatchBuild(name, func(b string) string {
			return strings.Replace(b, "    ],\n)\n", "    ],\n"+attrs+")\n", 1)
		})
	}
}

// absent names a declaration the source has and the file does not.
func absent(name, typ string) []*c.Obj {
	return []*c.Obj{c.O("scope", "tb", "name", name, "type", typ)}
}

// prm is a parameter of the truth, whose value is 32 bits. scope names
// the scope it sits in, since a parameter of the top itself has no
// instance to hang from.
func prm(scope, name string, v uint64) *c.Obj {
	return c.O("instance", "", "scope", scope, "name", name, "type", "",
		"declared", "parameter", "value", c.Bits(v, 32))
}

func main() {
	// A specify block's constant, which no case has declared.
	kase("t69_vcl_specprm_", "a specparam in a specify block",
		"specify\n        specparam d = 10;\n    endspecify", "t11_sv_logic____",
		opt{generics: []*c.Obj{prm("tb", "d", 10)}})
	// The nets whose value the language fixes rather than a driver.
	kase("t69_vcl_supply0_", "a supply0 net", "supply0 g;", "t11_sv_logic____",
		opt{sigs: []*c.Obj{c.Sig("tb", "g", "supply0", 1)},
			trs: []*c.Obj{c.Tr(0, "g", "X"), c.Tr(0, "g", "0")}})
	kase("t69_vcl_supply1_", "a supply1 net", "supply1 v;", "t69_vcl_supply0_",
		opt{sigs: []*c.Obj{c.Sig("tb", "v", "supply1", 1)},
			trs: []*c.Obj{c.Tr(0, "v", "X"), c.Tr(0, "v", "1")}})
	// Two forms have no case, because xsim rejects them outright:
	// `trireg (large) t;` is `ERROR: [XSIM 43-4096] Trireg is not
	// supported`, and `let five = 5;` is `ERROR: [XSIM 43-3980] The
	// SystemVerilog feature "Let" is not supported yet for
	// simulation`.
	// A const variable, which is elaborated like a parameter but
	// declared like a variable.
	kase("t69_vcl_const_v_", "a const vector", "const logic [7:0] k = 8'd5;", "t11_sv_logic____",
		opt{sigs: []*c.Obj{c.Sig("tb", "k", "logic", 8)},
			trs: []*c.Obj{c.Tr(0, "k", "00000101")}})
	kase("t69_vcl_const_i_", "a const int", "const int k = 5;", "t69_vcl_const_v_",
		opt{sigs: []*c.Obj{c.Sig("tb", "k", "int", 32)},
			trs: []*c.Obj{c.Tr(0, "k", "5")}})
	// A parameter given its value from outside the source.
	kase("t69_vcl_defparam", "a child parameter set by defparam",
		"kid u();\n    defparam u.P = 5;", "t11_sv_logic____", opt{
			sigs:     []*c.Obj{c.Sig("tb.u", "v", "logic", 8)},
			trs:      []*c.Obj{c.Tr(0, "tb.u.v", "00000101")},
			generics: []*c.Obj{prm("tb.u", "P", 5)},
			extra: `
module kid;
    parameter P = 1;
    logic [7:0] v = P;
endmodule
`})
	kase("t69_vcl_gtop_prm", "a parameter overridden on the xelab command line",
		"parameter P = 1;\n    logic [7:0] v = P;", "t11_sv_logic____",
		opt{
			// An override renames the elaborated top after itself.
			top:      "tb(P=5)",
			sigs:     []*c.Obj{c.Sig("tb(P=5)", "v", "logic", 8)},
			trs:      []*c.Obj{c.Tr(0, "tb(P=5).v", "00000101")},
			generics: []*c.Obj{prm("tb(P=5)", "P", 5)},
			xelab:    []string{"-generic_top", "P=5"}})
	// The forms SystemVerilog adds to a parameter.
	kase("t69_vcl_typeprm_", "a type parameter used by a variable",
		"parameter type T = logic [7:0];\n    T v = 5;", "t11_sv_logic____",
		opt{sigs: []*c.Obj{c.Sig("tb", "v", "T", 8)},
			trs: []*c.Obj{c.Tr(0, "v", "00000101")}})
	kase("t69_vcl_bits_prm", "a parameter from $bits",
		"parameter B = $bits(logic [7:0]);\n    logic [7:0] v = B;", "t11_sv_logic____",
		opt{sigs: []*c.Obj{c.Sig("tb", "v", "logic", 8)},
			trs:      []*c.Obj{c.Tr(0, "v", "00001000")},
			generics: []*c.Obj{prm("tb", "B", 8)}})
	// Initializers that convert between the state counts and the
	// number kinds.
	kase("t69_vcl_bit_real", "a two state bit from a real literal", "bit b = 1.5;", "t11_sv_logic____",
		opt{sigs: []*c.Obj{c.Sig("tb", "b", "bit", 1)}, trs: []*c.Obj{c.Tr(0, "b", "0")}})
	kase("t69_vcl_enum_xcs", "an enumeration from a cast of 'x",
		"typedef enum logic [1:0] {A, B} e_t;\n    e_t e = e_t'('x);", "t11_sv_logic____",
		opt{sigs: []*c.Obj{
			c.O("scope", "tb", "name", "xilinx_isim_temp_0_ln12castingOp", "width", 2,
				"type", "e_t", "declared", "hidden, from a cast"),
			c.Sig("tb", "e", "e_t", 2)},
			trs: []*c.Obj{c.Tr(0, "tb.xilinx_isim_temp_0_ln12castingOp", "XX"), c.Tr(0, "e", "XX")}})
	kase("t69_vcl_wire_ini", "a net with an initializer", "wire w = 1'b1;", "t11_sv_logic____",
		opt{sigs: []*c.Obj{c.Sig("tb", "w", "wire", 1)},
			trs: []*c.Obj{c.Tr(0, "w", "X"), c.Tr(0, "w", "1")}})
	// A class handle's cousin, which the corpus has never declared.
	kase("t69_vcl_chandle_", "a chandle", "chandle h;", "t11_sv_logic____",
		opt{absent: absent("h", "chandle")})
}
