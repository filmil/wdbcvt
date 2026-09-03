// SPDX-License-Identifier: Apache-2.0
// Corpus case: a memory of four rows of 2400 bits, row 1 written.

`timescale 1ns / 1ps

module tb;
    reg [2399:0] m [0:3];

    initial begin
        m[0] = 2400'h0;
        m[1] = 2400'h0;
        m[2] = 2400'h0;
        m[3] = 2400'h0;
        #50 m[1] = {75{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
