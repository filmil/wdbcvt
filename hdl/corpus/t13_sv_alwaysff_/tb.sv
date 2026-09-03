// SPDX-License-Identifier: Apache-2.0
// Corpus case: always_ff and always_comb, SystemVerilog.

`timescale 1ns / 1ps

module tb;
    logic clk = 1'b0;
    logic d = 1'b0;
    logic q = 1'b0;
    logic n;

    always #25 clk = ~clk;
    always_ff @(posedge clk) q <= d;
    always_comb n = ~q;

    initial begin
        #10 d = 1'b1;
        #90 $finish;
    end
endmodule
