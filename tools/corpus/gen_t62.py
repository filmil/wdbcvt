#!/usr/bin/env python3
"""Tier 62: net strengths, pull sources, switches and gate primitives."""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen_common import *
from gen_t60 import SV

def case(name, brief, decl, differs, signals=(), trs=()):
    axis = "strength. %s beside a logic, under typical, to see whether a net's drive strength, a pull source, a switch or a gate primitive leaves anything in the declaration, the hierarchy or the records." % brief
    body = SV % {"brief": brief, "axis": axis, "decl": decl, "write": ""}
    sigs = [sig("tb", "s", "logic")] + list(signals)
    t = [tr(0, "s", "0"), tr(50, "s", "1")] + list(trs)
    emit(name, axis, differs, [("tb.sv", body)], sigs, t, end=100, x=False)

def net(name, v0, v50, typ="wire", width=1, records=None):
    s = [sig("tb", name, typ, width, **({"records": records} if records else {}))]
    t = [tr(0, name, "X" * width), tr(0, name, v0)] + ([tr(50, name, v50)] if v50 != v0 else [])
    return s, t

if __name__ == "__main__":
    case("t62_str_none____", "nothing", "", "t11_sv_logic____")
    case("t62_str_wire____", "a wire driven by s", "wire w; assign w = s;", "t62_str_none____", *net("w", "0", "1", records=3))
    case("t62_str_tri_____", "a tri driven by s", "tri w; assign w = s;", "t62_str_wire____", *net("w", "0", "1", "tri"))
    case("t62_str_uwire___", "a uwire driven by s", "uwire w; assign w = s;", "t62_str_wire____", *net("w", "0", "1", "uwire"))
    case("t62_str_pullup__", "a wire with a pullup", "wire w; pullup (w);", "t62_str_wire____", *net("w", "1", "1", records=2))
    case("t62_str_pulldn__", "a wire with a pulldown", "wire w; pulldown (w);", "t62_str_pullup__", *net("w", "0", "0", records=2))
    case("t62_str_pu_drv__", "a pullup under a driver that releases", "wire w; pullup (w); assign w = s ? 1'bz : 1'b0;", "t62_str_pullup__", *net("w", "0", "1", records=4))
    case("t62_str_weak____", "a weak 1 under a driver that releases", "wire w; assign (weak0, weak1) w = 1'b1; assign w = s ? 1'bz : 1'b0;", "t62_str_pu_drv__", *net("w", "0", "1", records=4))
    case("t62_str_strong__", "a strong driver over a weak one", "wire w; assign (weak0, weak1) w = 1'b0; assign (strong0, strong1) w = s;", "t62_str_wire____", *net("w", "0", "1", records=4))
    case("t62_str_equal___", "two strong drivers that disagree", "wire w; assign w = 1'b0; assign w = s;", "t62_str_strong__", *net("w", "0", "X", records=3))
    case("t62_str_mixed___", "a strong0 weak1 driver against a pull1", "wire w; assign (strong0, weak1) w = s; assign (pull0, pull1) w = 1'b1;", "t62_str_strong__", *net("w", "0", "1", records=4))
    case("t62_str_supply__", "a supply driver against a strong one", "wire w; assign (supply0, supply1) w = 1'b0; assign w = s;", "t62_str_strong__", *net("w", "0", "0", records=4))
    case("t62_str_wand____", "a wand with a weak 0 and a strong s", "wand w; assign (weak0, weak1) w = 1'b0; assign w = s;", "t62_str_strong__", *net("w", "0", "1", "wand", records=4))
    case("t62_str_bufif___", "a bufif1 gate", "wire w; bufif1 (w, 1'b1, s);", "t62_str_wire____", *net("w", "Z", "1", records=4))
    case("t62_str_bufif_n_", "a named bufif1 gate", "wire w; bufif1 g1 (w, 1'b1, s);", "t62_str_bufif___", *net("w", "Z", "1", records=4))
    case("t62_str_and_____", "an and gate", "wire w; and (w, s, 1'b1);", "t62_str_bufif___", *net("w", "0", "1", records=3))
    case("t62_str_and_2___", "two and gates in one statement", "wire w, x; and g1 (w, s, 1'b1), g2 (x, s, s);", "t62_str_and_____",
         net("w", "0", "1")[0] + net("x", "0", "1")[0], net("w", "0", "1")[1] + net("x", "0", "1")[1])
    case("t62_str_nmos____", "an nmos switch", "wire w; nmos (w, 1'b1, s);", "t62_str_bufif___", *net("w", "Z", "1", records=4))
    case("t62_str_vec_pu__", "a vector with pullups", "wire [3:0] v; pullup p [3:0] (v); assign v = s ? 4'bzz01 : 4'b0000;", "t62_str_pu_drv__", [sig("tb", "v", "wire", 4, records=13)],
         [tr(0, "v", "XXXX"), tr(0, "v", "XXX0"), tr(0, "v", "XX00"), tr(0, "v", "X000"), tr(0, "v", "0000"),
          tr(50, "v", "0001"), tr(50, "v", "0101"), tr(50, "v", "1101")])
    case("t62_str_vec_1drv", "a vector with one driver", "wire [3:0] v; assign v = s ? 4'b1101 : 4'b0000;", "t62_str_wire____", *net("v", "0000", "1101", "wire", 4, records=3))
    case("t62_str_vec_2drv", "a vector with two drivers", "wire [3:0] v; assign v = s ? 4'bzz01 : 4'b0000; assign v = 4'bz1zz;", "t62_str_vec_1drv", [sig("tb", "v", "wire", 4, records=9)],
         [tr(0, "v", "XXXX"), tr(0, "v", "XXX0"), tr(0, "v", "XX00"), tr(0, "v", "0X00"),
          tr(50, "v", "0X01"), tr(50, "v", "0101"), tr(50, "v", "Z101")])
    case("t62_str_gate_dly", "a delayed and gate", "wire w; and #3 (w, s, 1'b1);", "t62_str_and_____", [sig("tb", "w", "wire")], [tr(0, "w", "X"), tr(3, "w", "0"), tr(53, "w", "1")])
