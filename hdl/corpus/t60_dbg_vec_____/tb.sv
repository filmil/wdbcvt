// SPDX-License-Identifier: Apache-2.0

// Corpus case: a 4 bit logic vector
//
// Axis: debugging. a 4 bit logic vector beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    logic [3:0] v = 4'b0000;

    initial begin
        #50 s = 1'b1;
        v = 4'b0101;
        #50 $finish;
    end
endmodule
