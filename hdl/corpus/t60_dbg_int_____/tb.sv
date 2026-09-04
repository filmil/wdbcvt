// SPDX-License-Identifier: Apache-2.0

// Corpus case: an int
//
// Axis: debugging. an int beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int i = 7;

    initial begin
        #50 s = 1'b1;
        i = 9;
        #50 $finish;
    end
endmodule
