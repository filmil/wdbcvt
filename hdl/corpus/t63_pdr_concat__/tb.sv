// SPDX-License-Identifier: Apache-2.0

// Corpus case: two scalar nets driven through a concatenation
//
// Axis: partial drivers. two scalar nets driven through a concatenation beside a logic, under typical, to see whether a driver of a bit, a slice or a port bound to part of a net records the whole net or the part.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire a, b; assign {a, b} = {s, ~s};

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
