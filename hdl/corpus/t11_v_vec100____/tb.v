// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 100 bit reg
//
// Differs from t11_v_bit_edge by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg [99:0] s = 100'h0;

    initial begin
        #50 s = {68'h5, 32'ha5};
        #50 $finish;
    end
endmodule
