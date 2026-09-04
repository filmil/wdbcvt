// SPDX-License-Identifier: Apache-2.0

// Corpus case: bits 0 and 3 driven from two assigns
//
// Axis: partial drivers. bits 0 and 3 driven from two assigns beside a logic, under typical, to see whether a driver of a bit, a slice or a port bound to part of a net records the whole net or the part.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v; assign v[0] = s; assign v[3] = ~s;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
