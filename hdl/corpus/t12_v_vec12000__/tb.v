// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 12000 bit reg
//
// Differs from t12_v_vec4800 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg [11999:0] s = 12000'h0;

    initial begin
        #50 s = {375{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
