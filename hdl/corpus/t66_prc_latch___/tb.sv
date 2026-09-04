// SPDX-License-Identifier: Apache-2.0

// Corpus case: an always_latch
//
// Axis: process scopes. an always_latch beside a logic, under typical, to see what scope the construct leaves and what it declares.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    logic q;
    always_latch if (s) q <= 1'b1;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
