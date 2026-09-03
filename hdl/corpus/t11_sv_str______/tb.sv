// SPDX-License-Identifier: Apache-2.0
// Corpus case: a SystemVerilog string
//
// Differs from t11_sv_logic by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    string s = "ab";

    initial begin
        #50 s = "xyz";
        #50 $finish;
    end
endmodule
