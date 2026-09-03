// SPDX-License-Identifier: Apache-2.0
// Corpus case: a SystemVerilog shortreal
//
// Differs from t11_v_real by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    shortreal s = 0.0;

    initial begin
        #50 s = 1.5;
        #50 $finish;
    end
endmodule
