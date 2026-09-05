// SPDX-License-Identifier: Apache-2.0

// Corpus case: an array of two real parameters
//
// Axis: the width of a real parameter. an array of two real parameters beside a logic, to see where the 16 bits a real parameter declares comes from, when a real variable declares 32 and both hold one float64.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    parameter real R [0:1] = '{1.5, 2.5};
    real v = R[1];

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
