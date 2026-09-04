#!/usr/bin/env python3
"""Tier 66: process scopes of the SystemVerilog constructs not yet seen."""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen_common import *

SV = """
// Corpus case: %(brief)s
//
// Axis: %(axis)s

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    %(decl)s

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
%(extra)s"""

def case(name, brief, decl, differs, sigs=(), trs=(), extra="", end=100, trs_all=None):
    axis = "process scopes. %s beside a logic, under typical, to see what scope the construct leaves and what it declares." % brief
    body = SV % {"brief": brief, "axis": axis, "decl": decl, "extra": extra}
    emit(name, axis, differs, [("tb.sv", body)], [sig("tb", "s", "logic")] + list(sigs),
         trs_all if trs_all is not None else [tr(0, "s", "0"), tr(50, "s", "1")] + list(trs), end=end, x=False)

if __name__ == "__main__":
    case("t66_prc_final___", "a final block", "final begin\n        $display(\"done\");\n    end", "t11_sv_logic____")
    case("t66_prc_latch___", "an always_latch", "logic q;\n    always_latch if (s) q <= 1'b1;", "t11_sv_logic____",
         [sig("tb", "q", "logic")], [tr(0, "q", "X"), tr(50, "q", "1")])
    case("t66_prc_ass_imm_", "an immediate assertion in an always block", "always @(s) assert (s !== 1'bx);", "t11_sv_logic____")
    case("t66_prc_ass_conc", "a concurrent assertion on a clock", "logic c = 1'b0;\n    always #10 c = ~c;\n    assert property (@(posedge c) 1'b1);", "t66_prc_ass_imm_",
         [sig("tb", "c", "logic")], [tr(0, "c", "0")] + [tr(10 * (i + 1), "c", str((i + 1) % 2)) for i in range(9)])
    case("t66_prc_prop____", "a named property and sequence", "logic c = 1'b0;\n    always #10 c = ~c;\n    sequence q; 1'b1; endsequence\n    property p; @(posedge c) q; endproperty\n    assert property (p);", "t66_prc_ass_conc",
         [sig("tb", "c", "logic")], [tr(0, "c", "0")] + [tr(10 * (i + 1), "c", str((i + 1) % 2)) for i in range(9)])
    case("t66_prc_task____", "a task called from an initial block", "task t; #10; endtask\n    initial t();", "t11_sv_logic____")
    case("t66_prc_func____", "a function called from an assign", "wire w;\n    function automatic logic f(input logic x); return ~x; endfunction\n    assign w = f(s);", "t66_prc_task____",
         [sig("tb", "w", "wire", records=3)], [tr(0, "w", "X"), tr(0, "w", "1"), tr(50, "w", "0")])
    case("t66_prc_program_", "a program block instantiated beside the module", "prog p();", "t11_sv_logic____", end=10, trs_all=[tr(0, "s", "0")],
         extra="""
program prog;
    initial begin
        #10;
    end
endprogram
""")
    case("t66_prc_bind____", "a child bound into the module with bind", "wire w; assign w = s;", "t66_prc_func____",
         [sig("tb", "w", "wire", records=3), sig("tb.b", "i", "wire", port="in", records=3)],
         [tr(0, "w", "X"), tr(0, "w", "0"), tr(50, "w", "1"), tr(0, "tb.b.i", "X"), tr(0, "tb.b.i", "0"), tr(50, "tb.b.i", "1")],
         extra="""
module watcher(input i);
endmodule

bind tb watcher b(.i(s));
""")
    case("t66_prc_specify_", "a specify block in a child", "wire w; kid u(.i(s), .o(w));", "t66_prc_func____",
         [sig("tb", "w", "wire", records=4), sig("tb.u", "i", "wire", port="in", records=3), sig("tb.u", "o", "wire", port="out", records=4)],
         [tr(0, "w", "X"), tr(1, "w", "0"), tr(51, "w", "1"), tr(0, "tb.u.i", "X"), tr(0, "tb.u.i", "0"), tr(50, "tb.u.i", "1"),
          tr(0, "tb.u.o", "X"), tr(1, "tb.u.o", "0"), tr(51, "tb.u.o", "1")],
         extra="""
module kid(input i, output o);
    assign o = i;
    specify
        (i => o) = 1;
    endspecify
endmodule
""")
    KID = """
module kid(input i, output o);
    assign o = i;
%s
endmodule
"""
    case("t66_prc_kid_____", "the same child without a specify block", "wire w; kid u(.i(s), .o(w));", "t66_prc_specify_",
         [sig("tb", "w", "wire", records=4), sig("tb.u", "i", "wire", port="in", records=3), sig("tb.u", "o", "wire", port="out", records=4)],
         [tr(0, "w", "X"), tr(0, "w", "0"), tr(50, "w", "1"), tr(0, "tb.u.i", "X"), tr(0, "tb.u.i", "0"), tr(50, "tb.u.i", "1"),
          tr(0, "tb.u.o", "X"), tr(0, "tb.u.o", "0"), tr(50, "tb.u.o", "1")], extra=KID % "")
    case("t66_prc_spec_0__", "the specify path delay set to 0", "wire w; kid u(.i(s), .o(w));", "t66_prc_specify_",
         [sig("tb", "w", "wire", records=4), sig("tb.u", "i", "wire", port="in", records=3), sig("tb.u", "o", "wire", port="out", records=4)],
         [tr(0, "w", "X"), tr(0, "w", "0"), tr(50, "w", "1"), tr(0, "tb.u.i", "X"), tr(0, "tb.u.i", "0"), tr(50, "tb.u.i", "1"),
          tr(0, "tb.u.o", "X"), tr(0, "tb.u.o", "0"), tr(50, "tb.u.o", "1")], extra=KID % "    specify\n        (i => o) = 0;\n    endspecify")
    case("t66_prc_covgrp__", "a covergroup sampled from an always block", "covergroup cg @(posedge s);\n        coverpoint s;\n    endgroup\n    cg c1 = new;", "t66_prc_ass_imm_")
