// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    logic d = 1'b0;
    logic n;
    always_comb n = ~d;
    initial begin
        #10 d = 1'b1;
        #90 $finish;
    end
endmodule
