// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 2272 bit reg
//
// Differs from t12_v_vec1089 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg [2271:0] s = 2272'h0;

    initial begin
        #50 s = {71{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
