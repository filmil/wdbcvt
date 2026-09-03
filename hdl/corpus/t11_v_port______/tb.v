// SPDX-License-Identifier: Apache-2.0
// Corpus case: an input and an output port, Verilog.

`timescale 1ns / 1ps

module tb;
    reg x = 1'b0;
    wire y;

    child dut(.a(x), .b(y));

    initial begin
        #50 x = 1'b1;
        #50 $finish;
    end
endmodule
