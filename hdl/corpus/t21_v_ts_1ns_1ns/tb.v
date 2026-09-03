// SPDX-License-Identifier: Apache-2.0
// Corpus case: timescale 1ns / 1ns, Verilog.

`timescale 1ns / 1ns

module tb;
    reg s = 1'b0;
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
