// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 33 bit reg
//
// Differs from t11_v_bit_edge by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg [32:0] s = 33'h0;

    initial begin
        #50 s = 33'h1_0000_00a5;
        #50 $finish;
    end
endmodule
