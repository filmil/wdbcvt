// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg clk = 1'b0;
    always #25 clk = ~clk;
    reg a = 1'b0;
    reg s = 1'b0;
    initial #30 a = 1'b1;

    always @(posedge clk) s <= a;

    initial begin
        #100 $finish;
    end
endmodule
