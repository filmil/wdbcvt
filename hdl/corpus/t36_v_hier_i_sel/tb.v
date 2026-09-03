// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg b = 1'b0;
    wire w;

    child u(.a(a), .b(b), .w(w));

    initial begin
        #10 a = 1'b1;
        #0 b = 1'b1;
        #90 $finish;
    end
endmodule

module child(input a, input b, output w);
    reg sel = 1'b0;
    wire i;
    assign i = sel ? a : b;
    assign w = i;

    initial begin
        #30 sel = 1'b1;
        #30 sel = 1'b0;
        #40 $finish;
    end
endmodule
