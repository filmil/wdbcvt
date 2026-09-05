// SPDX-License-Identifier: Apache-2.0

// Tier 71: the width a real parameter declares.
package main

import (
	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

const sv = `
// Corpus case: %(brief)s
//
// Axis: %(axis)s

` + "`" + `timescale 1ns / 1ps
%(pre)s
module tb;
    logic s = 1'b0;
    %(decl)s

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
%(extra)s`

const vhdl = `
--! @file
--! @brief Corpus case: %(brief)s
--!
--! Axis: %(axis)s

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
    generic (
        r : real := 1.5
    );
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
`

type opt struct {
	sigs     []*c.Obj
	trs      []*c.Obj
	generics []*c.Obj
	// pre is text before the module, for a package, and extra text
	// after it, for a child module.
	pre   string
	extra string
}

const axisTail = " beside a logic, to see where the 16 bits a real parameter declares comes from, when a real variable declares 32 and both hold one float64."

func kase(name, brief, decl, differs string, o opt) {
	axis := "the width of a real parameter. " + brief + axisTail
	body := c.Fill(sv, "brief", brief, "axis", axis, "decl", decl, "pre", o.pre, "extra", o.extra)
	var extra *c.Obj
	if o.generics != nil {
		extra = c.O("generics", o.generics)
	}
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: "tb.sv", Body: body}},
		Signals: append([]*c.Obj{c.Sig("tb", "s", "logic", 1)}, o.sigs...),
		Trs:     append([]*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}, o.trs...),
		Extra:   extra, NoX: true})
}

// real is a parameter of the truth whose value is a real, which the
// decoder spells as the float itself.
func realPrm(scope, name, value string) *c.Obj {
	return c.O("instance", "", "scope", scope, "name", name, "type", "real",
		"declared", "parameter", "value", value)
}

func main() {
	kase("t71_rlw_sreal_p_", "a shortreal parameter", "parameter shortreal R = 1.5;", "t25_v_prm_real__",
		opt{generics: []*c.Obj{realPrm("tb", "R", "1.5")}})
	kase("t71_rlw_rtime_p_", "a realtime parameter", "parameter realtime R = 1.5;", "t71_rlw_sreal_p_",
		opt{generics: []*c.Obj{realPrm("tb", "R", "1.5").With("type", "realtime")}})
	kase("t71_rlw_untyped_", "an untyped parameter from a real literal", "parameter R = 1.5;", "t71_rlw_sreal_p_",
		opt{generics: []*c.Obj{realPrm("tb", "R", "1.5")}})
	kase("t71_rlw_kid_prm_", "a real parameter of a child module", "kid u();", "t71_rlw_sreal_p_",
		opt{generics: []*c.Obj{realPrm("tb.u", "R", "1.5")}, extra: `
module kid;
    parameter real R = 1.5;
endmodule
`})
	kase("t71_rlw_specprm_", "a real specparam",
		"specify\n        specparam d = 1.5;\n    endspecify", "t71_rlw_sreal_p_",
		opt{generics: []*c.Obj{realPrm("tb", "d", "1.5")}})
	kase("t71_rlw_pkg_prm_", "a real parameter of a package",
		"import p::*;\n    real v = p::R;", "t71_rlw_sreal_p_",
		opt{sigs: []*c.Obj{c.Sig("tb", "v", "real", 64)},
			trs: []*c.Obj{c.Tr(0, "v", "1.5")},
			// A package parameter has an object and no record, as
			// t13_sv_pkg measured.
			generics: []*c.Obj{realPrm("p", "R", "1.5").With("logged", false)},
			pre: `
package p;
    parameter real R = 1.5;
endpackage
`})
	kase("t71_rlw_arr_prm_", "an array of two real parameters",
		"parameter real R [0:1] = '{1.5, 2.5};\n    real v = R[1];", "t71_rlw_sreal_p_",
		opt{sigs: []*c.Obj{c.Sig("tb", "v", "real", 64)},
			trs:      []*c.Obj{c.Tr(0, "v", "2.5")},
			generics: []*c.Obj{c.O("instance", "", "scope", "tb", "name", "R", "type", "",
				"declared", "parameter", "value", "(1.5, 2.5)")}})
	// A VHDL real generic, whose sizes the file counts in bytes, as the
	// other side of the same measurement.
	axis := "the width of a real parameter. a VHDL real generic" + axisTail
	c.Emit(c.Case{Name: "t71_rlw_vhdl_gen", Axis: axis, Differs: "t71_rlw_sreal_p_",
		Files:   []c.File{{Name: "tb.ent.vhdl", Body: c.Fill(vhdl, "brief", "a VHDL real generic", "axis", axis)}},
		Signals: []*c.Obj{c.Sig("tb", "s", "std_ulogic", 1)},
		Trs:     []*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")},
		Extra:   c.O("generics", []*c.Obj{realPrm("tb", "r", "1.5")}), NoX: true})
}
