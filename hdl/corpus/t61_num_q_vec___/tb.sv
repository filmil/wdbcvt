// SPDX-License-Identifier: Apache-2.0

// Corpus case: a queue of vectors
//
// Axis: debugging. a queue of vectors beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    logic [3:0] q[$];

    initial begin
        #50 s = 1'b1;
        q.push_back(5);
        #50 $finish;
    end
endmodule
