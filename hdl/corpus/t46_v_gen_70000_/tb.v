// SPDX-License-Identifier: Apache-2.0
// Corpus case: scale. 70000 generate iterations each declaring a reg, to see whether a scope count or index above 65535 is stored whole.

`timescale 1ns / 1ps

module tb;
    genvar i;
    generate
        for (i = 0; i < 70000; i = i + 1) begin : g
            reg r = 1'b0;
        end
    endgenerate

    initial begin
        #5 g[0].r = 1'b1;
        #10 g[69999].r = 1'b1;
        #5 $finish;
    end
endmodule
