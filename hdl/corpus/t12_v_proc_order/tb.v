// SPDX-License-Identifier: Apache-2.0
// Corpus case: process numbering across two modules, Verilog.

`timescale 1ns / 1ps

module tb;
    reg t = 1'b0;
    wire u;

    assign u = ~t;

    child dut();

    initial begin
        #50 t = 1'b1;
        #50 $finish;
    end
endmodule
