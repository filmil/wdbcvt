#!/usr/bin/env python3
"""Tier 67: an enumeration inside a packed struct."""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen_common import *

SV = """
// Corpus case: %(brief)s
//
// Axis: %(axis)s

`timescale 1ns / 1ps

module tb;
    %(decl)s

    initial begin
        #50 %(drive)s;
        #50 $finish;
    end
endmodule
"""

def case(name, brief, decl, drive, differs, sigs, trs):
    axis = "an enumeration in a packed struct. %s, under typical, to see how wide the enumeration's bits are where a neighbour depends on it." % brief
    body = SV % {"brief": brief, "axis": axis, "decl": decl, "drive": drive}
    emit(name, axis, differs, [("tb.sv", body)], sigs, trs, end=100, x=False)

def field(name, width, typ, declared=None):
    f = sig("", name, typ, width)
    if declared:
        f["declared"] = declared
    return f

if __name__ == "__main__":
    # An enum of two bits whose value needs one, between two logics.
    case("t67_esz_pk_2bit_", "a two bit enumeration between two logics",
         "typedef enum logic [1:0] {E0 = 2'd0, E1 = 2'd1} e2_t;\n"
         "    typedef struct packed { logic a; e2_t e; logic b; } rec_t;\n"
         "    rec_t s = '{a: 1'b0, e: E0, b: 1'b0};",
         "s = '{a: 1'b1, e: E1, b: 1'b1}", "t11_sv_struct___",
         [sig("tb", "s", "rec_t", 4, fields=[
             field("a", 1, "logic"), field("e", 2, "e2_t"), field("b", 1, "logic")])],
         [tr(0, "s.a", "0"), tr(0, "s.e", "E0"), tr(0, "s.b", "0"),
          tr(50, "s.a", "1"), tr(50, "s.e", "E1"), tr(50, "s.b", "1")])

    # The same with four bits, so the width of the field is not the
    # width of the value it holds.
    case("t67_esz_pk_4bit_", "a four bit enumeration between two logics",
         "typedef enum logic [3:0] {E0 = 4'd0, E1 = 4'd1} e4_t;\n"
         "    typedef struct packed { logic a; e4_t e; logic b; } rec_t;\n"
         "    rec_t s = '{a: 1'b0, e: E0, b: 1'b0};",
         "s = '{a: 1'b1, e: E1, b: 1'b1}", "t67_esz_pk_2bit_",
         [sig("tb", "s", "rec_t", 6, fields=[
             field("a", 1, "logic"), field("e", 4, "e4_t"), field("b", 1, "logic")])],
         [tr(0, "s.a", "0"), tr(0, "s.e", "E0"), tr(0, "s.b", "0"),
          tr(50, "s.a", "1"), tr(50, "s.e", "E1"), tr(50, "s.b", "1")])

    # An enumeration over a named type rather than a bare vector, which
    # is where the width sits on the base type instead.
    case("t67_esz_pk_int__", "an enumeration over int in a packed struct",
         "typedef enum int {E0 = 0, E1 = 1} ei_t;\n"
         "    typedef struct packed { logic a; ei_t e; logic b; } rec_t;\n"
         "    rec_t s = '{a: 1'b0, e: E0, b: 1'b0};",
         "s = '{a: 1'b1, e: E1, b: 1'b1}", "t67_esz_pk_2bit_",
         [sig("tb", "s", "rec_t", 34, fields=[
             field("a", 1, "logic"), field("e", 32, "ei_t"), field("b", 1, "logic")])],
         [tr(0, "s.a", "0"), tr(0, "s.e", "E0"), tr(0, "s.b", "0"),
          tr(50, "s.a", "1"), tr(50, "s.e", "E1"), tr(50, "s.b", "1")])
