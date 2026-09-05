// SPDX-License-Identifier: Apache-2.0

// Corpus case: a class with an associative array field
//
// Axis: debugging. a class with an associative array field beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    class c_t; int a[string]; endclass
    c_t h;

    initial begin
        #50 s = 1'b1;
        h = new;
        #50 $finish;
    end
endmodule
