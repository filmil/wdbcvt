// SPDX-License-Identifier: Apache-2.0

// Corpus case: a write at 999 ms, an end at 1 s
//
// Axis: time past 32 bits. a write at 999 ms, an end at 1 s, to see whether the tier 44 reading of the 8 byte times holds across a page, in another unit and at a second past the largest time so far

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;

    initial begin
        #999ms s = 1'b1;
        #1ms;
        $finish;
    end
endmodule
