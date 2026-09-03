// SPDX-License-Identifier: Apache-2.0
// Corpus case: a reg without initialiser first written after the shared arena spilled into a second page, Verilog.

`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg clk = 1'b0;
    reg d;

    always #1 clk = ~clk;

    initial begin
        #428 d = 1'b1;
        #2 $finish;
    end
endmodule
