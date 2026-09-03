// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg sel = 1'b0;
    reg a = 1'b0;
    reg b = 1'b0;

    child u(.sel(sel), .a(a), .b(b));

    initial begin
        #10 a = 1'b1;
        #0 b = 1'b1;
        #20 sel = 1'b1;
        #30 sel = 1'b0;
        #40 $finish;
    end
endmodule

module child(input sel, input a, input b);
    wire i;
    assign i = sel ? a : b;
endmodule
