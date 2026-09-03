// SPDX-License-Identifier: Apache-2.0
// Corpus case: a reg without initialiser written in the first page of the shared arena, which then spills, Verilog.

`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg clk = 1'b0;
    reg d;

    always #1 clk = ~clk;

    initial begin
        #5 d = 1'b1;
        #425 $finish;
    end
endmodule
