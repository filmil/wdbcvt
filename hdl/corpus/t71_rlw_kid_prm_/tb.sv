// SPDX-License-Identifier: Apache-2.0

// Corpus case: a real parameter of a child module
//
// Axis: the width of a real parameter. a real parameter of a child module beside a logic, to see where the 16 bits a real parameter declares comes from, when a real variable declares 32 and both hold one float64.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    kid u();

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module kid;
    parameter real R = 1.5;
endmodule
