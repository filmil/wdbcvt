// SPDX-License-Identifier: Apache-2.0
// Corpus case: a generate for loop, Verilog.

`timescale 1ns / 1ps

module tb;
    genvar i;
    generate
        for (i = 0; i < 2; i = i + 1) begin : g
            child dut();
        end
    endgenerate

    initial begin
        #100 $finish;
    end
endmodule
