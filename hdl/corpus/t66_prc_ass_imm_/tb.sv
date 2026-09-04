// SPDX-License-Identifier: Apache-2.0

// Corpus case: an immediate assertion in an always block
//
// Axis: process scopes. an immediate assertion in an always block beside a logic, under typical, to see what scope the construct leaves and what it declares.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    always @(s) assert (s !== 1'bx);

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
