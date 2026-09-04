// SPDX-License-Identifier: Apache-2.0

// Corpus case: a class with a handle field
//
// Axis: debugging. a class with a handle field beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    class b_t; int g; endclass
    class c_t; b_t hb; endclass
    c_t h;

    initial begin
        #50 s = 1'b1;
        h = new;
        #50 $finish;
    end
endmodule
