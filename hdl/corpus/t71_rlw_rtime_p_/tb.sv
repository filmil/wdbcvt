// SPDX-License-Identifier: Apache-2.0

// Corpus case: a realtime parameter
//
// Axis: the width of a real parameter. a realtime parameter beside a logic, to see where the 16 bits a real parameter declares comes from, when a real variable declares 32 and both hold one float64.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    parameter realtime R = 1.5;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
