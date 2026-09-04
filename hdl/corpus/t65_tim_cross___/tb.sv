// SPDX-License-Identifier: Apache-2.0

// Corpus case: writes every 1 ns from 4.293 ms to 4.296 ms, across 2^32 ps in one page
//
// Axis: time past 32 bits. writes every 1 ns from 4.293 ms to 4.296 ms, across 2^32 ps in one page, to see whether the tier 44 reading of the 8 byte times holds across a page, in another unit and at a second past the largest time so far

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;

    initial begin
        #4293us;
        repeat (3000) begin
            #1ns s = ~s;
        end
        $finish;
    end
endmodule
