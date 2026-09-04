// SPDX-License-Identifier: Apache-2.0

// Corpus case: the same child without a specify block
//
// Axis: process scopes. the same child without a specify block beside a logic, under typical, to see what scope the construct leaves and what it declares.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire w; kid u(.i(s), .o(w));

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module kid(input i, output o);
    assign o = i;

endmodule
