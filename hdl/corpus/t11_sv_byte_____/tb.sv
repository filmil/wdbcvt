// SPDX-License-Identifier: Apache-2.0
// Corpus case: one SystemVerilog byte
//
// Differs from t11_sv_int by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    byte s = 0;

    initial begin
        #50 s = 8'ha5;
        #50 $finish;
    end
endmodule
