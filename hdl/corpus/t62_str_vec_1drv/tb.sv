// SPDX-License-Identifier: Apache-2.0

// Corpus case: a vector with one driver
//
// Axis: strength. a vector with one driver beside a logic, under typical, to see whether a net's drive strength, a pull source, a switch or a gate primitive leaves anything in the declaration, the hierarchy or the records.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v; assign v = s ? 4'b1101 : 4'b0000;

    initial begin
        #50 s = 1'b1;
        
        #50 $finish;
    end
endmodule
