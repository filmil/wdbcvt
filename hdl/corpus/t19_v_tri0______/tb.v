// SPDX-License-Identifier: Apache-2.0
// Corpus case: a tri0 net with no driver, Verilog.

`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;
    tri0 w;
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
