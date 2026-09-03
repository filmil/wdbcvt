// SPDX-License-Identifier: Apache-2.0
// Corpus case: an enum of logic [3:0] with explicit values
//
// Differs from t11_sv_enum by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    typedef enum logic [3:0] {A = 4'd1, B = 4'd5, C = 4'd9} state_t;
    state_t s = A;

    initial begin
        #50 s = C;
        #50 $finish;
    end
endmodule
