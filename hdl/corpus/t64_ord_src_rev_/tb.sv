// SPDX-License-Identifier: Apache-2.0

// Corpus case: the two drivers of two bits in the other source order
//
// Axis: several partial drivers. the two drivers of two bits in the other source order beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v; assign v[3] = ~s; assign v[0] = s;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
