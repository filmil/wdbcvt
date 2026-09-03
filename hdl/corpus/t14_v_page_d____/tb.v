// SPDX-License-Identifier: Apache-2.0
// Corpus case: a reg without initialiser sharing an arena with a clock that stays in one page, Verilog.

`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg clk = 1'b0;
    reg d;

    always #1 clk = ~clk;

    initial begin
        #190 d = 1'b1;
        #10 $finish;
    end
endmodule
