// SPDX-License-Identifier: Apache-2.0
// Corpus case: the wire connected to an input port that nothing inside reads, Verilog.

`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;
    wire w;

    assign w = s;
    child dut(.i(w));
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
