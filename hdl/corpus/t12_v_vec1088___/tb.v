// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 1088 bit reg
//
// Differs from t11_v_vec100 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg [1087:0] s = 1088'h0;

    initial begin
        #50 s = {34{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
