// SPDX-License-Identifier: Apache-2.0
// Corpus case: an eight bit reg with an ascending range
//
// Differs from t11_v_bit_edge by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg [0:7] s = 8'h00;

    initial begin
        #50 s = 8'ha5;
        #50 $finish;
    end
endmodule
