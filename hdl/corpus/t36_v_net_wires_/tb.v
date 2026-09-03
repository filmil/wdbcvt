// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg sel = 1'b0;
    reg a = 1'b0;
    reg b = 1'b0;
    wire sel2 = sel;
    wire a2 = a;
    wire b2 = b;
    wire w;

    assign w = sel2 ? a2 : b2;

    initial begin
        #10 a = 1'b1;
        #0 b = 1'b1;
        #20 sel = 1'b1;
        #30 sel = 1'b0;
        #40 $finish;
    end
endmodule
