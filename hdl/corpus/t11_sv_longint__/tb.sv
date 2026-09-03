// SPDX-License-Identifier: Apache-2.0
// Corpus case: one SystemVerilog longint
//
// Differs from t11_sv_int by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    longint s = 0;

    initial begin
        #50 s = 64'd5000000000;
        #50 $finish;
    end
endmodule
