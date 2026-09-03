// SPDX-License-Identifier: Apache-2.0
// Corpus case: fork and join.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    logic f = 1'b0;

    initial begin
        #50 s = 1'b1;
        fork
            #10 f = 1'b1;
            #20 f = 1'b0;
        join
        #50 $finish;
    end
endmodule
