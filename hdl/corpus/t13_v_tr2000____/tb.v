// SPDX-License-Identifier: Apache-2.0
// Corpus case: two thousand transitions, Verilog.

`timescale 1ns / 1ps

module tb;
    reg clk = 1'b0;

    always #1 clk = ~clk;

    initial begin
        #2000 $finish;
    end
endmodule
