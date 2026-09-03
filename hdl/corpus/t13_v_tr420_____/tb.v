// SPDX-License-Identifier: Apache-2.0
// Corpus case: 420 transitions, Verilog.

`timescale 1ns / 1ps

module tb;
    reg clk = 1'b0;

    always #1 clk = ~clk;

    initial begin
        #420 $finish;
    end
endmodule
