// SPDX-License-Identifier: Apache-2.0
// Corpus case: a generate for loop with a reg in the block, Verilog.

`timescale 1ns / 1ps

module tb;
    genvar i;
    generate
        for (i = 0; i < 2; i = i + 1) begin : g
            reg r = 1'b0;
            initial begin
                #50 r = 1'b1;
            end
        end
    endgenerate

    initial begin
        #100 $finish;
    end
endmodule
