// SPDX-License-Identifier: Apache-2.0
// Corpus case: a memory of 2 words of 40 bits, for the element packing.

`timescale 1ns / 1ps

module tb;
    reg [39:0] m [0:1];

    initial begin
        m[0] = 40'h0; m[1] = 40'h0;
        #50 m[1] = 40'ha5c3f00f12;
        #50 $finish;
    end
endmodule
