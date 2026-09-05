// SPDX-License-Identifier: Apache-2.0

// Tier 67: an enumeration inside a packed struct.
package main

import (
	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

const sv = `
// Corpus case: %(brief)s
//
// Axis: %(axis)s

` + "`" + `timescale 1ns / 1ps

module tb;
    %(decl)s

    initial begin
        #50 %(drive)s;
        #50 $finish;
    end
endmodule
`

func kase(name, brief, decl, drive, differs string, sigs, trs []*c.Obj) {
	axis := "an enumeration in a packed struct. " + brief + ", under typical, to see how wide the enumeration's bits are where a neighbour depends on it."
	body := c.Fill(sv, "brief", brief, "axis", axis, "decl", decl, "drive", drive)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files: []c.File{{Name: "tb.sv", Body: body}}, Signals: sigs, Trs: trs})
}

func field(name string, width int, typ string) *c.Obj {
	return c.Sig("", name, typ, width)
}

func main() {
	// An enum of two bits whose value needs one, between two logics.
	kase("t67_esz_pk_2bit_", "a two bit enumeration between two logics",
		"typedef enum logic [1:0] {E0 = 2'd0, E1 = 2'd1} e2_t;\n"+
			"    typedef struct packed { logic a; e2_t e; logic b; } rec_t;\n"+
			"    rec_t s = '{a: 1'b0, e: E0, b: 1'b0};",
		"s = '{a: 1'b1, e: E1, b: 1'b1}", "t11_sv_struct___",
		[]*c.Obj{c.Sig("tb", "s", "rec_t", 4, "fields", []*c.Obj{
			field("a", 1, "logic"), field("e", 2, "e2_t"), field("b", 1, "logic")})},
		[]*c.Obj{c.Tr(0, "s.a", "0"), c.Tr(0, "s.e", "E0"), c.Tr(0, "s.b", "0"),
			c.Tr(50, "s.a", "1"), c.Tr(50, "s.e", "E1"), c.Tr(50, "s.b", "1")})

	// The same with four bits, so the width of the field is not the
	// width of the value it holds.
	kase("t67_esz_pk_4bit_", "a four bit enumeration between two logics",
		"typedef enum logic [3:0] {E0 = 4'd0, E1 = 4'd1} e4_t;\n"+
			"    typedef struct packed { logic a; e4_t e; logic b; } rec_t;\n"+
			"    rec_t s = '{a: 1'b0, e: E0, b: 1'b0};",
		"s = '{a: 1'b1, e: E1, b: 1'b1}", "t67_esz_pk_2bit_",
		[]*c.Obj{c.Sig("tb", "s", "rec_t", 6, "fields", []*c.Obj{
			field("a", 1, "logic"), field("e", 4, "e4_t"), field("b", 1, "logic")})},
		[]*c.Obj{c.Tr(0, "s.a", "0"), c.Tr(0, "s.e", "E0"), c.Tr(0, "s.b", "0"),
			c.Tr(50, "s.a", "1"), c.Tr(50, "s.e", "E1"), c.Tr(50, "s.b", "1")})

	// An enumeration over a named type rather than a bare vector, which
	// is where the width sits on the base type instead.
	kase("t67_esz_pk_int__", "an enumeration over int in a packed struct",
		"typedef enum int {E0 = 0, E1 = 1} ei_t;\n"+
			"    typedef struct packed { logic a; ei_t e; logic b; } rec_t;\n"+
			"    rec_t s = '{a: 1'b0, e: E0, b: 1'b0};",
		"s = '{a: 1'b1, e: E1, b: 1'b1}", "t67_esz_pk_2bit_",
		[]*c.Obj{c.Sig("tb", "s", "rec_t", 34, "fields", []*c.Obj{
			field("a", 1, "logic"), field("e", 32, "ei_t"), field("b", 1, "logic")})},
		[]*c.Obj{c.Tr(0, "s.a", "0"), c.Tr(0, "s.e", "E0"), c.Tr(0, "s.b", "0"),
			c.Tr(50, "s.a", "1"), c.Tr(50, "s.e", "E1"), c.Tr(50, "s.b", "1")})
}
