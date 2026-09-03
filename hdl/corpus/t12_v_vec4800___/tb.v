// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 4800 bit reg
//
// Differs from t12_v_vec2304 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg [4799:0] s = 4800'h0;

    initial begin
        #50 s = {150{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
