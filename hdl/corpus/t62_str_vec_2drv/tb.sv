// SPDX-License-Identifier: Apache-2.0

// Corpus case: a vector with two drivers
//
// Axis: strength. a vector with two drivers beside a logic, under typical, to see whether a net's drive strength, a pull source, a switch or a gate primitive leaves anything in the declaration, the hierarchy or the records.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v; assign v = s ? 4'bzz01 : 4'b0000; assign v = 4'bz1zz;

    initial begin
        #50 s = 1'b1;
        
        #50 $finish;
    end
endmodule
