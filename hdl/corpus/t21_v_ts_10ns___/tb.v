// SPDX-License-Identifier: Apache-2.0
// Corpus case: timescale 10ns / 1ns, Verilog.

`timescale 10ns / 1ns

module tb;
    reg s = 1'b0;
    initial begin
        #5 s = 1'b1;
        #5 $finish;
    end
endmodule
