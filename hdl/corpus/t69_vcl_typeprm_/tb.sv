// SPDX-License-Identifier: Apache-2.0

// Corpus case: a type parameter used by a variable
//
// Axis: value class. a type parameter used by a variable beside a logic, to see which value class of region 17 the declaration takes, and whether any form reaches the codes 2 and 5 that no case has produced.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    parameter type T = logic [7:0];
    T v = 5;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
