// SPDX-License-Identifier: Apache-2.0

// Corpus case: an associative array keyed by int
//
// Axis: debugging. an associative array keyed by int beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int a[int];

    initial begin
        #50 s = 1'b1;
        a[3] = 5;
        #50 $finish;
    end
endmodule
