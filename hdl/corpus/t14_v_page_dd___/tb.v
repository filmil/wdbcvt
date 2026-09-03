// SPDX-License-Identifier: Apache-2.0
// Corpus case: a reg written twice at one time, across #0, in a shared arena that stays in one page, Verilog.

`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg clk = 1'b0;
    reg d;

    always #1 clk = ~clk;

    initial begin
        #190 d = 1'b1;
        #0 d = 1'b0;
        #10 $finish;
    end
endmodule
