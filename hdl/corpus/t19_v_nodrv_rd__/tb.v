// SPDX-License-Identifier: Apache-2.0
// Corpus case: a wire with no driver read by a continuous assignment, Verilog.

`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;
    wire w;
    wire w2;
    assign w2 = w;
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
