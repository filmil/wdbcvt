// SPDX-License-Identifier: Apache-2.0
// Corpus case: a reg written twice at one time, across #0, in the first page of a shared arena that spills, Verilog.

`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg clk = 1'b0;
    reg d;

    always #1 clk = ~clk;

    initial begin
        #5 d = 1'b1;
        #0 d = 1'b0;
        #425 $finish;
    end
endmodule
