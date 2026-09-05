// SPDX-License-Identifier: Apache-2.0

// Corpus case: a specparam in a specify block
//
// Axis: value class. a specparam in a specify block beside a logic, to see which value class of region 17 the declaration takes, and whether any form reaches the codes 2 and 5 that no case has produced.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    specify
        specparam d = 10;
    endspecify

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
