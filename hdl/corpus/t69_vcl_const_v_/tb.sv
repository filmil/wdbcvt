// SPDX-License-Identifier: Apache-2.0

// Corpus case: a const vector
//
// Axis: value class. a const vector beside a logic, to see which value class of region 17 the declaration takes, and whether any form reaches the codes 2 and 5 that no case has produced.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    const logic [7:0] k = 8'd5;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
