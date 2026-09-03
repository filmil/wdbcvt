// SPDX-License-Identifier: Apache-2.0
// Corpus case: a wire held at 0 by its assignment, Verilog.

`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;
    wire w;

    assign w = s & 1'b0;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
