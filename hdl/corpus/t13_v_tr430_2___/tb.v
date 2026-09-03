// SPDX-License-Identifier: Apache-2.0
// Corpus case: 430 transitions and a second reg, Verilog.

`timescale 1ns / 1ps

module tb;
    reg clk = 1'b0;
    reg d = 1'b0;

    always #1 clk = ~clk;

    initial begin
        #5 d = 1'b1;
        #425 $finish;
    end
endmodule
