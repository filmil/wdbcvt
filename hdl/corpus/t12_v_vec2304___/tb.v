// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 2304 bit reg
//
// Differs from t12_v_vec2272 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    reg [2303:0] s = 2304'h0;

    initial begin
        #50 s = {72{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
