// SPDX-License-Identifier: Apache-2.0
// Corpus case: bits 16 to 2415 of a 4800 bit reg, inside 76 pairs
//
// Differs from t33_v_wsl_hi by the assignment only.

`timescale 1ns / 1ps

module tb;
    reg [4799:0] s = 4800'h0;

    initial begin
        #50 s[2415:16] = {75{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
