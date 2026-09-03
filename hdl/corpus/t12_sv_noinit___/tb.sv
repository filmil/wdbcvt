// SPDX-License-Identifier: Apache-2.0
// Corpus case: a logic without an initializer
//
// Differs from t11_sv_logic by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    logic s;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
