// SPDX-License-Identifier: Apache-2.0
// Corpus case: an always block and a named block.

`timescale 1ns / 1ps

module tb;
    reg clk = 1'b0;
    reg s = 1'b0;

    always #25 clk = ~clk;

    always @(posedge clk) begin : blk
        s <= ~s;
    end

    initial begin
        #100 $finish;
    end
endmodule
