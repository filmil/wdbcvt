// SPDX-License-Identifier: Apache-2.0

// Tier 65: times past 32 bits of the file's unit.
package main

import (
	"fmt"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

const sv = `
// Corpus case: %(brief)s
//
// Axis: %(axis)s

` + "`" + `timescale %(ts)s

module tb;
    logic s = 1'b0;

    initial begin
%(body)s
        $finish;
    end
endmodule
`

const vh = `
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
`

func kase(name, brief, ts, body, differs string, trs []*c.Obj, end int, vhdl bool) {
	axis := "time past 32 bits. " + brief + ", to see whether the tier 44 reading of the 8 byte times holds across a page, in another unit and at a second past the largest time so far"
	tmpl, fn, typ := sv, "tb.sv", "logic"
	if vhdl {
		tmpl, fn, typ = vh, "tb.vhdl", "std_ulogic"
	}
	src := c.Fill(tmpl, "brief", brief, "axis", axis, "ts", ts, "body", body)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: fn, Body: src}},
		Signals: []*c.Obj{c.Sig("tb", "s", typ, 1)},
		Trs:     append([]*c.Obj{c.Tr(0, "s", "0")}, trs...),
		End:     end, NoX: true})
}

func main() {
	kase("t65_tim_1s______", "a write at 999 ms, an end at 1 s", "1ns / 1ps",
		"        #999ms s = 1'b1;\n        #1ms;", "t44_time_5s_____",
		[]*c.Obj{c.Tr(999000000, "s", "1")}, 1000000000, false)

	var cross []*c.Obj
	for i := 0; i < 3000; i++ {
		cross = append(cross, c.Tr(4293000+i+1, "s", fmt.Sprint((i+1)%2)))
	}
	kase("t65_tim_cross___", "writes every 1 ns from 4.293 ms to 4.296 ms, across 2^32 ps in one page", "1ns / 1ps",
		"        #4293us;\n        repeat (3000) begin\n            #1ns s = ~s;\n        end", "t44_v_time_5ms__",
		cross, 4296000, false)

	kase("t65_tim_ns_5s___", "a write at 4.5 s under timescale 1ns / 1ns, past 2^32 ns", "1ns / 1ns",
		"        #4500ms s = 1'b1;\n        #500ms;", "t44_v_time_5ms__",
		[]*c.Obj{c.Tr(4500000000, "s", "1")}, 5000000000, false)
}
