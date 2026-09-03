// SPDX-License-Identifier: Apache-2.0
// Corpus case: a wor net with two drivers, Verilog.

`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;
    reg t = 1'b0;
    wor w;
    assign w = s;
    assign w = t;
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
