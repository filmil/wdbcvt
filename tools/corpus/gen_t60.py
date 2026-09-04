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

TCL = """open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/*]
log_wave -recursive *
%s
close_vcd
exit
"""

def case(name, brief, decl, write, differs, signals=(), trs=(), extra=None, dbg=True, tcl=None):
    axis = "debugging. %s beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section." % brief
    body = SV % {"brief": brief, "axis": axis, "decl": decl, "write": write}
    sigs = [sig("tb", "s", "logic")] + list(signals)
    t = [tr(0, "s", "0"), tr(50, "s", "1")] + list(trs)
    emit(name, axis, differs, [("tb.sv", body)], sigs, t, end=100, extra=extra, x=False)
    p = os.path.join(ROOT, name, "BUILD.bazel")
    b = open(p).read()
    attrs = ""
    if tcl:
        with open(os.path.join(ROOT, name, "xsim.tcl"), "w") as f:
            f.write(TCL % "\n".join(tcl))
        attrs += "    tcl = \"xsim.tcl\",\n"
    if dbg:
        attrs += "    xelab_args = [\n        \"-debug\",\n        \"all\",\n    ],\n"
    b = b.replace("    ],\n)\n", "    ],\n" + attrs + ")\n")
    open(p, "w").write(b)

# A string or class handle under -debug all is logged with one 32 bit
# record of zeros at time 0 and nothing at its write.
def held(name, typ):
    return [sig("tb", name, typ, 32)], [tr(0, name, "0" * 32)]

# A queue, dynamic array or associative array gets a declaration and a
# handle but is never logged.
def unlogged(name):
    return [sig("tb", name, "", 32, logged=False)], []

if __name__ == "__main__":
    case("t60_dbg_none____", "nothing", "", "", "t11_sv_bit______")
    case("t60_dbg_vec_____", "a 4 bit logic vector", "logic [3:0] v = 4'b0000;", "v = 4'b0101;", "t60_dbg_none____",
         [sig("tb", "v", "logic", 4)], [tr(0, "v", "0000"), tr(50, "v", "0101")])
    case("t60_dbg_int_____", "an int", "int i = 7;", "i = 9;", "t60_dbg_none____",
         [sig("tb", "i", "int", 32)], [tr(0, "i", "7"), tr(50, "i", "9")])
    case("t60_dbg_real____", "a real", "real r = 0.5;", "r = 1.5;", "t60_dbg_none____",
         [sig("tb", "r", "real", 64)], [tr(0, "r", "0.5"), tr(50, "r", "1.5")])
    case("t60_dbg_str_____", "a string variable", 'string str = "ab";', 'str = "xyz";', "t60_dbg_none____", *held("str", "string"))
    case("t60_dbg_queue___", "a queue", "int q[$];", "q.push_back(5);", "t60_dbg_int_____", *unlogged("q"))
    case("t60_dbg_dynarr__", "a dynamic array", "int d[];", "d = new[2]; d[1] = 5;", "t60_dbg_int_____", *unlogged("d"))
    case("t60_dbg_assoc___", "an associative array keyed by string", "int a[string];", 'a["k"] = 5;', "t60_dbg_int_____", *unlogged("a"))
    case("t60_dbg_assoc_i_", "an associative array keyed by int", "int a[int];", 'a[3] = 5;', "t60_dbg_assoc___", *unlogged("a"))
    case("t60_dbg_class___", "a class with one int field", "class c_t; int f = 1; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_int_____", *held("h", "c_t"))
    case("t60_dbg_class_2_", "a class with two fields", "class c_t; int f = 1; logic [3:0] g; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class___", *held("h", "c_t"))
    case("t60_dbg_class_d_", "a derived class", "class b_t; int f = 1; endclass\n    class c_t extends b_t; int g; endclass\n    c_t h;", "h = new; h.f = 5;", "t60_dbg_class___", *held("h", "c_t"))
    case("t60_dbg_class_2h", "two handles of one class", "class c_t; int f = 1; endclass\n    c_t h;\n    c_t h2;", "h = new; h2 = new; h.f = 5;", "t60_dbg_class___",
         held("h", "c_t")[0] + held("h2", "c_t")[0], held("h", "c_t")[1] + held("h2", "c_t")[1])
    case("t60_dbg_class_n_", "a class handle constructed at time 0", "class c_t; int f = 1; endclass\n    c_t h = new;", "h.f = 5;", "t60_dbg_class___", *held("h", "c_t"))
    # set_value /tb/str cd after run 25 ns ends the batch script without
    # a message, and the database closes at 25 ns; so the string is named
    # in log_wave instead.
    case("t60_dbg_str_log_", "a string named in log_wave", 'string str = "ab";', 'str = "xyz";', "t60_dbg_str_____",
         [sig("tb", "str", "string", 32, records=2)], held("str", "string")[1],
         tcl=["log_wave /tb/str", "run -all"])
    case("t60_dbg_q_log___", "a queue named in log_wave", "int q[$];", "q.push_back(5);", "t60_dbg_queue___", *unlogged("q"),
         tcl=["log_wave /tb/q", "run -all"])
    case("t60_dbg_struct__", "a packed struct", "typedef struct packed { logic a; logic [2:0] b; } st_t;\n    st_t st = '0;", "st = '{1, 3'b011};", "t60_dbg_vec_____",
         [{"scope": "tb", "name": "st", "type": "st_t", "fields": [
             {"name": "a", "width": 1, "type": "logic"}, {"name": "b", "width": 3, "type": ""}]}],
         [tr(0, "st.a", "0"), tr(0, "st.b", "000"), tr(50, "st.a", "1"), tr(50, "st.b", "011")])
    case("t60_dbg_mem_____", "a memory of two 4 bit words", "logic [3:0] m [0:1] = '{0, 0};", "m[1] = 4'd9;", "t60_dbg_vec_____",
         [sig("tb", "m", "memory", 8, elements=2, element_width=4, element_type="logic")],
         [tr(0, "m", "(0000, 0000)"), tr(50, "m", "(0000, 1001)")])
