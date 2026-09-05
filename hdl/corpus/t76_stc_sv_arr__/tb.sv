// SPDX-License-Identifier: Apache-2.0

// Corpus case: an unpacked array local of a function
//
// Axis: storage classes. an unpacked array local of a function, to see which storage class word 28 of the instance record gives it, and whether any form gives the 5 that no case has produced.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int r = 0;
    function automatic int sum2();
        int a[2];
        a[0] = 1;
        a[1] = 1;
        return a[0] + a[1];
    endfunction

    initial begin
        #50 s = 1'b1;
        r = sum2();
        #50 $finish;
    end
endmodule
