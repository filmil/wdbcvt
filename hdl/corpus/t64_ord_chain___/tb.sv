// SPDX-License-Identifier: Apache-2.0

// Corpus case: a driver of one bit of a second net from a bit of the first
//
// Axis: several partial drivers. a driver of one bit of a second net from a bit of the first beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v, w; assign v[0] = s; assign w[1] = v[0];

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
