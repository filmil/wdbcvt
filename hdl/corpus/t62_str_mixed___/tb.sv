// SPDX-License-Identifier: Apache-2.0

// Corpus case: a strong0 weak1 driver against a pull1
//
// Axis: strength. a strong0 weak1 driver against a pull1 beside a logic, under typical, to see whether a net's drive strength, a pull source, a switch or a gate primitive leaves anything in the declaration, the hierarchy or the records.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire w; assign (strong0, weak1) w = s; assign (pull0, pull1) w = 1'b1;

    initial begin
        #50 s = 1'b1;
        
        #50 $finish;
    end
endmodule
