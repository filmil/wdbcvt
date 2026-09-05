// SPDX-License-Identifier: Apache-2.0

// Tier 66: process scopes of the SystemVerilog constructs not yet seen.
package main

import (
	"fmt"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

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

// opt holds what a case adds to the plain design: its declarations and
// records, a second module or program beside the top, another end time,
// and, for a case where the logic is not written at all, the whole
// record list in place of the usual two.
type opt struct {
	sigs   []*c.Obj
	trs    []*c.Obj
	extra  string
	end    int
	trsAll []*c.Obj
}

func kase(name, brief, decl, differs string, o opt) {
	axis := "process scopes. " + brief + " beside a logic, under typical, to see what scope the construct leaves and what it declares."
	body := c.Fill(sv, "brief", brief, "axis", axis, "decl", decl, "extra", o.extra)
	trs := o.trsAll
	if trs == nil {
		trs = append([]*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}, o.trs...)
	}
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: "tb.sv", Body: body}},
		Signals: append([]*c.Obj{c.Sig("tb", "s", "logic", 1)}, o.sigs...),
		Trs:     trs, End: o.end, NoX: true})
}

// clock gives the records of a logic toggling every 10 ns for 9 edges.
func clock() []*c.Obj {
	trs := []*c.Obj{c.Tr(0, "c", "0")}
	for i := 0; i < 9; i++ {
		trs = append(trs, c.Tr(10*(i+1), "c", fmt.Sprint((i+1)%2)))
	}
	return trs
}

const kid = `
module kid(input i, output o);
    assign o = i;
%s
endmodule
`

func main() {
	kase("t66_prc_final___", "a final block", "final begin\n        $display(\"done\");\n    end", "t11_sv_logic____", opt{})
	kase("t66_prc_latch___", "an always_latch", "logic q;\n    always_latch if (s) q <= 1'b1;", "t11_sv_logic____",
		opt{sigs: []*c.Obj{c.Sig("tb", "q", "logic", 1)}, trs: []*c.Obj{c.Tr(0, "q", "X"), c.Tr(50, "q", "1")}})
	kase("t66_prc_ass_imm_", "an immediate assertion in an always block", "always @(s) assert (s !== 1'bx);", "t11_sv_logic____", opt{})
	kase("t66_prc_ass_conc", "a concurrent assertion on a clock", "logic c = 1'b0;\n    always #10 c = ~c;\n    assert property (@(posedge c) 1'b1);", "t66_prc_ass_imm_",
		opt{sigs: []*c.Obj{c.Sig("tb", "c", "logic", 1)}, trs: clock()})
	kase("t66_prc_prop____", "a named property and sequence", "logic c = 1'b0;\n    always #10 c = ~c;\n    sequence q; 1'b1; endsequence\n    property p; @(posedge c) q; endproperty\n    assert property (p);", "t66_prc_ass_conc",
		opt{sigs: []*c.Obj{c.Sig("tb", "c", "logic", 1)}, trs: clock()})
	kase("t66_prc_task____", "a task called from an initial block", "task t; #10; endtask\n    initial t();", "t11_sv_logic____", opt{})
	kase("t66_prc_func____", "a function called from an assign", "wire w;\n    function automatic logic f(input logic x); return ~x; endfunction\n    assign w = f(s);", "t66_prc_task____",
		opt{sigs: []*c.Obj{c.Sig("tb", "w", "wire", 1, "records", 3)},
			trs: []*c.Obj{c.Tr(0, "w", "X"), c.Tr(0, "w", "1"), c.Tr(50, "w", "0")}})
	kase("t66_prc_program_", "a program block instantiated beside the module", "prog p();", "t11_sv_logic____",
		opt{end: 10, trsAll: []*c.Obj{c.Tr(0, "s", "0")}, extra: `
program prog;
    initial begin
        #10;
    end
endprogram
`})
	kase("t66_prc_bind____", "a child bound into the module with bind", "wire w; assign w = s;", "t66_prc_func____",
		opt{sigs: []*c.Obj{c.Sig("tb", "w", "wire", 1, "records", 3), c.Sig("tb.b", "i", "wire", 1, "port", "in", "records", 3)},
			trs: []*c.Obj{c.Tr(0, "w", "X"), c.Tr(0, "w", "0"), c.Tr(50, "w", "1"),
				c.Tr(0, "tb.b.i", "X"), c.Tr(0, "tb.b.i", "0"), c.Tr(50, "tb.b.i", "1")},
			extra: `
module watcher(input i);
endmodule

bind tb watcher b(.i(s));
`})
	kase("t66_prc_specify_", "a specify block in a child", "wire w; kid u(.i(s), .o(w));", "t66_prc_func____",
		opt{sigs: []*c.Obj{c.Sig("tb", "w", "wire", 1, "records", 4), c.Sig("tb.u", "i", "wire", 1, "port", "in", "records", 3), c.Sig("tb.u", "o", "wire", 1, "port", "out", "records", 4)},
			trs: []*c.Obj{c.Tr(0, "w", "X"), c.Tr(1, "w", "0"), c.Tr(51, "w", "1"),
				c.Tr(0, "tb.u.i", "X"), c.Tr(0, "tb.u.i", "0"), c.Tr(50, "tb.u.i", "1"),
				c.Tr(0, "tb.u.o", "X"), c.Tr(1, "tb.u.o", "0"), c.Tr(51, "tb.u.o", "1")},
			extra: `
module kid(input i, output o);
    assign o = i;
    specify
        (i => o) = 1;
    endspecify
endmodule
`})
	kase("t66_prc_kid_____", "the same child without a specify block", "wire w; kid u(.i(s), .o(w));", "t66_prc_specify_",
		opt{sigs: []*c.Obj{c.Sig("tb", "w", "wire", 1, "records", 4), c.Sig("tb.u", "i", "wire", 1, "port", "in", "records", 3), c.Sig("tb.u", "o", "wire", 1, "port", "out", "records", 4)},
			trs: []*c.Obj{c.Tr(0, "w", "X"), c.Tr(0, "w", "0"), c.Tr(50, "w", "1"),
				c.Tr(0, "tb.u.i", "X"), c.Tr(0, "tb.u.i", "0"), c.Tr(50, "tb.u.i", "1"),
				c.Tr(0, "tb.u.o", "X"), c.Tr(0, "tb.u.o", "0"), c.Tr(50, "tb.u.o", "1")},
			extra: fmt.Sprintf(kid, "")})
	kase("t66_prc_spec_0__", "the specify path delay set to 0", "wire w; kid u(.i(s), .o(w));", "t66_prc_specify_",
		opt{sigs: []*c.Obj{c.Sig("tb", "w", "wire", 1, "records", 4), c.Sig("tb.u", "i", "wire", 1, "port", "in", "records", 3), c.Sig("tb.u", "o", "wire", 1, "port", "out", "records", 4)},
			trs: []*c.Obj{c.Tr(0, "w", "X"), c.Tr(0, "w", "0"), c.Tr(50, "w", "1"),
				c.Tr(0, "tb.u.i", "X"), c.Tr(0, "tb.u.i", "0"), c.Tr(50, "tb.u.i", "1"),
				c.Tr(0, "tb.u.o", "X"), c.Tr(0, "tb.u.o", "0"), c.Tr(50, "tb.u.o", "1")},
			extra: fmt.Sprintf(kid, "    specify\n        (i => o) = 0;\n    endspecify")})
	kase("t66_prc_covgrp__", "a covergroup sampled from an always block", "covergroup cg @(posedge s);\n        coverpoint s;\n    endgroup\n    cg c1 = new;", "t66_prc_ass_imm_", opt{})
}
