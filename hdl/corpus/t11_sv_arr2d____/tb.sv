// SPDX-License-Identifier: Apache-2.0
// Corpus case: a packed two dimensional array
//
// Differs from t11_v_vec8 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    logic [1:0][3:0] s = '0;

    initial begin
        #50 s = 8'ha5;
        #50 $finish;
    end
endmodule
