// SPDX-License-Identifier: Apache-2.0

// Corpus case: an associative array of int keyed by byte
//
// Axis: debugging. an associative array of int keyed by byte beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int a[byte];

    initial begin
        #50 s = 1'b1;
        a[8'd3] = 5;
        #50 $finish;
    end
endmodule
