// SPDX-License-Identifier: Apache-2.0

// Corpus case: a string local of a function
//
// Axis: storage classes. a string local of a function, to see which storage class word 28 of the instance record gives it, and whether any form gives the 5 that no case has produced.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int r = 0;
    function automatic int len();
        string t;
        t = "ab";
        return t.len();
    endfunction

    initial begin
        #50 s = 1'b1;
        r = len();
        #50 $finish;
    end
endmodule
