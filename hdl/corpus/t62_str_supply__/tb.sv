// SPDX-License-Identifier: Apache-2.0

// Corpus case: a supply driver against a strong one
//
// Axis: strength. a supply driver against a strong one beside a logic, under typical, to see whether a net's drive strength, a pull source, a switch or a gate primitive leaves anything in the declaration, the hierarchy or the records.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire w; assign (supply0, supply1) w = 1'b0; assign w = s;

    initial begin
        #50 s = 1'b1;
        
        #50 $finish;
    end
endmodule
