// SPDX-License-Identifier: Apache-2.0

// Corpus case: a child parameter set by defparam
//
// Axis: value class. a child parameter set by defparam beside a logic, to see which value class of region 17 the declaration takes, and whether any form reaches the codes 2 and 5 that no case has produced.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    kid u();
    defparam u.P = 5;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module kid;
    parameter P = 1;
    logic [7:0] v = P;
endmodule
