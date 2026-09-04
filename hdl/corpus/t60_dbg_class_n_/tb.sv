// SPDX-License-Identifier: Apache-2.0

// Corpus case: a class handle constructed at time 0
//
// Axis: debugging. a class handle constructed at time 0 beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    class c_t; int f = 1; endclass
    c_t h = new;

    initial begin
        #50 s = 1'b1;
        h.f = 5;
        #50 $finish;
    end
endmodule
