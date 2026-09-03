// SPDX-License-Identifier: Apache-2.0
// Corpus case: an eight bit reg with the range [-4:3]
//
// Differs from t11_v_vec8 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg [-4:3] s = 8'h00;

    initial begin
        #50 s = 8'ha5;
        #50 $finish;
    end
endmodule
