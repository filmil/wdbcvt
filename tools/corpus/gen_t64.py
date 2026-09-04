#!/usr/bin/env python3
"""Tier 64: several partial drivers on one net."""
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
module child(input i, output o);
    assign o = i;
endmodule
"""

CHILD4 = """
module child4(input a, input b, output c, output d);
    assign c = a;
    assign d = b;
endmodule
"""

CHILD2 = CHILD + """
module child2(input i, output o);
    assign o = ~i;
endmodule
"""

BIDI = """
module bidi(inout io);
    assign io = 1'b0;
endmodule
"""

def case(name, brief, decl, differs, sigs, trs, child=""):
    axis = "several partial drivers. %s beside a logic, under typical, to see the order and the place of the records the drivers write." % brief
    body = SV % {"brief": brief, "axis": axis, "decl": decl, "child": child}
    emit(name, axis, differs, [("tb.sv", body)], [sig("tb", "s", "logic")] + sigs, [tr(0, "s", "0"), tr(50, "s", "1")] + trs, end=100, x=False)

def net(w, records, name="v"):
    return sig("tb", name, "wire", w, records=records)

def seq(vals, name="v"):
    return [tr(t, name, v) for t, v in vals]

Z62 = "Z" * 62
Z2398 = "Z" * 2398

if __name__ == "__main__":
    case("t64_ord_src_rev_", "the two drivers of two bits in the other source order", "wire [3:0] v; assign v[3] = ~s; assign v[0] = s;", "t63_pdr_two_bits",
         [net(4, 5)], seq([(0, "XZZX"), (0, "1ZZX"), (0, "1ZZ0"), (50, "0ZZ0"), (50, "0ZZ1")]))
    case("t64_ord_gen4____", "four drivers of four bits from a generate loop", "wire [3:0] v; genvar i; for (i = 0; i < 4; i = i + 1) begin : g assign v[i] = s; end", "t63_pdr_two_bits",
         [net(4, 9)], seq([(0, "XXXX"), (0, "XXX0"), (0, "XX00"), (0, "X000"), (0, "0000"), (50, "0001"), (50, "0011"), (50, "0111"), (50, "1111")]))
    case("t64_ord_gen_rev_", "the generate loop counting down", "wire [3:0] v; genvar i; for (i = 3; i >= 0; i = i - 1) begin : g assign v[i] = s; end", "t64_ord_gen4____",
         [net(4, 9)], seq([(0, "XXXX"), (0, "0XXX"), (0, "00XX"), (0, "000X"), (0, "0000"), (50, "1000"), (50, "1100"), (50, "1110"), (50, "1111")]))
    case("t64_ord_w64_two_", "two drivers in the two pairs of a 64 bit net", "wire [63:0] v; assign v[0] = s; assign v[63] = ~s;", "t63_pdr_two_bits",
         [net(64, 5)], seq([(0, "X" + Z62 + "X"), (0, "X" + Z62 + "0"), (0, "1" + Z62 + "0"), (50, "1" + Z62 + "1"), (50, "0" + Z62 + "1")]))
    case("t64_ord_2400_two", "two drivers in the first and last chunk of a 2400 bit net", "wire [2399:0] v; assign v[0] = s; assign v[2399] = ~s;", "t64_ord_w64_two_",
         [net(2400, 5)], seq([(0, "X" + Z2398 + "X"), (0, "X" + Z2398 + "0"), (0, "1" + Z2398 + "0"), (50, "1" + Z2398 + "1"), (50, "0" + Z2398 + "1")]))
    case("t64_ord_unp_elem", "a driver of one bit of one element of an unpacked array of nets", "wire [3:0] v [0:1]; assign v[1][2] = s;", "t63_pdr_bit0____",
         [sig("tb", "v", "memory", 8, elements=2, element_width=4, element_type="wire", records=3)],
         seq([(0, "(ZZZZ, ZXZZ)"), (0, "(ZZZZ, Z0ZZ)"), (50, "(ZZZZ, Z1ZZ)")]))
    case("t64_ord_unp_whol", "a driver of one whole element of an unpacked array of nets", "wire [3:0] v [0:1]; assign v[1] = {4{s}};", "t64_ord_unp_elem",
         [sig("tb", "v", "memory", 8, elements=2, element_width=4, element_type="wire", records=3)],
         seq([(0, "(ZZZZ, XXXX)"), (0, "(ZZZZ, 0000)"), (50, "(ZZZZ, 1111)")]))
    case("t64_ord_two_kids", "two child outputs on two bits of one net", "wire [3:0] v; child u0(.i(s), .o(v[1])); child u1(.i(~s), .o(v[3]));", "t63_pdr_port_bit",
         [net(4, 7), sig("tb.u0", "i", "wire", port="in", records=3), sig("tb.u0", "o", "wire", port="out", records=7),
          sig("tb.u1", "i", "wire", port="in", records=3), sig("tb.u1", "o", "wire", port="out", records=7)],
         seq([(0, "XZXZ"), (0, "XZ0Z"), (0, "1Z0Z"), (50, "1Z1Z"), (50, "0Z1Z")])
         + seq([(0, "X"), (0, "0"), (50, "1")], "tb.u0.i") + seq([(0, "X"), (0, "0"), (50, "1")], "tb.u0.o")
         + seq([(0, "X"), (0, "1"), (50, "0")], "tb.u1.i") + seq([(0, "X"), (0, "1"), (50, "0")], "tb.u1.o"), child=CHILD)
    kid = lambda u, iv, ov, r=4: ([sig("tb." + u, "i", "wire", port="in", records=3), sig("tb." + u, "o", "wire", port="out", records=r)],
                             seq([(0, "X"), (0, iv), (50, ov)], "tb.%s.i" % u) + seq([(0, "X"), (0, iv), (50, ov)], "tb.%s.o" % u))
    case("t64_ord_pos_expr", "one child output on a scalar net, the input bound to an expression", "wire w; child u1(.i(~s), .o(w));", "t64_ord_two_kids",
         [net(1, 4, "w")] + kid("u1", "1", "0")[0], seq([(0, "X"), (0, "1"), (50, "0")], "w") + kid("u1", "1", "0")[1], child=CHILD)
    case("t64_ord_pos_bit3", "one child output on bit 3, the input bound to the logic", "wire [3:0] v; child u1(.i(s), .o(v[3]));", "t64_ord_two_kids",
         [net(4, 4)] + kid("u1", "0", "1")[0], seq([(0, "XZZZ"), (0, "0ZZZ"), (50, "1ZZZ")]) + kid("u1", "0", "1")[1], child=CHILD)
    case("t64_ord_two_same", "two child outputs on two bits, both inputs bound to the logic", "wire [3:0] v; child u0(.i(s), .o(v[1])); child u1(.i(s), .o(v[3]));", "t64_ord_two_kids",
         [net(4, 7)] + kid("u0", "0", "1", 7)[0] + kid("u1", "0", "1", 7)[0],
         seq([(0, "XZXZ"), (0, "XZ0Z"), (0, "0Z0Z"), (50, "1Z0Z"), (50, "1Z1Z")]) + kid("u0", "0", "1")[1] + kid("u1", "0", "1")[1], child=CHILD)
    case("t64_ord_two_nets", "two child outputs on two scalar nets", "wire a, b; child u0(.i(s), .o(a)); child u1(.i(~s), .o(b));", "t64_ord_two_kids",
         [net(1, 4, "a"), net(1, 4, "b")] + kid("u0", "0", "1")[0] + kid("u1", "1", "0")[0],
         seq([(0, "X"), (0, "0"), (50, "1")], "a") + seq([(0, "X"), (0, "1"), (50, "0")], "b") + kid("u0", "0", "1")[1] + kid("u1", "1", "0")[1], child=CHILD)
    kid4 = lambda u: ([sig("tb." + u, "a", "wire", port="in", records=3), sig("tb." + u, "b", "wire", port="in", records=3),
                       sig("tb." + u, "c", "wire", port="out", records=4), sig("tb." + u, "d", "wire", port="out", records=4)],
                      sum([seq([(0, "X"), (0, "0"), (50, "1")], "tb.%s.%s" % (u, n)) for n in "abcd"], []))
    case("t64_ord_two_pos4", "two instances of a child with four ports", "wire c0, d0, c1, d1; child4 u0(.a(s), .b(s), .c(c0), .d(d0)); child4 u1(.a(s), .b(s), .c(c1), .d(d1));", "t64_ord_two_nets",
         [net(1, 4, n) for n in ("c0", "d0", "c1", "d1")] + kid4("u0")[0] + kid4("u1")[0],
         sum([seq([(0, "X"), (0, "0"), (50, "1")], n) for n in ("c0", "d0", "c1", "d1")], []) + kid4("u0")[1] + kid4("u1")[1], child=CHILD4)
    case("t64_ord_three___", "three instances of the child on three scalar nets", "wire a, b, c; child u0(.i(s), .o(a)); child u1(.i(s), .o(b)); child u2(.i(s), .o(c));", "t64_ord_two_nets",
         [net(1, 4, n) for n in "abc"] + kid("u0", "0", "1")[0] + kid("u1", "0", "1")[0] + kid("u2", "0", "1")[0],
         sum([seq([(0, "X"), (0, "0"), (50, "1")], n) for n in "abc"], []) + kid("u0", "0", "1")[1] + kid("u1", "0", "1")[1] + kid("u2", "0", "1")[1], child=CHILD)
    case("t64_ord_two_mods", "two children of two modules on two scalar nets", "wire a, b; child u0(.i(s), .o(a)); child2 u1(.i(s), .o(b));", "t64_ord_two_nets",
         [net(1, 4, "a"), net(1, 4, "b")] + kid("u0", "0", "1")[0] + kid("u1", "0", "1")[0],
         seq([(0, "X"), (0, "0"), (50, "1")], "a") + seq([(0, "X"), (0, "1"), (50, "0")], "b") + kid("u0", "0", "1")[1]
         + seq([(0, "X"), (0, "0"), (50, "1")], "tb.u1.i") + seq([(0, "X"), (0, "1"), (50, "0")], "tb.u1.o"), child=CHILD2)
    case("t64_ord_gen_kids", "two child outputs on two bits from a generate loop", "wire [3:0] v; genvar i; for (i = 0; i < 2; i = i + 1) begin : g child u(.i(s), .o(v[i * 3])); end", "t64_ord_two_same",
         [net(4, 7)] + kid("g[0].u", "0", "1", 7)[0] + kid("g[1].u", "0", "1", 7)[0],
         seq([(0, "XZZX"), (0, "XZZ0"), (0, "0ZZ0"), (50, "1ZZ0"), (50, "1ZZ1")]) + kid("g[0].u", "0", "1")[1] + kid("g[1].u", "0", "1")[1], child=CHILD)
    case("t64_ord_inout___", "a child inout on one bit beside a driver of another", "wire [3:0] v; assign v[1] = s; bidi u(.io(v[3]));", "t63_pdr_port_bit",
         [net(4, 5), sig("tb.u", "io", "wire", port="inout", records=5)],
         seq([(0, "XZXZ"), (0, "XZ0Z"), (0, "0Z0Z"), (50, "0Z1Z")]) + seq([(0, "X"), (0, "0")], "tb.u.io"), child=BIDI)
    case("t64_ord_self____", "a driver of one bit from another bit of the same net", "wire [3:0] v; assign v[0] = s; assign v[1] = v[0];", "t63_pdr_two_bits",
         [net(4, 6)], seq([(0, "ZZXX"), (0, "ZZX0"), (0, "ZZ00"), (50, "ZZ01"), (50, "ZZ11")]))
    case("t64_ord_chain___", "a driver of one bit of a second net from a bit of the first", "wire [3:0] v, w; assign v[0] = s; assign w[1] = v[0];", "t64_ord_self____",
         [net(4, 3), net(4, 3, "w")], seq([(0, "ZZZX"), (0, "ZZZ0"), (50, "ZZZ1")]) + seq([(0, "ZZXZ"), (0, "ZZ0Z"), (50, "ZZ1Z")], "w"))
