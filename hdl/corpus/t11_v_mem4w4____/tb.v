// SPDX-License-Identifier: Apache-2.0
// Corpus case: a memory of 4 words of 4 bits, for the element packing.

`timescale 1ns / 1ps

module tb;
    reg [3:0] m [0:3];

    initial begin
        m[0] = 4'h0; m[1] = 4'h0; m[2] = 4'h0; m[3] = 4'h0;
        #50 m[1] = 4'h5;
        #50 $finish;
    end
endmodule
