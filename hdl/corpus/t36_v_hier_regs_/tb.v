// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    child u();
endmodule

module child;
    reg sel = 1'b0;
    reg a = 1'b0;
    reg b = 1'b0;
    wire i;

    assign i = sel ? a : b;

    initial begin
        #10 a = 1'b1;
        #0 b = 1'b1;
        #20 sel = 1'b1;
        #30 sel = 1'b0;
        #40 $finish;
    end
endmodule
