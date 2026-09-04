// SPDX-License-Identifier: Apache-2.0

// Corpus case: an int beside a class handle
//
// Axis: debugging. an int beside a class handle beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    class c_t; int f = 1; endclass
    c_t h;
    int i = 7;

    initial begin
        #50 s = 1'b1;
        h = new; i = 9;
        #50 $finish;
    end
endmodule
