// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg clk = 1'b0;
    always #25 clk = ~clk;
    reg s = 1'b0;

    initial begin
        @(posedge clk) s <= 1'b0;
        @(posedge clk) s <= 1'b0;
    end

    initial begin
        #100 $finish;
    end
endmodule
