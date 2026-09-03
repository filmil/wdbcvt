// SPDX-License-Identifier: Apache-2.0
// Corpus case: a memory of 2 words of 16 bits, for the element packing.

`timescale 1ns / 1ps

module tb;
    reg [15:0] m [0:1];

    initial begin
        m[0] = 16'h0; m[1] = 16'h0;
        #50 m[1] = 16'ha5c3;
        #50 $finish;
    end
endmodule
