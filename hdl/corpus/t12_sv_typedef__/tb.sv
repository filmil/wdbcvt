// SPDX-License-Identifier: Apache-2.0
// Corpus case: a vector declared through a typedef
//
// Differs from t11_v_vec8 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    typedef logic [7:0] byte_t;
    byte_t s = '0;

    initial begin
        #50 s = 8'ha5;
        #50 $finish;
    end
endmodule
