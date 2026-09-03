// SPDX-License-Identifier: Apache-2.0
// Corpus case: a typedef of an unpacked array, SystemVerilog.

`timescale 1ns / 1ps

module tb;
    typedef logic [3:0] arr_t [0:1];
    arr_t m;

    initial begin
        m[0] = 4'h0;
        m[1] = 4'h0;
        #50 m[1] = 4'ha;
        #50 $finish;
    end
endmodule
