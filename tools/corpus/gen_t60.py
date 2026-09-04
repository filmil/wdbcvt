#!/usr/bin/env python3
"""Tier 60: SystemVerilog under -debug all."""
import sys, os, json
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
        %(write)s
        #50 $finish;
    end
endmodule
"""

def case(name, brief, decl, write, differs, signals=(), trs=(), extra=None, dbg=True):
    axis = "debugging. %s beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section." % brief
    body = SV % {"brief": brief, "axis": axis, "decl": decl, "write": write}
    sigs = [sig("tb", "s", "logic")] + list(signals)
    t = [tr(0, "s", "0"), tr(50, "s", "1")] + list(trs)
    emit(name, axis, differs, [("tb.sv", body)], sigs, t, end=100, extra=extra, x=False)
    if dbg:
        p = os.path.join(ROOT, name, "BUILD.bazel")
        b = open(p).read()
        b = b.replace("    ],\n)\n", "    ],\n    xelab_args = [\n        \"-debug\",\n        \"all\",\n    ],\n)\n")
        open(p, "w").write(b)

if __name__ == "__main__":
    case("t60_dbg_none____", "nothing", "", "", "t11_sv_bit______")
    case("t60_dbg_vec_____", "a 4 bit logic vector", "logic [3:0] v = 4'b0000;", "v = 4'b0101;", "t60_dbg_none____",
         [sig("tb", "v", "logic", 4)], [tr(0, "v", "0000"), tr(50, "v", "0101")])
    case("t60_dbg_int_____", "an int", "int i = 7;", "i = 9;", "t60_dbg_none____",
         [sig("tb", "i", "int", 32)], [tr(0, "i", bits(7, 32)), tr(50, "i", bits(9, 32))])
    case("t60_dbg_real____", "a real", "real r = 0.5;", "r = 1.5;", "t60_dbg_none____",
         [sig("tb", "r", "real", 64)], [tr(0, "r", "0.5"), tr(50, "r", "1.5")])
    case("t60_dbg_str_____", "a string variable", 'string str = "ab";', 'str = "xyz";', "t60_dbg_none____")
    case("t60_dbg_queue___", "a queue", "int q[$];", "q.push_back(5);", "t60_dbg_int_____")
    case("t60_dbg_dynarr__", "a dynamic array", "int d[];", "d = new[2]; d[1] = 5;", "t60_dbg_int_____")
    case("t60_dbg_assoc___", "an associative array keyed by string", "int a[string];", 'a["k"] = 5;', "t60_dbg_int_____")
    case("t60_dbg_assoc_i_", "an associative array keyed by int", "int a[int];", 'a[3] = 5;', "t60_dbg_assoc___")
    case("t60_dbg_class___", "a class with one int field", "class c_t; int f = 1; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_int_____")
    case("t60_dbg_class_2_", "a class with two fields", "class c_t; int f = 1; logic [3:0] g; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class___")
    case("t60_dbg_class_d_", "a derived class", "class b_t; int f = 1; endclass\n    class c_t extends b_t; int g; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class___")
    case("t60_dbg_struct__", "a packed struct", "typedef struct packed { logic a; logic [2:0] b; } st_t;\n    st_t st = '0;", "st = '{1, 3'b011};", "t60_dbg_vec_____",
         [{"scope": "tb", "name": "st", "type": "st_t", "fields": [
             {"name": "a", "width": 1, "type": "logic"}, {"name": "b", "width": 3, "type": ""}]}],
         [tr(0, "st.a", "0"), tr(0, "st.b", "000"), tr(50, "st.a", "1"), tr(50, "st.b", "011")])
    case("t60_dbg_mem_____", "a memory of two 4 bit words", "logic [3:0] m [0:1] = '{0, 0};", "m[1] = 4'd9;", "t60_dbg_vec_____",
         [sig("tb", "m", "memory", 8, elements=2, element_width=4, element_type="logic")],
         [tr(0, "m", "(0000, 0000)"), tr(50, "m", "(0000, 1001)")])
