// SPDX-License-Identifier: Apache-2.0

// Corpus case: a memory of two 4 bit words
//
// Axis: debugging. a memory of two 4 bit words beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    logic [3:0] m [0:1] = '{0, 0};

    initial begin
        #50 s = 1'b1;
        m[1] = 4'd9;
        #50 $finish;
    end
endmodule
