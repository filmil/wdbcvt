// SPDX-License-Identifier: Apache-2.0
// Corpus case: four bits of a 4800 bit reg
//
// Differs from t33_v_wsl_lo by the assignment only.

`timescale 1ns / 1ps

module tb;
    reg [4799:0] s = 4800'h0;

    initial begin
        #50 s[3:0] = {1{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
