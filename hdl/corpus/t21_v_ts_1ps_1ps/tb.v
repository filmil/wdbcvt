// SPDX-License-Identifier: Apache-2.0
// Corpus case: timescale 1ps / 1ps, Verilog.

`timescale 1ps / 1ps

module tb;
    reg s = 1'b0;
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
