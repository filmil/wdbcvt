// SPDX-License-Identifier: Apache-2.0
// Corpus case: a memory of 3 words of 5 bits, for the element packing.

`timescale 1ns / 1ps

module tb;
    reg [4:0] m [0:2];

    initial begin
        m[0] = 5'h0; m[1] = 5'h0; m[2] = 5'h0;
        #50 m[1] = 5'h5;
        #50 $finish;
    end
endmodule
