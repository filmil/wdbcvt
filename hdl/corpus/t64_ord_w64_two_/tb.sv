// SPDX-License-Identifier: Apache-2.0

// Corpus case: two drivers in the two pairs of a 64 bit net
//
// Axis: several partial drivers. two drivers in the two pairs of a 64 bit net beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [63:0] v; assign v[0] = s; assign v[63] = ~s;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
