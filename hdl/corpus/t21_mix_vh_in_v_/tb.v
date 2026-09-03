// SPDX-License-Identifier: Apache-2.0
// Corpus case: a VHDL child under a Verilog testbench.

`timescale 1ns / 1ps

module tb;
    reg x = 1'b0;
    child dut(.a(x));
    initial begin
        #50 x = 1'b1;
        #50 $finish;
    end
endmodule
