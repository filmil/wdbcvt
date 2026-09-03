// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg clk = 1'b0;
    always #25 clk = ~clk;
    reg a = 1'b0;
    reg b = 1'b0;
    wire w;

    always @(posedge clk) a <= ~a;
    assign w = a & b;

    initial begin
        #100 $finish;
    end
endmodule
