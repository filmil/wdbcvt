#!/usr/bin/env python3
"""Tier 65: times past 32 bits of the file's unit."""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen_common import *

SV = """
// Corpus case: %(brief)s
//
// Axis: %(axis)s

`timescale %(ts)s

module tb;
    logic s = 1'b0;

    initial begin
%(body)s
        $finish;
    end
endmodule
"""

VH = """
-- Corpus case: %(brief)s
--
-- Axis: %(axis)s

library ieee;
use ieee.std_logic_1164.all;

entity tb is
end entity tb;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    process
    begin
%(body)s
        std.env.stop;
    end process;
end architecture sim;
"""

def case(name, brief, ts, body, differs, trs, end, vhdl=False):
    axis = "time past 32 bits. %s, to see whether the tier 44 reading of the 8 byte times holds across a page, in another unit and at a second past the largest time so far" % brief
    tmpl = VH if vhdl else SV
    src = tmpl % {"brief": brief, "axis": axis, "ts": ts, "body": body}
    emit(name, axis, differs, [("tb.vhdl" if vhdl else "tb.sv", src)], [sig("tb", "s", "std_ulogic" if vhdl else "logic")],
         [tr(0, "s", "0")] + trs, end=end, x=False)

if __name__ == "__main__":
    case("t65_tim_1s______", "a write at 999 ms, an end at 1 s", "1ns / 1ps",
         "        #999ms s = 1'b1;\n        #1ms;", "t44_time_5s_____", [tr(999000000, "s", "1")], 1000000000)
    case("t65_tim_cross___", "writes every 1 ns from 4.293 ms to 4.296 ms, across 2^32 ps in one page", "1ns / 1ps",
         "        #4293us;\n        repeat (3000) begin\n            #1ns s = ~s;\n        end", "t44_v_time_5ms__",
         [tr(4293000 + i + 1, "s", str((i + 1) % 2)) for i in range(3000)], 4296000)
    case("t65_tim_ns_5s___", "a write at 4.5 s under timescale 1ns / 1ns, past 2^32 ns", "1ns / 1ns",
         "        #4500ms s = 1'b1;\n        #500ms;", "t44_v_time_5ms__", [tr(4500000000, "s", "1")], 5000000000)
