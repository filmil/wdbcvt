// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg b = 1'b1;
    wire w;

    assign w = a & b;

    initial begin
        #20 b = 1'b0;
        #10 a = 1'b1;
        #30 a = 1'b0;
        #40 $finish;
    end
endmodule
