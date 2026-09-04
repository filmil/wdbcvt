// SPDX-License-Identifier: Apache-2.0

// Corpus case: two vector fields
//
// Axis: debugging. two vector fields beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    class c_t; logic [3:0] g; logic [7:0] k; endclass
    c_t h;

    initial begin
        #50 s = 1'b1;
        h = new; h.g = 5;
        #50 $finish;
    end
endmodule
