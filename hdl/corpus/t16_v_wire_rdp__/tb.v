// SPDX-License-Identifier: Apache-2.0
// Corpus case: the wire read by an always process, Verilog.

`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;
    wire w;
    reg q;
    assign w = s;
    always @(w) q = w;
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
