// SPDX-License-Identifier: Apache-2.0
// Corpus case: three writes in one time step, Verilog.

`timescale 1ns / 1ps

module tb;
    reg [7:0] s;

    initial begin
        s = 8'h01;
        s = 8'h02;
        s = 8'h03;
        #50 s = 8'ha5;
        #50 $finish;
    end
endmodule
