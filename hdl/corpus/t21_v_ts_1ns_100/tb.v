// SPDX-License-Identifier: Apache-2.0
// Corpus case: timescale 1ns / 100ps with a fractional delay, Verilog.

`timescale 1ns / 100ps

module tb;
    reg s = 1'b0;
    initial begin
        #50.55 s = 1'b1;
        #49.45 $finish;
    end
endmodule
