// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg b = 1'b0;
    wire c;
    wire d;

    child u(a, b, c, d);

    initial begin
        #50 a = 1'b1;
        #50 $finish;
    end
endmodule

module child(input a, input b, output c, output d);
    assign c = a & b;
    assign d = a | b;
endmodule

