// SPDX-License-Identifier: Apache-2.0
// Corpus case: a string parameter, Verilog.

`timescale 1ns / 1ps

module tb;
    child #(.P("hello")) dut();

    initial begin
        #100 $finish;
    end
endmodule
