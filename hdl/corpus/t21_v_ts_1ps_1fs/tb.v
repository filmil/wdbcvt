// SPDX-License-Identifier: Apache-2.0
// Corpus case: timescale 1ps / 1fs with a half picosecond delay, Verilog.

`timescale 1ps / 1fs

module tb;
    reg s = 1'b0;
    initial begin
        #50.5 s = 1'b1;
        #49.5 $finish;
    end
endmodule
