// SPDX-License-Identifier: Apache-2.0
// Corpus case: a two dimensional unpacked array, logic [3:0] m [0:1][0:2].

`timescale 1ns / 1ps

module tb;
    logic [3:0] m [0:1][0:2];

    initial begin
        m[0][0] = 4'h0; m[0][1] = 4'h0; m[0][2] = 4'h0; m[1][0] = 4'h0; m[1][1] = 4'h0; m[1][2] = 4'h0;
        #50 m[1][2] = 4'ha;
        #50 $finish;
    end
endmodule
