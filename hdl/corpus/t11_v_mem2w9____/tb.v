// SPDX-License-Identifier: Apache-2.0
// Corpus case: a memory of 2 words of 9 bits, for the element packing.

`timescale 1ns / 1ps

module tb;
    reg [8:0] m [0:1];

    initial begin
        m[0] = 9'h0; m[1] = 9'h0;
        #50 m[1] = 9'h1a5;
        #50 $finish;
    end
endmodule
