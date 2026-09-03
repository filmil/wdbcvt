// SPDX-License-Identifier: Apache-2.0
// Corpus case: one Verilog reg with one edge at 5 ms.

`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;

    initial begin
        #5000000 s = 1'b1;
        #5000000 $finish;
    end
endmodule
