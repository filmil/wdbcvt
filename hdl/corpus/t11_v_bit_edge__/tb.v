// SPDX-License-Identifier: Apache-2.0
// Corpus case: one Verilog reg with one transition.
//
// The Tier 11 baseline. It is t1_bit_one_edge written in Verilog and
// nothing else, to find out whether the source language reaches the
// database.
//
// Differs from t1_bit_one_edge by the source language only.

`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
