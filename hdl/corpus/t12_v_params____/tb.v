// SPDX-License-Identifier: Apache-2.0
// Corpus case: parameters of every kind, Verilog.

`timescale 1ns / 1ps

module tb;
    child #(.K(7)) dut();

    initial begin
        #100 $finish;
    end
endmodule
