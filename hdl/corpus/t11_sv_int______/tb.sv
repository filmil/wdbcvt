// SPDX-License-Identifier: Apache-2.0
// Corpus case: one SystemVerilog int
//
// Differs from t11_sv_logic by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    int s = 0;

    initial begin
        #50 s = 165;
        #50 $finish;
    end
endmodule
