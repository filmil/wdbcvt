// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    logic clk = 1'b0;
    logic d = 1'b0;
    logic q = 1'b0;
    always #25 clk = ~clk;
    always_ff @(posedge clk) q <= d;
    initial begin
        #10 d = 1'b1;
        #90 $finish;
    end
endmodule
