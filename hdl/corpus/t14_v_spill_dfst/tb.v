// SPDX-License-Identifier: Apache-2.0
// Corpus case: the reg without initialiser declared before the clock, so its X record opens the spilling arena, Verilog.

`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg d;
    reg clk = 1'b0;

    always #1 clk = ~clk;

    initial begin
        #428 d = 1'b1;
        #2 $finish;
    end
endmodule
