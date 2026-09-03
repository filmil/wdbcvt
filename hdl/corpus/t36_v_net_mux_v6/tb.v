// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg sel = 1'b0;
    reg [5:0] a = 6'd0;
    reg [5:0] b = 6'd0;
    wire [5:0] w;

    assign w = sel ? a : b;

    initial begin
        #10 a = 6'd1;
        #0 b = 6'd1;
        #20 sel = 1'b1;
        #30 sel = 1'b0;
        #40 $finish;
    end
endmodule
