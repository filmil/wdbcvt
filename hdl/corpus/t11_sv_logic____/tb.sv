// SPDX-License-Identifier: Apache-2.0
// Corpus case: one SystemVerilog logic with one transition
//
// Differs from t11_v_bit_edge__ by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
