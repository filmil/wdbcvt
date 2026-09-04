// SPDX-License-Identifier: Apache-2.0

// Corpus case: a string named in log_wave
//
// Axis: debugging. a string named in log_wave beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    string str = "ab";

    initial begin
        #50 s = 1'b1;
        str = "xyz";
        #50 $finish;
    end
endmodule
