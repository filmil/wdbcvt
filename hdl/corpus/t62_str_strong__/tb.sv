// SPDX-License-Identifier: Apache-2.0

// Corpus case: a strong driver over a weak one
//
// Axis: strength. a strong driver over a weak one beside a logic, under typical, to see whether a net's drive strength, a pull source, a switch or a gate primitive leaves anything in the declaration, the hierarchy or the records.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire w; assign (weak0, weak1) w = 1'b0; assign (strong0, strong1) w = s;

    initial begin
        #50 s = 1'b1;
        
        #50 $finish;
    end
endmodule
