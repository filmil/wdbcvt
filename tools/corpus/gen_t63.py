#!/usr/bin/env python3
"""Tier 63: partial drivers on a net."""
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
%(child)s"""

CHILD = """
module child(input i, output %so);
    assign o = %s;
endmodule
"""

def child_of(w):
    return CHILD % ("", "i") if w == 1 else CHILD % ("[%d:0] " % (w - 1), "{%d{i}}" % w)

# drv: {bit: expr} with expr "s" or "~s"; s is 0, 1 or "X" for the first
# record, which holds X on the driven bits and Z on the rest.
def val(w, drv, s):
    out = ["Z"] * w
    for b, e in drv.items():
        out[w - 1 - b] = "X" if s == "X" else str(s if e == "s" else 1 - s)
    return "".join(out)

def case(name, brief, decl, differs, w, drv, child=None, extra_sigs=(), extra_trs=(), records=None, vname="v", trs=None):
    axis = "partial drivers. %s beside a logic, under typical, to see whether a driver of a bit, a slice or a port bound to part of a net records the whole net or the part." % brief
    body = SV % {"brief": brief, "axis": axis, "decl": decl, "child": child_of(child) if child else ""}
    sigs = [sig("tb", "s", "logic")] + ([sig("tb", vname, "wire", w, **({"records": records} if records else {}))] if w else []) + list(extra_sigs)
    t = [tr(0, "s", "0"), tr(50, "s", "1")]
    if w:
        t += [tr(0, vname, val(w, drv, "X"))] + (trs if trs else [tr(0, vname, val(w, drv, 0)), tr(50, vname, val(w, drv, 1))])
    t += list(extra_trs)
    emit(name, axis, differs, [("tb.sv", body)], sigs, t, end=100, x=False)

def rng(lo, hi, e="s"):
    return {b: e for b in range(lo, hi + 1)}

if __name__ == "__main__":
    case("t63_pdr_bit0____", "bit 0 of a 4 bit net driven", "wire [3:0] v; assign v[0] = s;", "t62_str_vec_1drv", 4, {0: "s"}, records=3)
    case("t63_pdr_bit3____", "bit 3 of a 4 bit net driven", "wire [3:0] v; assign v[3] = s;", "t63_pdr_bit0____", 4, {3: "s"}, records=3)
    case("t63_pdr_two_bits", "bits 0 and 3 driven from two assigns", "wire [3:0] v; assign v[0] = s; assign v[3] = ~s;", "t63_pdr_bit0____", 4, {0: "s", 3: "~s"},
         trs=[tr(0, "v", "XZZ0"), tr(0, "v", "1ZZ0"), tr(50, "v", "1ZZ1"), tr(50, "v", "0ZZ1")], records=5)
    case("t63_pdr_slice___", "the low nibble of an 8 bit net driven", "wire [7:0] v; assign v[3:0] = {4{s}};", "t63_pdr_bit0____", 8, rng(0, 3), records=3)
    case("t63_pdr_w64_bit0", "bit 0 of a 64 bit net driven", "wire [63:0] v; assign v[0] = s;", "t63_pdr_bit0____", 64, {0: "s"}, records=3)
    case("t63_pdr_w64_bit6", "bit 63 of a 64 bit net driven", "wire [63:0] v; assign v[63] = s;", "t63_pdr_w64_bit0", 64, {63: "s"}, records=3)
    case("t63_pdr_w64_hi__", "the high word of a 64 bit net driven", "wire [63:0] v; assign v[63:32] = {32{s}};", "t63_pdr_w64_bit0", 64, rng(32, 63), records=3)
    case("t63_pdr_w64_all_", "all 64 bits of a 64 bit net driven", "wire [63:0] v; assign v = {64{s}};", "t63_pdr_w64_hi__", 64, rng(0, 63), records=3)
    case("t63_pdr_2400_bit", "bit 0 of a 2400 bit net driven", "wire [2399:0] v; assign v[0] = s;", "t63_pdr_w64_bit0", 2400, {0: "s"}, records=3)
    case("t63_pdr_2400_hi_", "the top 400 bits of a 2400 bit net driven", "wire [2399:0] v; assign v[2399:2000] = {400{s}};", "t63_pdr_2400_bit", 2400, rng(2000, 2399), records=3)
    case("t63_pdr_2400_all", "all 2400 bits of a 2400 bit net driven", "wire [2399:0] v; assign v = {2400{s}};", "t63_pdr_2400_hi_", 2400, rng(0, 2399), records=3)
    case("t63_pdr_concat__", "two scalar nets driven through a concatenation", "wire a, b; assign {a, b} = {s, ~s};", "t62_str_wire____", 0, {},
         extra_sigs=[sig("tb", "a", "wire", records=3), sig("tb", "b", "wire", records=3)],
         extra_trs=[tr(0, "a", "X"), tr(0, "a", "0"), tr(50, "a", "1"), tr(0, "b", "X"), tr(0, "b", "1"), tr(50, "b", "0")])
    port = lambda w: [sig("tb.u", "i", "wire", port="in", records=3), sig("tb.u", "o", "wire", w, port="out", records=4)]
    ptrs = lambda w: [tr(0, "tb.u.i", "X"), tr(0, "tb.u.i", "0"), tr(50, "tb.u.i", "1"), tr(0, "tb.u.o", "X" * w), tr(0, "tb.u.o", "0" * w), tr(50, "tb.u.o", "1" * w)]
    case("t63_pdr_port_bit", "a child output bound to bit 1 of a 4 bit net", "wire [3:0] v; child u(.i(s), .o(v[1]));", "t63_pdr_bit0____", 4, {1: "s"}, child=1,
         extra_sigs=port(1), extra_trs=ptrs(1), records=4)
    case("t63_pdr_port_slc", "a child output bound to the high nibble of an 8 bit net", "wire [7:0] v; child u(.i(s), .o(v[7:4]));", "t63_pdr_port_bit", 8, rng(4, 7), child=4,
         extra_sigs=port(4), extra_trs=ptrs(4), records=4)
    case("t63_pdr_port_hi_", "a child output bound to the high word of a 64 bit net", "wire [63:0] v; child u(.i(s), .o(v[63:32]));", "t63_pdr_port_slc", 64, rng(32, 63), child=32,
         extra_sigs=port(32), extra_trs=ptrs(32), records=4)