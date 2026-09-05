// SPDX-License-Identifier: Apache-2.0

// Corpus case: an enumeration from a cast of 'x
//
// Axis: value class. an enumeration from a cast of 'x beside a logic, to see which value class of region 17 the declaration takes, and whether any form reaches the codes 2 and 5 that no case has produced.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    typedef enum logic [1:0] {A, B} e_t;
    e_t e = e_t'('x);

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
