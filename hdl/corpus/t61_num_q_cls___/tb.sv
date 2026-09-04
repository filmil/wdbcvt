// SPDX-License-Identifier: Apache-2.0

// Corpus case: a queue of class handles
//
// Axis: debugging. a queue of class handles beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    class c_t; int f = 1; endclass
    c_t q[$];

    initial begin
        #50 s = 1'b1;
        q.push_back(null);
        #50 $finish;
    end
endmodule
