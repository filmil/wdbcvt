// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg b = 1'b0;
    wire c;
    wire d;

    child u(.a(a), .b(b), .c(c), .d(d));

    initial begin
        #50 a = 1'b1;
        #50 $finish;
    end
endmodule

module child(a, b, c, d);
    output d;
    output c;
    input b;
    input a;
    assign c = a & b;
    assign d = a | b;
endmodule

