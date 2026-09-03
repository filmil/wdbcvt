// SPDX-License-Identifier: Apache-2.0
// Corpus case: the high half of a 4800 bit reg, 75 pairs from pair 75
//
// Differs from t12_v_vec4800x by the assignment only.

`timescale 1ns / 1ps

module tb;
    reg [4799:0] s = 4800'h0;

    initial begin
        #50 s[4799:2400] = {75{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
