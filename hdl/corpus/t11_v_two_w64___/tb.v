// SPDX-License-Identifier: Apache-2.0
// Corpus case: two regs of different widths, for the handle spacing.

`timescale 1ns / 1ps

module tb;
    reg [63:0] p = 64'h0;
    reg q = 1'b0;

    initial begin
        #50 p = 64'h1;
        #10 q = 1'b1;
        #40 $finish;
    end
endmodule
