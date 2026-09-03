// SPDX-License-Identifier: Apache-2.0
// Corpus case: 1088 bits, 34 pairs, 272 record bytes from pair 34
//
// Differs from t33_v_wsl_hi by the assignment only.

`timescale 1ns / 1ps

module tb;
    reg [2175:0] s = 2176'h0;

    initial begin
        #50 s[2175:1088] = {34{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
