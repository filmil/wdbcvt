// SPDX-License-Identifier: Apache-2.0
// Corpus case: a reg without an initializer
//
// Differs from t11_v_bit_edge by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg s;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
