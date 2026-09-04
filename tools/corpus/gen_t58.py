#!/usr/bin/env python3
"""Tier 58: what log_wave can name, in SystemVerilog."""
import sys, os, json
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen_common import *

TB = """// Corpus case: %(brief)s
//
// Axis: %(axis)s

`timescale 1ns / 1ps

module tb;
    typedef struct packed { logic a; logic [3:0] b; } st_t;
    parameter P = 3;
    localparam L = 4;
    logic s = 1'b0;
    logic [3:0] v = 4'b0000;
    logic [3:0] m [0:1] = '{4'd0, 4'd0};
    st_t st = '{a: 1'b0, b: 4'b0000};
    int i = 7;
    real r = 0.5;
    genvar g;
    for (g = 0; g < 2; g++) begin : gb
        wire gw = s;
    end
    task inc(input int x);
        int tmp;
        tmp = x + 1;
        i = tmp;
    endtask
    initial begin : blk
        int bv;
        bv = 1;
        #10;
        s = 1'b1;
        v = 4'b0101;
        m[1] = 4'd9;
        st = '{a: 1'b1, b: 4'b0011};
        r = 1.5;
        inc(1);
        bv = 2;
        #10;
        $finish;
    end
endmodule
"""

TCL = """open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
%s
run -all
close_vcd
exit
"""

def case(name, brief, axis, lines, logged, differs="t58_sv_log_all__"):
    tb = TB % {"brief": brief, "axis": axis}
    def lg(p):
        return {} if p in logged else {"logged": False}
    signals = [
        dict(sig("tb", "s", "logic"), **lg("tb.s")),
        dict(sig("tb", "v", "logic", 4), **lg("tb.v")),
        dict(sig("tb", "m", "memory", 8, elements=2, element_width=4, element_type="logic"), **lg("tb.m")),
        dict({"scope": "tb", "name": "st", "type": "st_t", "width": 5, "fields": [
            {"name": "a", "width": 1, "type": "logic"},
            {"name": "b", "width": 4, "type": "logic"}]}, **lg("tb.st")),
        dict(sig("tb", "i", "int", 32), **lg("tb.i")),
        dict(sig("tb", "r", "real", 64), **lg("tb.r")),
        dict(sig("tb.gb[0]", "gw", "wire"), **lg("tb.gb[0].gw")),
        dict(sig("tb.gb[1]", "gw", "wire"), **lg("tb.gb[1].gw")),
        dict(sig("tb.blk", "bv", "int", 32), **lg("tb.blk.bv")),
        dict(sig("tb.inc", "x", "int", 32, port="in"), **lg("tb.inc.x")),
        dict(sig("tb.inc", "tmp", "int", 32), **lg("tb.inc.tmp")),
    ]
    trs = []
    def add(p, *xs):
        if p in logged:
            trs.extend(xs)
    add("tb.s", tr(0, "s", "0"), tr(10, "s", "1"))
    add("tb.v", tr(0, "v", "0000"), tr(10, "v", "0101"))
    add("tb.m", tr(0, "m", "(0000, 0000)"), tr(10, "m", "(0000, 1001)"))
    add("tb.st", tr(0, "st.a", "0"), tr(0, "st.b", "0000"), tr(10, "st.a", "1"), tr(10, "st.b", "0011"))
    add("tb.i", tr(0, "i", "7"), tr(10, "i", "2"))
    add("tb.r", tr(0, "r", "0.5"), tr(10, "r", "1.5"))
    for k in (0, 1):
        p = "tb.gb[%d].gw" % k
        add(p, tr(0, p, "X"), tr(0, p, "0"), tr(10, p, "1"))
    add("tb.blk.bv", tr(0, "tb.blk.bv", "0"), tr(0, "tb.blk.bv", "1"), tr(10, "tb.blk.bv", "2"))
    add("tb.inc.x", tr(0, "tb.inc.x", "0"), tr(10, "tb.inc.x", "1"))
    add("tb.inc.tmp", tr(0, "tb.inc.tmp", "0"), tr(10, "tb.inc.tmp", "2"))
    generics = [
        dict({"instance": "", "scope": "tb", "name": "P", "type": "", "declared": "parameter",
              "value": bits(3, 32)}, **lg("tb.P")),
        dict({"instance": "", "scope": "tb", "name": "L", "type": "", "declared": "localparam",
              "value": bits(4, 32)}, **lg("tb.L")),
    ]
    emit(name, axis, differs, [("tb.sv", tb)], signals, trs, end=20,
         extra={"generics": generics}, x=False)
    d = os.path.join(ROOT, name)
    with open(os.path.join(d, "xsim.tcl"), "w") as f:
        f.write(TCL % "\n".join(lines))
    p = os.path.join(d, "BUILD.bazel")
    b = open(p).read().replace("    ],\n)\n", "    ],\n    tcl = \"xsim.tcl\",\n)\n")
    open(p, "w").write(b)

