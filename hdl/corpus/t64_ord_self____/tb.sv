// SPDX-License-Identifier: Apache-2.0

// Corpus case: a driver of one bit from another bit of the same net
//
// Axis: several partial drivers. a driver of one bit from another bit of the same net beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v; assign v[0] = s; assign v[1] = v[0];

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
