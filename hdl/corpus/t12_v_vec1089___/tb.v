// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 1089 bit reg
//
// Differs from t12_v_vec1088 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg [1088:0] s = 1089'h0;

    initial begin
        #50 s = {35{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
