// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg sel = 1'b0;
    reg a = 1'b0;
    reg b = 1'b0;
    wire w;

    child u(.sel(sel), .a(a), .b(b), .w(w));

    initial begin
        #10 a = 1'b1;
        #0 b = 1'b1;
        #20 sel = 1'b1;
        #30 sel = 1'b0;
        #40 $finish;
    end
endmodule

module child(input sel, input a, input b, output w);
    assign w = sel ? a : b;
endmodule
