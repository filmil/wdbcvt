// SPDX-License-Identifier: Apache-2.0
// Corpus case: a reg in an if generate block, Verilog.

`timescale 1ns / 1ps

module tb;
    generate
        if (1) begin : g
            reg r = 1'b0;
        end
    endgenerate

    initial begin
        #50 g.r = 1'b1;
        #50 $finish;
    end
endmodule
