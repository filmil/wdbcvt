// SPDX-License-Identifier: Apache-2.0
// Corpus case: a memory of 40 words of 32 bits, 320 record bytes,
// written element by element at time zero.

`timescale 1ns / 1ps

module tb;
    reg [31:0] m [0:39];

    initial begin
        m[0] = 32'h0;
        m[1] = 32'h0;
        m[2] = 32'h0;
        m[3] = 32'h0;
        m[4] = 32'h0;
        m[5] = 32'h0;
        m[6] = 32'h0;
        m[7] = 32'h0;
        m[8] = 32'h0;
        m[9] = 32'h0;
        m[10] = 32'h0;
        m[11] = 32'h0;
        m[12] = 32'h0;
        m[13] = 32'h0;
        m[14] = 32'h0;
        m[15] = 32'h0;
        m[16] = 32'h0;
        m[17] = 32'h0;
        m[18] = 32'h0;
        m[19] = 32'h0;
        m[20] = 32'h0;
        m[21] = 32'h0;
        m[22] = 32'h0;
        m[23] = 32'h0;
        m[24] = 32'h0;
        m[25] = 32'h0;
        m[26] = 32'h0;
        m[27] = 32'h0;
        m[28] = 32'h0;
        m[29] = 32'h0;
        m[30] = 32'h0;
        m[31] = 32'h0;
        m[32] = 32'h0;
        m[33] = 32'h0;
        m[34] = 32'h0;
        m[35] = 32'h0;
        m[36] = 32'h0;
        m[37] = 32'h0;
        m[38] = 32'h0;
        m[39] = 32'h0;
        #50 m[1] = 32'ha5c3f00f;
        #50 $finish;
    end
endmodule
