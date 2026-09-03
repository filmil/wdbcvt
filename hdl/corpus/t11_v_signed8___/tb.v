// SPDX-License-Identifier: Apache-2.0
// Corpus case: a signed eight bit reg
//
// Differs from t11_v_bit_edge by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg signed [7:0] s = 8'h00;

    initial begin
        #50 s = -8'sd91;
        #50 $finish;
    end
endmodule
