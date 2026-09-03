// SPDX-License-Identifier: Apache-2.0
// Corpus case: a Verilog integer
//
// Differs from t11_v_bit_edge by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    integer s = 0;

    initial begin
        #50 s = 165;
        #50 $finish;
    end
endmodule
