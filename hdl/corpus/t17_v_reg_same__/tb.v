// SPDX-License-Identifier: Apache-2.0
// Corpus case: a write of the value already held, Verilog.

`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;

    initial begin
        #50 s = 1'b0;
        #50 $finish;
    end
endmodule
