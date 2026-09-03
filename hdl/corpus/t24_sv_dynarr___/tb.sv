// SPDX-License-Identifier: Apache-2.0
// Corpus case: a dynamic array: int d[].

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int d[];

    initial begin
        #50 s = 1'b1;
        d = new[4];
        d[1] = 5;
        #50 $finish;
    end
endmodule