ALL = ["tb.s", "tb.v", "tb.m", "tb.st", "tb.i", "tb.r", "tb.gb[0].gw", "tb.gb[1].gw",
       "tb.blk.bv", "tb.inc.x", "tb.inc.tmp", "tb.P", "tb.L"]
TOP = ["tb.s", "tb.v", "tb.m", "tb.st", "tb.i", "tb.r", "tb.gb[0].gw", "tb.gb[1].gw", "tb.P", "tb.L"]
BRIEF = "log_wave naming %s of a SystemVerilog design with every kind of object"
AXIS = "logging. log_wave names %s, in a SystemVerilog design with a logic, a vector, a memory, a packed struct, an int, a real, a parameter, a localparam, a generate with a wire, a named block with a variable and a static task, to see what the database logs."

def one(name, what, obj, logged, vcd=None):
    lines = ["log_wave %s" % obj]
    if vcd != "":
        lines.insert(0, "log_vcd %s" % (vcd or obj))
    case(name, BRIEF % what, AXIS % what, lines, logged)

if __name__ == "__main__":
    case("t58_sv_log_all__", BRIEF % "everything, -recursive *", AXIS % "everything with -recursive *",
         ["log_vcd [get_objects -r /* ]", "log_wave -recursive *"], ALL, differs="t57_log_all_____")
    case("t58_sv_log_none_", BRIEF % "nothing", AXIS % "nothing, the script has no log_wave", [], [])
    one("t58_sv_log_bit__", "one bit of a vector", "{/tb/v[3]}", [], vcd="")
    one("t58_sv_log_slc__", "a slice of a vector", "{/tb/v[2:1]}", ["tb.v"])
    one("t58_sv_log_mem_e", "one element of a memory", "{/tb/m[1]}", [], vcd="")
    one("t58_sv_log_mem__", "a memory", "/tb/m", ["tb.m"])
    one("t58_sv_log_st_fl", "a field of a packed struct", "/tb/st.a", [], vcd="")
    one("t58_sv_log_st___", "a packed struct", "/tb/st", ["tb.st"])
    one("t58_sv_log_int__", "an int variable of the module", "/tb/i", ["tb.i"])
    one("t58_sv_log_real_", "a real variable of the module", "/tb/r", ["tb.r"])
    one("t58_sv_log_prm__", "a parameter", "/tb/P", ["tb.P"])
    one("t58_sv_log_lprm_", "a localparam", "/tb/L", ["tb.L"])
    one("t58_sv_log_blkv_", "a variable of a named block", "/tb/blk/bv", ["tb.blk.bv"])
    one("t58_sv_log_blk__", "a named block", "/tb/blk", ["tb.blk.bv"], vcd="[get_objects /tb/blk/*]")
    one("t58_sv_log_tsk_l", "a local of a static task", "/tb/inc/tmp", ["tb.inc.tmp"])
    one("t58_sv_log_tsk_a", "an argument of a static task", "/tb/inc/x", ["tb.inc.x"])
    one("t58_sv_log_tsk__", "a static task", "/tb/inc", ["tb.inc.x", "tb.inc.tmp"], vcd="[get_objects /tb/inc/*]")
    one("t58_sv_log_gen_w", "the wire of one generate block, through get_objects -regexp",
        "[get_objects -regexp {/tb/.*gb\\[1\\].*}]", ["tb.gb[1].gw"])
    one("t58_sv_log_gen__", "a generate block by path", "{/tb/gb[1]}", [], vcd="")
    one("t58_sv_log_top__", "the top without -recursive", "/tb", TOP, vcd="[get_objects /tb/*]")
