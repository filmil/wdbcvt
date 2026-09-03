// SPDX-License-Identifier: Apache-2.0
// Corpus case: two instances with different parameter values, Verilog.

`timescale 1ns / 1ps

module tb;
    child #(.K(7)) dut();
    child #(.K(9)) dut2();
    initial begin
        #100 $finish;
    end
endmodule
