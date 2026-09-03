// SPDX-License-Identifier: Apache-2.0
// Corpus case: 1089 bits, 35 pairs, 280 record bytes from pair 34
//
// Differs from t33_v_wsl_272 by the assignment only.

`timescale 1ns / 1ps

module tb;
    reg [2177:0] s = 2178'h0;

    initial begin
        #50 s[2177:1089] = {35{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
