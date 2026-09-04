// SPDX-License-Identifier: Apache-2.0

// Corpus case: a tri driven by s
//
// Axis: strength. a tri driven by s beside a logic, under typical, to see whether a net's drive strength, a pull source, a switch or a gate primitive leaves anything in the declaration, the hierarchy or the records.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    tri w; assign w = s;

    initial begin
        #50 s = 1'b1;
        
        #50 $finish;
    end
endmodule
