// SPDX-License-Identifier: Apache-2.0
// Corpus case: 430 transitions, Verilog.

`timescale 1ns / 1ps

module tb;
    reg clk = 1'b0;

    always #1 clk = ~clk;

    initial begin
        #430 $finish;
    end
endmodule
