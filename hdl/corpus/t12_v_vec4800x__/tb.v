// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 4800 bit reg, then one X bit in its middle.

`timescale 1ns / 1ps

module tb;
    reg [4799:0] s = 4800'h0;

    initial begin
        #50 s = {150{32'ha5c3f00f}};
        #25 s[2400] = 1'bx;
        #25 $finish;
    end
endmodule
