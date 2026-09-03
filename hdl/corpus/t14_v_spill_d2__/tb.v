// SPDX-License-Identifier: Apache-2.0
// Corpus case: a reg without initialiser written in both pages of the shared arena, Verilog.

`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg clk = 1'b0;
    reg d;

    always #1 clk = ~clk;

    initial begin
        #5 d = 1'b1;
        #423 d = 1'b0;
        #2 $finish;
    end
endmodule
