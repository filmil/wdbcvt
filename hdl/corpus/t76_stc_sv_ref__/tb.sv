// SPDX-License-Identifier: Apache-2.0

// Corpus case: a ref argument of a function
//
// Axis: storage classes. a ref argument of a function, to see which storage class word 28 of the instance record gives it, and whether any form gives the 5 that no case has produced.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int r = 0;
    function automatic void bump(ref int x);
        x = x + 1;
    endfunction

    initial begin
        #50 s = 1'b1;
        bump(r); bump(r);
        #50 $finish;
    end
endmodule
