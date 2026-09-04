// SPDX-License-Identifier: Apache-2.0

// Corpus case: the top 400 bits of a 2400 bit net driven
//
// Axis: partial drivers. the top 400 bits of a 2400 bit net driven beside a logic, under typical, to see whether a driver of a bit, a slice or a port bound to part of a net records the whole net or the part.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [2399:0] v; assign v[2399:2000] = {400{s}};

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
