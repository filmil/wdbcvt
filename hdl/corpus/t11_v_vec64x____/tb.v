// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 64 bit reg, then one X bit in its upper word.

`timescale 1ns / 1ps

module tb;
    reg [63:0] s = 64'h0;

    initial begin
        #50 s = 64'hDEADBEEFCAFEBABE;
        #25 s[40] = 1'bx;
        #25 $finish;
    end
endmodule
