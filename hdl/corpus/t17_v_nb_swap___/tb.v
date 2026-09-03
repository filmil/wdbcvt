// SPDX-License-Identifier: Apache-2.0
// Corpus case: a nonblocking swap, Verilog.

`timescale 1ns / 1ps

module tb;
    reg a = 1'b0;
    reg b = 1'b1;

    initial begin
        #50 a <= b;
        b <= a;
        #50 $finish;
    end
endmodule
