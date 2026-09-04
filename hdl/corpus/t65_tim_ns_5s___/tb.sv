// SPDX-License-Identifier: Apache-2.0

// Corpus case: a write at 4.5 s under timescale 1ns / 1ns, past 2^32 ns
//
// Axis: time past 32 bits. a write at 4.5 s under timescale 1ns / 1ns, past 2^32 ns, to see whether the tier 44 reading of the 8 byte times holds across a page, in another unit and at a second past the largest time so far

`timescale 1ns / 1ns

module tb;
    logic s = 1'b0;

    initial begin
        #4500ms s = 1'b1;
        #500ms;
        $finish;
    end
endmodule
